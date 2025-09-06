package application

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/services"
)

type Application struct {
	Logger           *slog.Logger
	DB               *pgxpool.Pool
	HttpClient       *http.Client
	CameraService    *services.CameraService
	DiscoveryService *services.DiscoveryService
	SseChan          chan v1.DiscoveryEvent
}
