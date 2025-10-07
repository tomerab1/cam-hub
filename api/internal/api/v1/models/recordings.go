package models

import (
	"time"
)

type Evidence struct {
	Conf float32 `json:"conf"`
	Xmin float32 `json:"x_min"`
	Ymin float32 `json:"y_min"`
	Xmax float32 `json:"x_max"`
	Ymax float32 `json:"y_max"`
}

type Recordings struct {
	Id                 string    `json:"id" db:"id"`
	CamUUID            string    `json:"cam_id" db:"cam_id"`
	BucketName         string    `json:"bucket_name" db:"bucket_name"`
	VidBucketKey       string    `json:"vid_key" db:"vid_key"`
	BestFrameBucketKey string    `json:"best_frame_key" db:"best_frame_key"`
	Evidence           Evidence  `json:"evidence" db:"-"`
	EvidenceRaw        []byte    `json:"-" db:"evidence"`
	Score              float32   `json:"score" db:"score"`
	State              string    `json:"state" db:"state"`
	NeedsPublish       bool      `json:"needs_publish" db:"needs_publish"`
	PromotedAt         time.Time `json:"promoted_at" db:"promoted_at"`
	RetentionDays      int       `json:"retention_days" db:"retention_days"`
	StartTs            time.Time `json:"start_ts" db:"start_ts"`
	EndTs              time.Time `json:"end_ts" db:"end_ts"`
}
