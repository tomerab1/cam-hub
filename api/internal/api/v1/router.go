package v1

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func LoadRoutes() *chi.Mux {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from chi"))
	})

	r.Route("/cameras", func(r chi.Router) {
		r.With(loggerMiddleware(logger)).Get("/", GetDiscoveredDevices)
	})

	return r
}
