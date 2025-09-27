package frameanalyzer

import (
	"context"
	"log/slog"
	"os"

	"golang.org/x/sync/errgroup"
	"tomerab.com/cam-hub/internal/events"
	objectstorage "tomerab.com/cam-hub/internal/object_storage"
)

const (
	maxConcurrentAnalysis    = 16
	maxConcurrentObjRetrival = 4
	tensorW                  = 544 // tensor width
	tensorH                  = 320 // tensor heigt
	tensorC                  = 3   // tensor channels (3 channels R-G-B)
	tensorN                  = 4   // tensor batch size
	frameSz                  = tensorW * tensorH * tensorC
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
			analyzer.onAnalyze(&ev)
		}
	}
}

func (analyzer *FrameAnalyzer) NotifyCtrl(ev AnalyzeImgsEvent) {
	analyzer.imgAnalysisCh <- ev
}

func (analyzer *FrameAnalyzer) buildTensor(ev *AnalyzeImgsEvent) error {
	batch := make([]float32, tensorW*tensorH*tensorC*tensorN)

	return nil
}

func (analyzer *FrameAnalyzer) onAnalyze(ev *AnalyzeImgsEvent) error {
	eg, _ := errgroup.WithContext(analyzer.ctx)
	eg.SetLimit(maxConcurrentObjRetrival)

	for _, path := range ev.Paths {
		p := path
		eg.Go(func() error {
			obj, err := analyzer.minioClient.GetObject(bucketName, p)
			if err != nil {
				return err
			}
			defer obj.Close()

			return nil
		})
	}

	return eg.Wait()
}
