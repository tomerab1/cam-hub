package application

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"tomerab.com/cam-hub/internal/services"
)

type Application struct {
	Logger        *slog.Logger
	DB            *pgxpool.Pool
	HttpClient    *http.Client
	CameraService *services.CameraService
}
