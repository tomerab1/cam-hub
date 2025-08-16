package onvif

import (
	"encoding/xml"
	"io"
	"log/slog"
	"net/http"

	"github.com/use-go/onvif"
)

type OnvifClientParams struct {
	Xaddr    string
	Username string
	Password string
	Logger   *slog.Logger
}

type OnvifClient struct {
	logger *slog.Logger
	device *onvif.Device
}

func NewOnvifClient(params OnvifClientParams) (*OnvifClient, error) {
	dev, err := onvif.NewDevice(onvif.DeviceParams{
		Xaddr:    params.Xaddr,
		Username: params.Username,
		Password: params.Password,
	})

	if err != nil {
		return nil, err
	}

	return &OnvifClient{
		logger: params.Logger,
		device: dev,
	}, nil
}

func parseResp[T any](resp *http.Response, out *T, logger *slog.Logger) {
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)

	if err != nil {
		logger.Error(err.Error())
		return
	}

	xml.Unmarshal(raw, out)
}
