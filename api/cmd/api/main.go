package main

import (
	"log/slog"
	"os"

	"tomerab.com/cam-hub/internal/application"
	"tomerab.com/cam-hub/internal/httpserver"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	addr := ":5555"

	app := application.NewApplication(logger, httpserver.NewRouter(), &application.Config{
		Addr: addr,
	})

	if err := app.Start(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
