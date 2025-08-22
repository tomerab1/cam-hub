package onvif

import (
	"log/slog"

	"github.com/IOTechSystems/onvif"
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
