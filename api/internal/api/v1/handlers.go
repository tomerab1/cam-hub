package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/onvif"
	"tomerab.com/cam-hub/internal/onvif/discovery"
	"tomerab.com/cam-hub/internal/repos"
)

func filterUUIDS(ctx context.Context, camRepo *repos.PgxCameraRepo, outDiscovered *discovery.WsDiscoveryDto) error {
	var uuids []string
	var lastErr error = nil

	for _, match := range outDiscovered.Matches {
		uuids = append(uuids, match.UUID)
	}

	filters, err := camRepo.ExistsMany(ctx, uuids)

	if err != nil {
		lastErr = err
	}

	for i := range filters {
		if filters[i] {
			// The WsDiscoveryMatch has 'omitempty' specified so we empty the values to filter them in json ser.
			outDiscovered.Matches[i].UUID = ""
			outDiscovered.Matches[i].Xaddr = ""
		}
	}

	return lastErr
}

func getDiscoveredDevices(w http.ResponseWriter, r *http.Request) {
	app := appFromCtx(r.Context())
	discovered := onvif.DiscoverNewCameras(app.Logger)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err := filterUUIDS(ctx, app.CameraService.CamRepo, &discovered)
	if err != nil {
		serverError(w, r, err, app.Logger)
		return
	}

	raw, err := json.Marshal(discovered)
	if err != nil {
		serverError(w, r, err, app.Logger)
		return
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
		serverError(w, r, err, app.Logger)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(camera)
}

func unpairCamera(w http.ResponseWriter, r *http.Request) {
	app := appFromCtx(r.Context())
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req v1.UnpairDeviceReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := app.CameraService.Unpair(ctx, req); err != nil {
		serverError(w, r, err, app.Logger)
		return
	}

	w.WriteHeader(http.StatusOK)
}
