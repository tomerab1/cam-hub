package v1

import (
	"encoding/json"
	"net/http"

	"tomerab.com/cam-hub/internal/onvif"
)

func getDiscoveredDevices(w http.ResponseWriter, r *http.Request) {
	logger := appFromCtx(r.Context()).Logger
	discovered := onvif.DiscoverNewCameras(logger)

	raw, err := json.Marshal(discovered)
	if err != nil {
		serverError(w, r, err, logger)
	}
	w.Write(raw)
}

func getProfiles(w http.ResponseWriter, r *http.Request) {
	client := appFromCtx(r.Context()).Client
	profiles := client.GetProfiles()

	w.Header().Add("Content-Type", "application/json")
	w.Write(profiles)
}
