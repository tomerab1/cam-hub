package v1

import (
	"time"

	"tomerab.com/cam-hub/internal/utils"
)

type PairDeviceReq struct {
	Addr       string `json:"addr"`
	CameraName string `json:"camera_name"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

type DiscoveryEvent struct {
	Type string    `json:"type"`
	UUID string    `json:"uuid"`
	Addr string    `json:"addr"`
	At   time.Time `json:"at"`
}

type CameraPairedEvent struct {
	UUID      string `json:"uuid"`
	StreamUrl string `json:"url"`
	Revision  int    `json:"revision"`
}

type CameraStreamUrl struct {
	Url string `json:"url"`
}

type MoveCameraReq struct {
	Translation *utils.Vec2D `json:"translation"`
	Zoom        *float32     `json:"zoom"`
}
