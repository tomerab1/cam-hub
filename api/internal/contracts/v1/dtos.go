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

type AnalyzeImgsEvent struct {
	UUID       string   `json:"uuid"`
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
	Translation *utils.Vec2D `json:"translation"`
	Zoom        *float32     `json:"zoom"`
}

type AddRecordingReq struct {
	BucketName         string `json:"bucket_name"`
	VidBucketKey       string `json:"vid_key"`
	BestFrameBucketKey string `json:"best_frame_key"`
	Evidence           struct {
		Conf float32 `json:"conf"`
		Xmin float32 `json:"x_min"`
		Ymin float32 `json:"y_min"`
		Xmax float32 `json:"x_max"`
		Ymax float32 `json:"y_max"`
	} `json:"evidence"`
	Score float32 `json:"score"`

	RetentionDays int `json:"retention_days"`

	StartTs time.Time `json:"start_ts"`
	EndTs   time.Time `json:"end_ts"`
}
