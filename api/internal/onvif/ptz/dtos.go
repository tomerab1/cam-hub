package ptz

import "tomerab.com/cam-hub/internal/utils"

type MoveCameraDto struct {
	Token       string
	Translation utils.Vec2D
}

type StopCameraMovementDto struct {
	Token string
}
