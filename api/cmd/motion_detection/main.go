package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"gocv.io/x/gocv"
	"gopkg.in/lumberjack.v3"
	"tomerab.com/cam-hub/internal/motion"
	"tomerab.com/cam-hub/internal/utils"
)

var (
	lastMotionEvent time.Time
	motionCoolDown  = 10 * time.Second
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(fmt.Sprintf("failed to connect to load .env: %s", err.Error()))
	}

	addr := flag.String("addr", "", "rtsp url")
	flag.Parse()
	if *addr == "" {
		panic("missing -addr")
	}

	parts := strings.Split(*addr, "/")
	cameraUUID := parts[len(parts)-1]

	fileHandler, err := lumberjack.New(
		lumberjack.WithFileName(os.Getenv("LOGGER_PATH")+fmt.Sprintf("/%s/", cameraUUID)+"motion.log"),
		lumberjack.WithMaxBytes(25*lumberjack.MB),
		lumberjack.WithMaxDays(14),
		lumberjack.WithCompress(),
	)
	if err != nil {
		panic(err.Error())
	}
	logger := slog.New(slog.NewJSONHandler(fileHandler, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	cap, err := gocv.VideoCaptureFile(*addr)
	if err != nil {
		panic(err)
	}
	defer cap.Close()

	cap.Set(gocv.VideoCaptureBufferSize, 1)

	frame := gocv.NewMat()
	defer frame.Close()

	det := motion.NewMotionDetector(50, 15000, 150)
	defer det.Close()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	ctx, cancel := utils.GracefullShutdown(context.Background(), func() {
		logger.Debug("caught signal, terminating")
	}, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	runner := motion.NewRunner(ctx, 8)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ok := cap.Read(&frame); !ok || frame.Empty() {
				continue
			}

			isMotion, score := det.Detect(&frame)
			if isMotion && time.Since(lastMotionEvent) > motionCoolDown {
				lastMotionEvent = time.Now()
				go runner.PostJob(motion.MotionCtx{
					UUID:      cameraUUID,
					Score:     score,
					TimePoint: lastMotionEvent,
				})
			}
		}
	}
}
