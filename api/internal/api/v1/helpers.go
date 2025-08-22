package v1

import (
	"context"
	"log/slog"
	"net/http"

	"tomerab.com/cam-hub/internal/application"
)

type ctxAppKey struct{}

func appFromCtx(ctx context.Context) *application.Application {
	if l, ok := ctx.Value(ctxAppKey{}).(*application.Application); ok {
		return l
	}
	return nil
}

func serverError(w http.ResponseWriter, r *http.Request, err error, logger *slog.Logger) {
	var (
		uri    = r.URL.RequestURI()
		method = r.Method
	)

	logger.Error(err.Error(), "method", method, "uri", uri)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
