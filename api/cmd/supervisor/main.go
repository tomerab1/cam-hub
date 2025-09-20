package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"syscall"

	"github.com/joho/godotenv"
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

	bus, err := rabbitmq.NewBus(os.Getenv("RABBITMQ_ADDR"))
	if err != nil {
		panic(err.Error())
	}

	supervisor := visor.NewSupervisor(10)
	onShutdown := func() {
		supervisor.Shutdown()
		_ = bus.Close()
	}
	ctx, cancel := utils.GracefullShutdown(context.Background(), onShutdown, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	bus.DeclareQueue("supervisor.pair", true, nil)
	bus.Consume(ctx, "supervisor.pair", "supervisor", func(ctx context.Context, m events.Message) events.AckAction {
		var ev v1.CameraPairedEvent
		json.Unmarshal(m.Body, &ev)

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

	supervisor.Run()
}
