package v1

import (
	"context"
	"net/http"

	"tomerab.com/cam-hub/internal/application"
)

func appMiddleware(app *application.Application) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ctxAppKey{}, app)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
