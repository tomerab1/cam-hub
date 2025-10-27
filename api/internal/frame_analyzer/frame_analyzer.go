package frameanalyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"log/slog"
	"os"
	"path"
	"strconv"
	"strings"

	google_protobuf "github.com/golang/protobuf/ptypes/wrappers"
	"gocv.io/x/gocv"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/events"
	"tomerab.com/cam-hub/internal/frame_analyzer/tensorflow/core/framework"
	pb "tomerab.com/cam-hub/internal/frame_analyzer/tensorflow_serving/apis"
	objectstorage "tomerab.com/cam-hub/internal/object_storage"
	"tomerab.com/cam-hub/internal/repos"
	"tomerab.com/cam-hub/internal/services"
)

const (
	maxConcurrentAnalysis = 16
	tensorW               = 544 // tensor width
	tensorH               = 320 // tensor heigt
	tensorC               = 3   // tensor channels (3 channels R-G-B)
	tensorN               = 4   // tensor batch size
	frameSz               = tensorW * tensorH * tensorC
)

type FrameAnalyzer struct {
	logger           *slog.Logger
	bus              events.BusIface
	minioClient      *objectstorage.MinIOStore
	recordingService *services.RecordingsService
	ctx              context.Context
	imgAnalysisCh    chan v1.AnalyzeImgsEvent
}

type tensorData struct {
	numClasses int
	numRows    int
	imageIndex int
	maxConf    float32
	maxLoc     image.Point
	evidence   v1.Evidence
}

func New(ctx context.Context,
	logger *slog.Logger,
	bus events.BusIface,
	minioClient *objectstorage.MinIOStore,
	recordingsRepo repos.RecordingsRepoIface,
	camerasRepo repos.CameraRepoIface,
) *FrameAnalyzer {
	recordingServiceLogger := slog.New(logger.Handler()).With("service", "recordings")

	return &FrameAnalyzer{
		logger:           logger,
		bus:              bus,
		minioClient:      minioClient,
		recordingService: services.NewRecordingsService(recordingServiceLogger, recordingsRepo, camerasRepo),
		ctx:              ctx,
		imgAnalysisCh:    make(chan v1.AnalyzeImgsEvent, maxConcurrentAnalysis),
	}
}

func (analyzer *FrameAnalyzer) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-analyzer.imgAnalysisCh:
			if err := analyzer.onAnalyze(&ev); err != nil {
				analyzer.logger.Error("onAnalyze failed", "err", err.Error())
			}
		}
	}
}

func (analyzer *FrameAnalyzer) NotifyCtrl(ev v1.AnalyzeImgsEvent) {
	analyzer.imgAnalysisCh <- ev
}

