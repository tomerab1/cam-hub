package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"tomerab.com/cam-hub/internal/application"
)

func LoadRoutes(app *application.Application) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from chi"))
	})

	r.Route("/cameras", func(r chi.Router) {
		r.With(appMiddleware(app)).Get("/", getDiscoveredDevices)
		r.With(appMiddleware(app)).Get("/profiles", getProfiles)
	})

	return r
}
