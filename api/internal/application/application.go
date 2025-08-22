package application

import (
	"log/slog"

	"tomerab.com/cam-hub/internal/onvif"
)

type Application struct {
	Logger *slog.Logger
	Client *onvif.OnvifClient
}
