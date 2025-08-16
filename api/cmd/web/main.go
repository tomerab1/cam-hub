package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"

	"tomerab.com/cam-hub/internal/onvif"
)

var gClient *onvif.OnvifClient

func GetDevDatetime(w http.ResponseWriter, r *http.Request) {
	datetime := gClient.GetDatetime()
	raw, err := json.Marshal(datetime)

	if err != nil {
		log.Println(err.Error())
		return
	}

	w.Write(raw)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	params := onvif.OnvifClientParams{
		Xaddr:  "10.0.0.19:8899",
		Logger: logger,
	}

	client, err := onvif.NewOnvifClient(params)

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	gClient = client
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/{dev_id}/datetime", GetDevDatetime)

	err = http.ListenAndServe(":5555", mux)
	log.Fatal(err.Error())
}
