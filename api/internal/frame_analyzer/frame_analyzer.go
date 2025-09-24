package frameanalyzer

import (
	"context"
	"log/slog"

	"gopkg.in/lumberjack.v3"
	"tomerab.com/cam-hub/internal/events"
)

type FrameAnalyzer struct {
	logger *slog.Logger
	bus    events.BusIface
}

func New(loggerSink lumberjack.Writer, bus events.BusIface) *FrameAnalyzer {
	logger := slog.New(slog.NewJSONHandler(loggerSink, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return &FrameAnalyzer{
		logger: logger,
		bus:    bus,
	}
}

func (analyzer *FrameAnalyzer) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		}
	}
}
