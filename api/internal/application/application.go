package application

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Config struct {
	Addr string
}

type Application struct {
	conf   *Config
	logger *slog.Logger
	router *chi.Mux
}

func NewApplication(logger *slog.Logger, router *chi.Mux, config *Config) *Application {
	return &Application{
		logger: logger,
		router: router,
		conf:   config,
	}
}

func (app *Application) Start() error {
	app.logger.Info("Starting server...", "Addr", app.conf.Addr)
	if err := http.ListenAndServe(app.conf.Addr, app.router); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
