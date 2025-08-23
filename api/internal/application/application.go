package application

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Application struct {
	Logger     *slog.Logger
	DB         *pgxpool.Pool
	HttpClient *http.Client
}
