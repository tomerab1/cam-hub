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

	"github.com/go-co-op/gocron/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"tomerab.com/cam-hub/internal/application"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/httpserver"
	"tomerab.com/cam-hub/internal/repos"
	"tomerab.com/cam-hub/internal/services"
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
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_CACHE"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	defer rdb.Close()

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Error("redis ping failed", "err", err)
		os.Exit(1)
	}

	sched, err := gocron.NewScheduler()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer sched.Shutdown()

	camRepo := repos.NewPgxCameraRepo(dbpool)

	sseChan := make(chan v1.DiscoveryEvent, 24)
	dscSvc := &services.DiscoveryService{
		Rdb: &repos.RedisRepo{
			Rdb:    rdb,
			Logger: logger,
		},
		CamerasRepo: camRepo,
		Sched:       sched,
		Logger:      logger,
		SseChan:     sseChan,
	}
	err = dscSvc.InitJobs(context.Background())
	if err != nil {
		logger.Error("Failed to init jobs", "err", err.Error())
		os.Exit(1)
	}
	dscSvc.Sched.Start()

	app := &application.Application{
		Logger:     logger,
		DB:         dbpool,
		HttpClient: &httpClient,
		CameraService: &services.CameraService{
			CamRepo:      camRepo,
			CamCredsRepo: repos.NewPgxCameraCredsRepo(dbpool),
			Logger:       logger,
		},
		DiscoveryService: dscSvc,
		SseChan:          sseChan,
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
