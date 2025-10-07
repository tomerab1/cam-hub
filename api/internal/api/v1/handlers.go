package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"tomerab.com/cam-hub/internal/api"
	"tomerab.com/cam-hub/internal/application"
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

func getDiscoveredDevices(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		discovered := onvif.DiscoverNewCameras(ctx, app.Logger)

		filteredMatches, err := filterUUIDS(ctx, app.CameraService.CamRepo, discovered.Matches)
		if err != nil {
			serverError(w, r, err, app.Logger)
			return
		}

		app.WriteJSON(w, r, discovery.WsDiscoveryDto{Matches: filteredMatches}, http.StatusOK)
	}
}

func pairCamera(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req v1.PairDeviceReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		uuid := r.PathValue("uuid")
		camera, err := app.CameraService.Pair(ctx, uuid, req)
		if err != nil {
			serverError(w, r, err, app.Logger)
			return
		}

		app.WriteJSON(w, r, camera, http.StatusCreated)
	}
}

func unpairCamera(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		uuid := r.PathValue("uuid")
		if err := app.CameraService.Unpair(ctx, uuid); err != nil {
			serverError(w, r, err, app.Logger)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func discoverySSE(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		ctx := r.Context()
		for {
			select {
			case evt := <-app.SseChan:
				data, _ := json.Marshal(evt)
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			case <-ctx.Done():
				return
			}
		}
	}
}

func alertsSSE(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		ctx := r.Context()
		uuid := r.PathValue("uuid")
		if !app.CameraService.CameraExists(ctx, uuid) {
			http.Error(w, fmt.Sprintf("cameras (%s) does not exist", uuid), http.StatusBadRequest)
			return
		}

		subCh := app.PubSub.Subscribe(uuid)
		for {
			select {
			case msg := <-subCh:
				data, _ := json.Marshal(msg)
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			case <-ctx.Done():
				return
			}
		}
	}
}

func getCameras(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		queryParams := r.URL.Query()
		offset, err := strconv.Atoi(queryParams.Get("offset"))

		if err != nil {
			app.WriteJSON(w, r, api.ErrorEnvp{"error": err.Error()}, http.StatusBadRequest)
			return
		}

		limit, err := strconv.Atoi(queryParams.Get("limit"))
		if err != nil {
			app.WriteJSON(w, r, api.ErrorEnvp{"error": err.Error()}, http.StatusBadRequest)
			return
		}

		cams, err := app.CameraService.GetCameras(ctx, offset, limit)
		if err != nil {
			app.WriteJSON(w, r, api.ErrorEnvp{"error": err.Error()}, http.StatusBadRequest)
			return
		}

		app.WriteJSON(w, r, cams, http.StatusOK)
	}
}

func getCameraStream(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		uuid := r.PathValue("uuid")
		streamUrl, err := app.MtxClient.Publish(ctx, uuid)
		if err != nil {
			app.WriteJSON(w, r, api.ErrorEnvp{"error": err.Error()}, http.StatusInternalServerError)
			return
		}

		app.WriteJSON(w, r, v1.CameraStreamUrl{Url: streamUrl}, http.StatusOK)
	}
}

func deleteCameraStream(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		uuid := r.PathValue("uuid")
		err := app.MtxClient.Delete(ctx, uuid)
		if err != nil {
			app.WriteJSON(w, r, api.ErrorEnvp{"error": err.Error()}, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func moveCamera(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		defer r.Body.Close()
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		uuid := r.PathValue("uuid")
		var req v1.MoveCameraReq
		if err := dec.Decode(&req); err != nil {
			app.WriteJSON(w, r, api.ErrorEnvp{"error": err.Error()}, http.StatusBadRequest)
			return
		}

		dto := v1.MoveCameraReq{
			Translation: req.Translation,
			Zoom:        req.Zoom,
		}

		if err := app.PtzService.MoveCamera(ctx, uuid, dto); err != nil {
			switch {
			case errors.Is(err, repos.ErrNotFound):
				app.WriteJSON(w, r, api.ErrorEnvp{"error": "camera not found"}, http.StatusNotFound)
			case errors.Is(err, onvif.ErrNoPtz):
				app.WriteJSON(w, r, api.ErrorEnvp{"error": "ptz not supported for this camera"}, http.StatusUnprocessableEntity)
			case errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded):
				app.WriteJSON(w, r, api.ErrorEnvp{"error": "operation timed out"}, http.StatusGatewayTimeout)
			default:
				app.WriteJSON(w, r, api.ErrorEnvp{"error": err.Error()}, http.StatusBadRequest)
			}
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
