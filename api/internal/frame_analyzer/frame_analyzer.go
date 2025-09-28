package frameanalyzer

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	google_protobuf "github.com/golang/protobuf/ptypes/wrappers"
	"gocv.io/x/gocv"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"tomerab.com/cam-hub/internal/events"
	"tomerab.com/cam-hub/internal/frame_analyzer/tensorflow/core/framework"
	pb "tomerab.com/cam-hub/internal/frame_analyzer/tensorflow_serving/apis"
	objectstorage "tomerab.com/cam-hub/internal/object_storage"
)

const (
	maxConcurrentAnalysis = 16
	tensorW               = 544 // tensor width
	tensorH               = 320 // tensor heigt
	tensorC               = 3   // tensor channels (3 channels R-G-B)
	tensorN               = 4   // tensor batch size
	frameSz               = tensorW * tensorH * tensorC
)

var (
	bucketName = os.Getenv("MINIO_BUCKET_NAME")
)

type FrameAnalyzer struct {
	logger        *slog.Logger
	bus           events.BusIface
	minioClient   *objectstorage.MinIOStore
	ctx           context.Context
	imgAnalysisCh chan AnalyzeImgsEvent
}

func New(ctx context.Context, logger *slog.Logger, bus events.BusIface, minioClient *objectstorage.MinIOStore) *FrameAnalyzer {
	return &FrameAnalyzer{
		logger:        logger,
		bus:           bus,
		minioClient:   minioClient,
		ctx:           ctx,
		imgAnalysisCh: make(chan AnalyzeImgsEvent, maxConcurrentAnalysis),
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

func (analyzer *FrameAnalyzer) NotifyCtrl(ev AnalyzeImgsEvent) {
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
			obj, err := analyzer.minioClient.GetObject("recordings", p)
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

func (analyzer *FrameAnalyzer) onAnalyze(ev *AnalyzeImgsEvent) error {
	tensor, err := analyzer.buildTensor(ev.Paths)
	if err != nil {
		analyzer.logger.Error("onAnalyzer: failed to create tensor", "err", err.Error())
		return err
	}

	const MODEL_NAME = "person-detection-retail-0013"
	const INPUT_NAME = "data"
	const OUTPUT_NAME = "detection_out"

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
			INPUT_NAME: &framework.TensorProto{
				Dtype: framework.DataType_DT_FLOAT,
				TensorShape: &framework.TensorShapeProto{
					Dim: []*framework.TensorShapeProto_Dim{
						&framework.TensorShapeProto_Dim{
							Size: tensorN,
						},
						&framework.TensorShapeProto_Dim{
							Size: tensorC,
						},
						&framework.TensorShapeProto_Dim{
							Size: tensorH,
						},
						&framework.TensorShapeProto_Dim{
							Size: tensorW,
						},
					},
				},
				TensorContent: tensor,
			},
		},
	}

	conn, err := grpc.NewClient("localhost:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewPredictionServiceClient(conn)
	predictResp, err := client.Predict(analyzer.ctx, predicRequest)
	if err != nil {
		return fmt.Errorf("gRPC request failed: %w", err)
	}

	respProto, ok := predictResp.Outputs[OUTPUT_NAME]
	if !ok {
		return fmt.Errorf("expected output: %s does not exist in the response", OUTPUT_NAME)
	}
	responseContent := respProto.GetTensorContent()
	outputShape := respProto.GetTensorShape()
	dims := outputShape.GetDim()

	maxDetections := int(dims[2].GetSize())
	detectionFeatures := int(dims[3].GetSize())

	outMat, err := gocv.NewMatFromBytes(maxDetections, detectionFeatures, gocv.MatTypeCV32FC1, responseContent)
	if err != nil {
		return err
	}
	defer outMat.Close()

	detectionCount := 0
	for row := 0; row < outMat.Rows(); row++ {
		confidence := outMat.GetFloatAt(row, 2)
		if confidence > 0.5 {
			detectionCount++
		}
	}

	analyzer.logger.Info("detections found", "count", detectionCount)

	return nil
}
