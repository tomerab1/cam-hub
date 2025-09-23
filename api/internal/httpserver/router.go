package httpserver

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
	v1 "tomerab.com/cam-hub/internal/api/v1"
	"tomerab.com/cam-hub/internal/application"
)

func NewRouter(app *application.Application) *chi.Mux {
	r := chi.NewRouter()
	logger := slog.New(slog.NewJSONHandler(app.LogSink, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	r.Use(httplog.RequestLogger(logger, &httplog.Options{
		Level:  slog.LevelDebug,
		Schema: httplog.SchemaECS,
	}))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(commonHeaders)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Mount("/api/v1", v1.LoadRoutes(app))

	return r
}
