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
	"gopkg.in/lumberjack.v3"
	"tomerab.com/cam-hub/internal/application"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/events"
	inmemory "tomerab.com/cam-hub/internal/events/in_memory"
	"tomerab.com/cam-hub/internal/events/rabbitmq"
	"tomerab.com/cam-hub/internal/httpserver"
	"tomerab.com/cam-hub/internal/mtxapi"
	"tomerab.com/cam-hub/internal/repos"
	"tomerab.com/cam-hub/internal/services"
	"tomerab.com/cam-hub/internal/utils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err.Error())
	}

	fileHandler, err := lumberjack.New(
		lumberjack.WithFileName(os.Getenv("LOGGER_PATH")+"/api.log"),
		lumberjack.WithMaxBytes(25*lumberjack.MB),
		lumberjack.WithMaxDays(14),
		lumberjack.WithCompress(),
	)
	if err != nil {
		panic(err.Error())
	}

	base := slog.NewJSONHandler(fileHandler, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	appLogger := slog.New(base)
	redisRepoLogger := slog.New(base).With("repository", "redis")
	discoveryServiceLogger := slog.New(base).With("service", "discovery")
	cameraServiceLogger := slog.New(base).With("service", "camera")
	ptzServiceLogger := slog.New(base).With("service", "ptz")
	mtxServiceLogger := slog.New(base).With("service", "mtx")

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
		panic(err.Error())
	}

	sched, err := gocron.NewScheduler()
	if err != nil {
		panic(err.Error())
	}
	defer sched.Shutdown()

	camRepo := repos.NewPgxCameraRepo(dbpool)
	ptzRepo := repos.NewPgxPtzTokenRepo(dbpool)

	sseChan := make(chan v1.DiscoveryEvent, 24)
	dscSvc := &services.DiscoveryService{
		Rdb: &repos.RedisRepo{
			Rdb:    rdb,
			Logger: redisRepoLogger,
		},
		CamerasRepo: camRepo,
		Sched:       sched,
		Logger:      discoveryServiceLogger,
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

	if err := bus.DeclareQueue("motion.detections", true, nil); err != nil {
		panic(err.Error())
	}

	app := &application.Application{
		Logger:           appLogger,
		LogSink:          fileHandler,
		DB:               dbpool,
		DiscoveryService: dscSvc,
		SseChan:          sseChan,
		HttpClient:       &httpClient,
		CameraService: &services.CameraService{
			CamRepo:      camRepo,
			CamCredsRepo: credsRepo,
			Logger:       cameraServiceLogger,
		},
		PtzService: &services.PtzService{
			CamRepo:      camRepo,
			PtzTokenRepo: ptzRepo,
			CamCredsRepo: credsRepo,
			Rdb:          dscSvc.Rdb,
			Logger:       ptzServiceLogger,
		},
		MtxClient: &mtxapi.MtxClient{
			Logger:       mtxServiceLogger,
			CamRepo:      camRepo,
			CamCredsRepo: credsRepo,
			HttpClient:   &httpClient,
		},
		Bus:    bus,
		PubSub: inmemory.NewInMemoryPubSub(),
	}

	app.Bus.Consume(rootCtx, "motion.detections", "", func(ctx context.Context, m events.Message) events.AckAction {
		uuid, ok := m.Headers["uuid"].(string)
		if !ok {
			app.Logger.Error("missing uuid in message headers")
			return events.NackDiscard
		}

		app.PubSub.Broadcast(uuid, m.Body)
		return events.Ack
	})

	srv := http.Server{
		Addr:         os.Getenv("SERVER_ADDR"),
		Handler:      httpserver.NewRouter(app),
		WriteTimeout: 0,
		IdleTimeout:  0,
	}
	shutdownErrChan := make(chan error)
	onShutdown := func() {
		ctx, cancel := context.WithTimeout(rootCtx, 10*time.Second)
		defer cancel()

		shutdownErrChan <- srv.Shutdown(ctx)
	}
	_, cancel := utils.GracefullShutdown(rootCtx, onShutdown, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	appLogger.Info(fmt.Sprintf("Server is listening on %s", srv.Addr))
	err = srv.ListenAndServe()

	if !errors.Is(err, http.ErrServerClosed) {
		panic(err.Error())
	}

	err = <-shutdownErrChan
	if err != nil {
		panic(err.Error())
	}

	appLogger.Info("Shutting down the server...")
	os.Exit(0)
}
