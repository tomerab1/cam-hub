package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/onvif"
)

func getDiscoveredDevices(w http.ResponseWriter, r *http.Request) {
	logger := appFromCtx(r.Context()).Logger
	discovered := onvif.DiscoverNewCameras(logger)

	raw, err := json.Marshal(discovered)
	if err != nil {
		serverError(w, r, err, logger)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func pairCamera(w http.ResponseWriter, r *http.Request) {
	app := appFromCtx(r.Context())

	var req v1.PairDeviceReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	camera, err := app.CameraService.Pair(ctx, req)
	if err != nil {
		app.Logger.Error(err.Error())
		serverError(w, r, err, app.Logger)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(camera)
}
