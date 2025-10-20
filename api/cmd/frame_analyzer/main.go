package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"gopkg.in/lumberjack.v3"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/events"
	"tomerab.com/cam-hub/internal/events/rabbitmq"
	frameanalyzer "tomerab.com/cam-hub/internal/frame_analyzer"
	objectstorage "tomerab.com/cam-hub/internal/object_storage"
	"tomerab.com/cam-hub/internal/repos"
	"tomerab.com/cam-hub/internal/utils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err.Error())
	}

	fileHandler, err := lumberjack.New(
		lumberjack.WithFileName(os.Getenv("LOGGER_PATH")+"/frame_analyzer.log"),
		lumberjack.WithMaxBytes(25*lumberjack.MB),
		lumberjack.WithMaxDays(14),
		lumberjack.WithCompress(),
	)
	if err != nil {
		panic(err.Error())
	}
	logger := slog.New(slog.NewJSONHandler(fileHandler, &slog.HandlerOptions{Level: slog.LevelDebug}))

	bus, err := rabbitmq.NewBus(os.Getenv("RABBITMQ_ADDR"))
	if err != nil {
		panic(err.Error())
	}

	if err := bus.DeclareQueue("motion.analyze", true, nil); err != nil {
		panic(err.Error())
	}
	if err := bus.DeclareQueue("motion.detections", true, nil); err != nil {
		panic(err.Error())
	}

	ctx, cancel := utils.GracefullShutdown(context.Background(), func() {}, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	minioClient, err := objectstorage.NewMinIOStore(ctx, logger, false)
	if err != nil {
		logger.Error("failed to create MinIO client", "err", err.Error())
		panic(err.Error())
	}

	dbpool, err := pgxpool.New(ctx, os.Getenv("POSTGRES_DSN"))
	if err != nil {
		panic(err.Error())
	}
	defer dbpool.Close()

	recordingsRepo := repos.NewPgxRecordingsRepo(dbpool)
	camerasRepo := repos.NewPgxCameraRepo(dbpool)

	analyzer := frameanalyzer.New(ctx, logger, bus, minioClient, recordingsRepo, camerasRepo)

	bus.Consume(ctx, "motion.analyze", "", func(ctx context.Context, m events.Message) events.AckAction {
		var msg v1.AnalyzeImgsEvent
		if err := json.Unmarshal(m.Body, &msg); err != nil {
			logger.Error("consume: failed to parse json", "err", err.Error())
			return events.NackDiscard
		}

		analyzer.NotifyCtrl(msg)
		return events.Ack
	})

	analyzer.Run(ctx)
}
