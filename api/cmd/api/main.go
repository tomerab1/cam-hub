package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"tomerab.com/cam-hub/internal/application"
	"tomerab.com/cam-hub/internal/httpserver"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	err := godotenv.Load()
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
	}

	srv := http.Server{
		Addr:    os.Getenv("SERVER_ADDR"),
		Handler: httpserver.NewRouter(app),
	}

	logger.Info(fmt.Sprintf("Server is listening on %s", srv.Addr))
	if err := srv.ListenAndServe(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
