package v1

import (
	"context"
	"log/slog"
	"net/http"
)

func loggerMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ctxLoggerKey{}, logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
