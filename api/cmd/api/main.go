package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"tomerab.com/cam-hub/internal/application"
	"tomerab.com/cam-hub/internal/httpserver"
	"tomerab.com/cam-hub/internal/repos"
	"tomerab.com/cam-hub/internal/services"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	err := godotenv.Load("/home/tomerab/VSCProjects/cam-hub/api/.env")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	dbpool, err := pgxpool.New(context.Background(), os.Getenv("POSTGRES_DSN"))
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer dbpool.Close()

	transport := http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
	httpClient := http.Client{
		Transport: &transport,
		Timeout:   5 * time.Second,
	}

	app := &application.Application{
		Logger:     logger,
		DB:         dbpool,
		HttpClient: &httpClient,
		CameraService: &services.CameraService{
			CamRepo:      repos.NewPgxCameraRepo(dbpool),
			CamCredsRepo: repos.NewPgxCameraCredsRepo(dbpool),
			Logger:       logger,
		},
	}

	srv := http.Server{
		Addr:    os.Getenv("SERVER_ADDR"),
		Handler: httpserver.NewRouter(app),
	}

	shutdownErrChan := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit
		logger.Info("Caught a signal", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		shutdownErrChan <- srv.Shutdown(ctx)
	}()

	logger.Info(fmt.Sprintf("Server is listening on %s", srv.Addr))
	err = srv.ListenAndServe()

	if !errors.Is(err, http.ErrServerClosed) {
		logger.Error("Error happend while returning for 'ListenAndServer()'", "err", err.Error())
		os.Exit(1)
	}

	err = <-shutdownErrChan
	if err != nil {
		logger.Error("Error happend while returning from 'Shutdown()'", "err", err.Error())
		os.Exit(1)
	}

	logger.Info("Shutting down the server...")
	os.Exit(0)
}
