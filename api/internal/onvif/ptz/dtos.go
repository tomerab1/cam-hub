package ptz

import "tomerab.com/cam-hub/internal/utils"

type MoveCameraDto struct {
	Token       string
	Translation *utils.Vec2D
	Zoom        *float32
}

type StopCameraMovementDto struct {
	Token string
}
