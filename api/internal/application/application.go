package application

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/events"
	"tomerab.com/cam-hub/internal/mtxapi"
	"tomerab.com/cam-hub/internal/services"
)

type Application struct {
	Logger           *slog.Logger
	DB               *pgxpool.Pool
	HttpClient       *http.Client
	CameraService    *services.CameraService
	DiscoveryService *services.DiscoveryService
	PtzService       *services.PtzService
	MtxClient        *mtxapi.MtxClient
	Bus              events.BusIface
	SseChan          chan v1.DiscoveryEvent
}

func (app *Application) WriteJSON(w http.ResponseWriter, r *http.Request, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		app.Logger.Warn("Error writing json response", "err", err)
	}
}