func (analyzer *FrameAnalyzer) buildTensor(paths []string) ([]byte, error) {
	batch := make([]byte, tensorW*tensorH*tensorC*tensorN*4) // *4 for float32
	var eg errgroup.Group

	for i, path := range paths {
		p := path
		start := i * frameSz * 4
		end := (i + 1) * frameSz * 4

		eg.Go(func() error {
			obj, err := analyzer.minioClient.GetObject(os.Getenv("MINIO_BUCKET_NAME"), p)
			if err != nil {
				return err
			}
			defer obj.Close()

			buf, err := io.ReadAll(obj)
			if err != nil {
				return err
			}

			mat, err := gocv.IMDecode(buf, gocv.IMReadColor)
			if err != nil {
				return fmt.Errorf("imdecode failed: %w", err)
			}
			defer mat.Close()
			if mat.Empty() {
				return fmt.Errorf("unexpected dims for %s: got %dx%dx%d (HWC)", p, mat.Rows(), mat.Cols(), mat.Channels())
			}

			bgrFloat := gocv.NewMat()
			defer bgrFloat.Close()
			mat.ConvertTo(&bgrFloat, gocv.MatTypeCV32F)

			reshaped := bgrFloat.ReshapeWithSize(1, []int{1, tensorH, tensorW, tensorC})
			defer reshaped.Close()

			transposed := gocv.NewMat()
			defer transposed.Close()
			if err := gocv.TransposeND(reshaped, []int{0, 3, 1, 2}, &transposed); err != nil {
				return fmt.Errorf("transposeND failed: %w", err)
			}

			bytesData := transposed.ToBytes()
			copy(batch[start:end], bytesData)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return batch, nil
}

func (analyzer *FrameAnalyzer) minioMoveObjects(whereToStore string, ev *v1.AnalyzeImgsEvent) error {
	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	stagingKey := os.Getenv("MINIO_STAGING_KEY")

	dirPrefix := path.Join(stagingKey, ev.UUID, ev.Tp)
	allObjs := append([]string{ev.VidPath}, ev.FramePaths...)
	for _, srcKey := range allObjs {
		objName := strings.TrimPrefix(srcKey, stagingKey+"/")

		if _, err := analyzer.minioClient.CopyObjectWithinBucket(
			bucketName,
			stagingKey,
			whereToStore,
			objName,
		); err != nil {
			return err
		}
	}

	// bulk remove all the objects under 'dirPrefix' (all the objects we remove are under the
	// same 'dir' object)
	analyzer.logger.Debug("deleting objects from", "bucket_name", bucketName, "prefix", dirPrefix)
	if err := analyzer.minioClient.RemoveObjects(bucketName, dirPrefix); err != nil {
		return err
	}

	return nil
}

func (analyzer *FrameAnalyzer) onAnalyze(ev *v1.AnalyzeImgsEvent) error {
	const OUTPUT_NAME = "detection_out"
	var stagingKey = os.Getenv("MINIO_STAGING_KEY")

	tensor, err := analyzer.buildTensor(ev.FramePaths)
	if err != nil {
		analyzer.logger.Error("onAnalyzer: failed to create tensor", "err", err.Error())
		return err
	}

	predictResp, err := analyzer.predict(tensor)
	if err != nil {
		return err
	}

	respProto, ok := predictResp.Outputs[OUTPUT_NAME]
	if !ok {
		return fmt.Errorf("expected output: %s does not exist in the response", OUTPUT_NAME)
	}

	tensorData, err := analyzer.extractDataFromTensor(respProto)
	if err != nil {
		return err
	}

	state := services.StateDiscarded
	whereToStore := os.Getenv("MINIO_FALSE_POSITIVES_KEY")
	retentionDaysStr := os.Getenv("MINIO_FALSE_POSITIVES_DAYS")
	if tensorData.maxConf >= 0.5 {
		state = services.StatePromoted
		whereToStore = os.Getenv("MINIO_DETECTIONS_KEY")
		retentionDaysStr = os.Getenv("MINIO_DETECTIONS_DAYS")
	}

	retentionDays, err := strconv.Atoi(retentionDaysStr)
	if err != nil {
		return err
	}

	if err := analyzer.minioMoveObjects(whereToStore, ev); err != nil {
		return err
	}

	req := v1.AddRecordingReq{
		BucketName:         os.Getenv("MINIO_BUCKET_NAME"),
		VidBucketKey:       path.Join(whereToStore, strings.TrimPrefix(ev.VidPath, stagingKey)),
		BestFrameBucketKey: path.Join(whereToStore, strings.TrimPrefix(ev.FramePaths[tensorData.imageIndex], stagingKey)),
		Evidence:           tensorData.evidence,
		Score:              tensorData.maxConf,
		RetentionDays:      retentionDays,
	}

	model, err := analyzer.recordingService.Upsert(analyzer.ctx, ev.UUID, state, req)
	if err != nil {
		return err
	}

	analyzer.logger.Debug("upserted new recording", "where_to_store", whereToStore, "state", state, "max_conf", tensorData.maxConf)

	if model.Score >= 0.5 {
		modelBytes, err := json.Marshal(model)
		if err != nil {
			return err
		}
		if err := analyzer.bus.Publish(analyzer.ctx, "", "motion.detections", modelBytes, map[string]any{"uuid": model.CamUUID}); err != nil {
			return err
		}
	}

	return nil
}

func (analyzer *FrameAnalyzer) extractDataFromTensor(respProto *framework.TensorProto) (*tensorData, error) {
	// shape should be [4, 1, 200, 7]
	// 4 - batch size, 1 - class (person or not), 200 - max number of detections per image, 7 - detection info
	responseContent := respProto.GetTensorContent()
	dim := respProto.GetTensorShape().GetDim()
	numClasses := int(dim[3].GetSize())
	numRows := int(dim[2].GetSize())

	outMat, err := gocv.NewMatFromBytes(numRows, numClasses, gocv.MatTypeCV32FC1, responseContent)
	if err != nil {
		return nil, err
	}
	defer outMat.Close()

	confCol := outMat.Col(2)
	defer confCol.Close()

	_, maxConf, _, maxLoc := gocv.MinMaxLoc(confCol)
	imageIndex := int(maxLoc.Y / numRows)
	detectionRow := outMat.Row(maxLoc.Y)
	defer detectionRow.Close()

	evidence := v1.Evidence{
		Conf: maxConf,
		Xmin: detectionRow.GetFloatAt(0, 3),
		Ymin: detectionRow.GetFloatAt(0, 4),
		Xmax: detectionRow.GetFloatAt(0, 5),
		Ymax: detectionRow.GetFloatAt(0, 6),
	}

	return &tensorData{
		numClasses: numClasses,
		numRows:    numRows,
		imageIndex: imageIndex,
		evidence:   evidence,
		maxConf:    maxConf,
		maxLoc:     maxLoc,
	}, nil
}

func (analyzer *FrameAnalyzer) predict(tensor []byte) (*pb.PredictResponse, error) {
	const MODEL_NAME = "person-detection-retail-0013"
	const INPUT_NAME = "data"

	predicRequest := &pb.PredictRequest{
		ModelSpec: &pb.ModelSpec{
			Name:          MODEL_NAME,
			SignatureName: "serving_default",
			VersionChoice: &pb.ModelSpec_Version{
				Version: &google_protobuf.Int64Value{
					Value: 1,
				},
			},
		},
		Inputs: map[string]*framework.TensorProto{
			INPUT_NAME: {
				Dtype: framework.DataType_DT_FLOAT,
				TensorShape: &framework.TensorShapeProto{
					Dim: []*framework.TensorShapeProto_Dim{
						{
							Size: tensorN,
						},
						{
							Size: tensorC,
						},
						{
							Size: tensorH,
						},
						{
							Size: tensorW,
						},
					},
				},
				TensorContent: tensor,
			},
		},
	}

	conn, err := grpc.NewClient(
		os.Getenv("OVMS_GRPC_ADDR"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewPredictionServiceClient(conn)
	predictResp, err := client.Predict(analyzer.ctx, predicRequest)
	if err != nil {
		return nil, fmt.Errorf("gRPC request failed: %w", err)
	}

	return predictResp, nil
}
