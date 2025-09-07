package v1

import (
	"github.com/go-chi/chi/v5"
	"tomerab.com/cam-hub/internal/application"
)

func LoadRoutes(app *application.Application) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/cameras", func(r chi.Router) {
		r.Get("/discovery", getDiscoveredDevices(app))
		r.Get("/", getCameras(app))
		r.Post("/{uuid}/pair", pairCamera(app))
		r.Patch("/{uuid}/pair", unpairCamera(app))
	})

	r.Get("/events/discovery", discoverySSE(app))

	return r
}
