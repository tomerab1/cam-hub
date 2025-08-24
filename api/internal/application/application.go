package application

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	v1 "tomerab.com/cam-hub/internal/services/v1"
)

type Application struct {
	Logger        *slog.Logger
	DB            *pgxpool.Pool
	HttpClient    *http.Client
	CameraService *v1.CameraService
}
