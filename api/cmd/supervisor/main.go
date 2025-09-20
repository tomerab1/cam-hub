package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/joho/godotenv"
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
		log.Println(err.Error())
		os.Exit(1)
	}
	defer bus.Close()

	supervisor := visor.NewSupervisor(10)
	onShutdown := func() {
		supervisor.Shutdown()
	}
	ctx, cancel := utils.GracefullShutdown(context.Background(), onShutdown, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	bus.Consume(ctx, "test", func(ctx context.Context, m events.Message) events.AckAction {
		fmt.Println(string(m.Body))
		return events.Ack
	})

	supervisor.Run()
}
