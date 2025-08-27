package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"tomerab.com/cam-hub/internal/api"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/onvif"
	"tomerab.com/cam-hub/internal/onvif/discovery"
	"tomerab.com/cam-hub/internal/repos"
)

func filterUUIDS(ctx context.Context, camRepo repos.CameraRepoIface, matches []discovery.WsDiscoveryMatch) ([]discovery.WsDiscoveryMatch, error) {
	var uuids []string

	for _, match := range matches {
		uuids = append(uuids, match.UUID)
	}

	filters, err := camRepo.FindExistingPaired(ctx, uuids)

	fmt.Println(filters, uuids)

	if err != nil {
		return nil, err
	}

	predCount := func(elem bool) bool {
		return !elem
	}

	predFilter := func(idx int, elem discovery.WsDiscoveryMatch) bool {
		return !filters[idx]
	}

	filtered := api.FilterElems(matches, api.CountElems(filters, predCount), predFilter)

	return filtered, err
}

func getDiscoveredDevices(w http.ResponseWriter, r *http.Request) {
	app := appFromCtx(r.Context())
	discovered := onvif.DiscoverNewCameras(app.Logger)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	filteredMatches, err := filterUUIDS(ctx, app.CameraService.CamRepo, discovered.Matches)
	if err != nil {
		serverError(w, r, err, app.Logger)
		return
	}

	raw, err := json.Marshal(discovery.WsDiscoveryDto{Matches: filteredMatches})
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
