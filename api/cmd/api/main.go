package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"tomerab.com/cam-hub/internal/application"
	"tomerab.com/cam-hub/internal/httpserver"
	"tomerab.com/cam-hub/internal/onvif"
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

	client, err := onvif.NewOnvifClient(onvif.OnvifClientParams{
		Xaddr: "10.0.0.5:8899",
	})

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &application.Application{
		Logger: logger,
		Client: client,
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
