package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"syscall"

	"github.com/joho/godotenv"
	"gopkg.in/lumberjack.v3"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/events"
	"tomerab.com/cam-hub/internal/events/rabbitmq"
	visor "tomerab.com/cam-hub/internal/supervisor"
	"tomerab.com/cam-hub/internal/utils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(fmt.Sprintf("failed to connect to load .env: %s", err.Error()))
	}

	var (
		PairKey   = os.Getenv("RABBITMQ_PAIR_KEY")
		UnpairKey = os.Getenv("RABBITMQ_UNPAIR_KEY")
	)

	fileHandler, err := lumberjack.New(
		lumberjack.WithFileName(os.Getenv("LOGGER_PATH")+"/supervisor.log"),
		lumberjack.WithMaxBytes(25*lumberjack.MB),
		lumberjack.WithMaxDays(14),
		lumberjack.WithCompress(),
	)
	if err != nil {
		panic(err.Error())
	}

	bus, err := rabbitmq.NewBus(os.Getenv("RABBITMQ_ADDR"))
	if err != nil {
		panic(err.Error())
	}

	logger := slog.New(slog.NewJSONHandler(fileHandler, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	supervisor := visor.NewSupervisor(10, logger)
	onShutdown := func() {
		supervisor.Shutdown()
		_ = bus.Close()
	}
	ctx, cancel := utils.GracefullShutdown(context.Background(), onShutdown, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	bus.DeclareQueue(PairKey, true, nil)
	bus.DeclareQueue(UnpairKey, true, nil)

	bus.Consume(ctx, PairKey, "supervisor", func(ctx context.Context, m events.Message) events.AckAction {
		var ev v1.CameraPairedEvent
		if err := json.Unmarshal(m.Body, &ev); err != nil {
			return events.NackRequeue
		}

		rev, err := supervisor.GetCameraRevision(ev.UUID)
		if err != nil {
			supervisor.NotifyCtrl(visor.CtrlEvent{
				Kind:    visor.CtrlRegister,
				CamUUID: ev.UUID,
				Args: []string{
					"run", "./cmd/motion_detection",
					"-addr", ev.StreamUrl,
				},
			})
		} else {
			if rev < ev.Revision {
				supervisor.NotifyCtrl(visor.CtrlEvent{
					Kind:    visor.CtrlUnregister,
					CamUUID: ev.UUID,
				})
			}

			supervisor.NotifyCtrl(visor.CtrlEvent{
				Kind:    visor.CtrlRegister,
				CamUUID: ev.UUID,
				Args:    []string{"-addr", ev.StreamUrl},
			})
		}

		return events.Ack
	})

	bus.Consume(ctx, UnpairKey, "supervisor", func(ctx context.Context, m events.Message) events.AckAction {
		var ev v1.CameraUnpairedEvent
		if err := json.Unmarshal(m.Body, &ev); err != nil {
			return events.NackDiscard
		}

		supervisor.NotifyCtrl(visor.CtrlEvent{
			Kind:    visor.CtrlUnregister,
			CamUUID: ev.UUID,
		})

		return events.Ack
	})

	supervisor.Run(ctx)
}
