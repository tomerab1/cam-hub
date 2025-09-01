package services

import (
	"log/slog"

	"tomerab.com/cam-hub/internal/repos"
)

type CameraCredsService struct {
	CameraCredsRep *repos.PgxCameraCredsRepo
	Logger         *slog.Logger
}
