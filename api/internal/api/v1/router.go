package v1

import (
	"github.com/go-chi/chi/v5"
	"tomerab.com/cam-hub/internal/application"
)

func LoadRoutes(app *application.Application) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/cameras", func(r chi.Router) {
		r.With(appMiddleware(app)).Get("/", getDiscoveredDevices)
		r.With(appMiddleware(app)).Post("/", pairCamera)
		r.With(appMiddleware(app)).Patch("/", unpairCamera)
	})

	r.With(appMiddleware(app)).Get("/events/discovery", discoverySSE(app))

	return r
}
