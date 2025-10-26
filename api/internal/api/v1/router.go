package v1

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"tomerab.com/cam-hub/internal/application"
)

func LoadRoutes(app *application.Application) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/cameras", func(r chi.Router) {
		rt := r.With(middleware.Timeout(60 * time.Second))

		rt.Get("/discovery", getDiscoveredDevices(app))
		rt.Get("/{uuid}/stream", getCameraStream(app))
		rt.Delete("/{uuid}/stream", deleteCameraStream(app))
		rt.Get("/", getCameras(app))
		rt.Post("/{uuid}/pair", pairCamera(app))
		rt.Delete("/{uuid}/pair", unpairCamera(app))
		rt.Post("/{uuid}/ptz/move", moveCamera(app))
	})

	r.Get("/events/discovery", discoverySSE(app))
	r.Get("/events/recordings/{uuid}", alertsSSE(app))

	return r
}
