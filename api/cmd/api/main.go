package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"tomerab.com/cam-hub/internal/httpserver"
	"tomerab.com/cam-hub/internal/onvif"
)

var gClient *onvif.OnvifClient
var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelDebug,
}))

func GetDevDatetime(w http.ResponseWriter, r *http.Request) {
	datetime := gClient.GetDatetime()

	raw, err := json.Marshal(datetime)

	if err != nil {
		logger.Error(err.Error())
		return
	}

	w.Write(raw)
}

func GetDiscoveredDevices(w http.ResponseWriter, r *http.Request) {
	discovered := onvif.DiscoverNewCameras(logger)

	raw, err := json.Marshal(discovered)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	w.Write(raw)
}

type application struct {
}

func main() {
	logger.Debug(fmt.Sprintf("Serving on :%d", 5555))
	log.Fatal(http.ListenAndServe(":5555", httpserver.NewRouter()))
}
