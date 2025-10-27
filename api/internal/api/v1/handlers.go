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

// filterUUIDS filters out discovered cameras that are already paired in the system.
// It takes a context, a camera repository interface, and a slice of WsDiscoveryMatch objects containing discovered cameras.
// The function extracts UUIDs from matches, checks which ones already exist in the database through the camera repository,
// and returns a filtered slice containing only unpaired cameras.
// Returns filtered matches slice and any error encountered during the database lookup.
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

// getDiscoveredDevices creates an HTTP handler function that discovers ONVIF cameras on the network.
// It performs both ONVIF WS-Discovery and internal discovery service operations.
//
// The handler:
// - Executes WS-Discovery with a 5 second timeout
// - Triggers the application's discovery service
// - Filters out already known devices by UUID
// - Returns the filtered discovery results as JSON
//
// Parameters:
//   - app: Application instance containing required services and dependencies
//
// Returns:
//   - http.HandlerFunc that handles the device discovery endpoint
//
// The response contains a WsDiscoveryDto with filtered matches of discovered devices.
// On error, returns a 500 Internal Server Error.
func getDiscoveredDevices(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		discovered := onvif.DiscoverNewCameras(ctx, app.Logger)
		app.DiscoveryService.Discover(r.Context())

		filteredMatches, err := filterUUIDS(ctx, app.CameraService.CamRepo, discovered.Matches)
		if err != nil {
			serverError(w, r, err, app.Logger)
			return
		}

		app.WriteJSON(w, r, discovery.WsDiscoveryDto{Matches: filteredMatches}, http.StatusOK)
	}
}

// pairCamera creates an HTTP handler function that pairs a camera.
//
// The handler:
// - Extracts the cameras uuid from the uri.
// - Calls CameraService Pair method.
//
// Parameters:
//   - app: Application instance containing required services and dependencies
//
// The response contains the camera model, returns HTTP status 201.
// On error, Returns a bad request error if the request is malformed.
// Returns an InternalServerError (500) otherwise.
func pairCamera(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		body := getValidatedBody[v1.PairDeviceReq](r)
		uuid := r.PathValue("uuid")

		camera, err := app.CameraService.Pair(ctx, uuid, body)
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
		keepAliveTicker := time.NewTicker(time.Second * 20)
		defer keepAliveTicker.Stop()

		for {
			select {
			case evt := <-app.SseChan:
				data, _ := json.Marshal(evt)
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			case <-keepAliveTicker.C:
				fmt.Fprint(w, ":\n\n")
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
		defer app.PubSub.Unsubscribe(uuid, subCh)
		keepAliveTicker := time.NewTicker(time.Second * 10)
		defer keepAliveTicker.Stop()

		for {
			select {
			case msg := <-subCh:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			case <-keepAliveTicker.C:
				fmt.Fprint(w, ":\n\n")
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
		body := getValidatedBody[v1.MoveCameraReq](r)

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		uuid := r.PathValue("uuid")
		dto := v1.MoveCameraReq{
			Translation: body.Translation,
			Zoom:        body.Zoom,
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
