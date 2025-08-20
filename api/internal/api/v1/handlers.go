package v1

import (
	"encoding/json"
	"net/http"

	"tomerab.com/cam-hub/internal/onvif"
)

func GetDiscoveredDevices(w http.ResponseWriter, r *http.Request) {
	logger := LoggerFromCtx(r.Context())
	discovered := onvif.DiscoverNewCameras(logger)

	raw, err := json.Marshal(discovered)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	w.Write(raw)
}
