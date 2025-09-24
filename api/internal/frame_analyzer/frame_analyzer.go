package frameanalyzer

import (
	"context"
	"fmt"
	"log/slog"

	"tomerab.com/cam-hub/internal/events"
)

const (
	maxConcurrentAnalysis = 16
)

type FrameAnalyzer struct {
	logger        *slog.Logger
	bus           events.BusIface
	imgAnalysisCh chan AnalyzeImgsEvent
}

func New(logger *slog.Logger, bus events.BusIface) *FrameAnalyzer {
	return &FrameAnalyzer{
		logger:        logger,
		bus:           bus,
		imgAnalysisCh: make(chan AnalyzeImgsEvent, maxConcurrentAnalysis),
	}
}

func (analyzer *FrameAnalyzer) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-analyzer.imgAnalysisCh:
			analyzer.logger.Info(fmt.Sprintf("%+v\n", ev))
		}
	}
}

func (analyzer *FrameAnalyzer) NotifyCtrl(ev AnalyzeImgsEvent) {
	analyzer.imgAnalysisCh <- ev
}
