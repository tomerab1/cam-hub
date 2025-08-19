package onvif

import (
	"encoding/xml"
	"io"
	"log/slog"
	"net/http"
)

func parseResp[T any](resp *http.Response, out *T, logger *slog.Logger) {
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)

	if err != nil {
		logger.Error(err.Error())
		return
	}

	xml.Unmarshal(raw, out)
}
