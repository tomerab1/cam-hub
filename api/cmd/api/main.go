package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"tomerab.com/cam-hub/internal/application"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/events/rabbitmq"
	"tomerab.com/cam-hub/internal/httpserver"
	"tomerab.com/cam-hub/internal/mtxapi"
	"tomerab.com/cam-hub/internal/repos"
	"tomerab.com/cam-hub/internal/services"
	"tomerab.com/cam-hub/internal/utils"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	err := godotenv.Load()
	if err != nil {
		panic(err.Error())
	}

	rootCtx := context.Background()
	dbpool, err := pgxpool.New(rootCtx, os.Getenv("POSTGRES_DSN"))
	if err != nil {
		panic(err.Error())
	}
	defer dbpool.Close()

	transport := http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   true,
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

	if err := rdb.Ping(rootCtx).Err(); err != nil {
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
	ptzRepo := repos.NewPgxPtzTokenRepo(dbpool)

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
	err = dscSvc.InitJobs(rootCtx)
	if err != nil {
		panic(err.Error())
	}
	dscSvc.Sched.Start()

	credsRepo := repos.NewPgxCameraCredsRepo(dbpool)
	bus, err := rabbitmq.NewBus(os.Getenv("RABBITMQ_ADDR"))
	if err != nil {
		panic(err.Error())
	}

	app := &application.Application{
		Logger:           logger,
		DB:               dbpool,
		DiscoveryService: dscSvc,
		SseChan:          sseChan,
		HttpClient:       &httpClient,
		CameraService: &services.CameraService{
			CamRepo:      camRepo,
			CamCredsRepo: credsRepo,
			Logger:       logger,
		},
		PtzService: &services.PtzService{
			CamRepo:      camRepo,
			PtzTokenRepo: ptzRepo,
			CamCredsRepo: credsRepo,
			Rdb:          dscSvc.Rdb,
			Logger:       logger,
		},
		MtxClient: &mtxapi.MtxClient{
			Logger:       logger,
			CamRepo:      camRepo,
			CamCredsRepo: credsRepo,
			HttpClient:   &httpClient,
		},
		Bus: bus,
	}

	srv := http.Server{
		Addr:    os.Getenv("SERVER_ADDR"),
		Handler: httpserver.NewRouter(app),
	}
	shutdownErrChan := make(chan error)
	onShutdown := func() {
		ctx, cancel := context.WithTimeout(rootCtx, 10*time.Second)
		defer cancel()

		shutdownErrChan <- srv.Shutdown(ctx)
	}
	_, cancel := utils.GracefullShutdown(rootCtx, onShutdown, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.Info(fmt.Sprintf("Server is listening on %s", srv.Addr))
	err = srv.ListenAndServe()

	if !errors.Is(err, http.ErrServerClosed) {
		panic(err.Error())
	}

	err = <-shutdownErrChan
	if err != nil {
		panic(err.Error())
	}

	logger.Info("Shutting down the server...")
	os.Exit(0)
}
