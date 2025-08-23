package httpserver

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	v1 "tomerab.com/cam-hub/internal/api/v1"
	"tomerab.com/cam-hub/internal/application"
)

func NewRouter(app *application.Application) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(commonHeaders)

	r.Mount("/api/v1", v1.LoadRoutes(app))

	return r
}
