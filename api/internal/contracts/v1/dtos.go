package v1

import (
	"time"

	"tomerab.com/cam-hub/internal/utils"
)

type PairDeviceReq struct {
	Addr         string `json:"addr" validate:"required,hostname_port"`
	CameraName   string `json:"camera_name" validate:"required,min=1"`
	Username     string `json:"username" validate:"required,min=1"`
	Password     string `json:"password" validate:"required,min=4,max=8"`
	WifiName     string `json:"wifi_name" validate:"omitempty,min=1"`     // SSID
	WifiPassword string `json:"wifi_password" validate:"omitempty,min=1"` // PSK
}

type UnpairDeviceReq struct {
	UUID string `json:"uuid" validate:"required,uuid"`
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

type CameraProxyEvent struct {
	CameraPairedEvent   *CameraPairedEvent
	CameraUnpairedEvent *CameraUnpairedEvent
}

type AnalyzeImgsEvent struct {
	UUID       string   `json:"uuid"`
	Tp         string   `json:"tp"`
	VidPath    string   `json:"vid_path"`
	FramePaths []string `json:"frame_paths"`
}

type CameraUnpairedEvent struct {
	UUID string `json:"uuid"`
}

type CameraStreamUrl struct {
	Url string `json:"url"`
}

type MoveCameraReq struct {
	Translation *utils.Vec2D `json:"translation" validate:"omitempty,dive"`
	Zoom        *float32     `json:"zoom" validate:"omitempty,gte=-1,lte=1"`
}

type Evidence struct {
	Conf float32 `json:"conf"`
	Xmin float32 `json:"x_min"`
	Ymin float32 `json:"y_min"`
	Xmax float32 `json:"x_max"`
	Ymax float32 `json:"y_max"`
}

type AddRecordingReq struct {
	BucketName         string   `json:"bucket_name"`
	VidBucketKey       string   `json:"vid_key"`
	BestFrameBucketKey string   `json:"best_frame_key"`
	Evidence           Evidence `json:"evidence"`
	Score              float32  `json:"score"`

	RetentionDays int `json:"retention_days"`

	StartTs time.Time `json:"start_ts"`
	EndTs   time.Time `json:"end_ts"`
}

type SetLightModeReq struct {
	Mode string `json:"mode"`
}
