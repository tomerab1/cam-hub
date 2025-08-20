package v1

import (
	"context"
	"log/slog"
)

type ctxLoggerKey struct{}

func LoggerFromCtx(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxLoggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
