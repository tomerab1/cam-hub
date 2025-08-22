package onvif

import (
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type soapEnvelope[T any] struct {
	Body struct {
		Resp T `xml:",any"`
	} `xml:"Body"`
}

func parseResp[T any](resp *http.Response, out *T, logger *slog.Logger) {
	raw, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		logger.Error(err.Error())
		return
	}

	fmt.Println(string(raw))
	var env soapEnvelope[T]
	if err := xml.Unmarshal(raw, &env); err != nil {
		logger.Error("xml unmarshal envelope: " + err.Error())
		return
	}

	*out = env.Body.Resp
}
