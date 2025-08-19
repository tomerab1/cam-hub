package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"tomerab.com/cam-hub/internal/onvif"
)

var gClient *onvif.OnvifClient
var logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelDebug,
}))

func GetDevDatetime(w http.ResponseWriter, r *http.Request) {
	datetime := gClient.GetDatetime()
	resp := ApiResponse{
		Data: datetime,
	}

	raw, err := json.Marshal(resp)

	if err != nil {
		logger.Error(err.Error())
		return
	}

	w.Write(raw)
}

func GetDiscoveredDevices(w http.ResponseWriter, r *http.Request) {
	discovered := onvif.DiscoverNewCameras(logger)

	resp := ApiResponse{
		Data: discovered,
	}

	raw, err := json.Marshal(resp)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	w.Write(raw)
}

type application struct {
}

func main() {
	// params := onvif.OnvifClientParams{
	// 	Xaddr:  "10.0.0.4:8899",
	// 	Logger: logger,
	// }

	// client, err := onvif.NewOnvifClient(params)

	// if err != nil {
	// 	logger.Error(err.Error())
	// 	os.Exit(1)
	// }

	// gClient = client
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/{dev_id}/datetime", GetDevDatetime)
	mux.HandleFunc("/api/v1/enumerate-devices", GetDiscoveredDevices)

	logger.Debug(fmt.Sprintf("Serving on :%d", 5555))
	log.Fatal(http.ListenAndServe(":5555", mux))
}
