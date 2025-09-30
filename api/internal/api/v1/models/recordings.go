package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
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
	Id                 string   `db:"id"`
	CamUUID            string   `db:"cam_id"`
	BucketName         string   `db:"bucket_name"`
	VidBucketKey       string   `db:"vid_key"`
	BestFrameBucketKey string   `db:"best_frame_key"`
	Evidence           Evidence `db:"evidence"`
	Score              float32  `db:"score"`
	State              string   `db:"state"`
	NeedsPublish       bool     `db:"needs_publish"`

	PromotedAt    time.Time `db:"promoted_at"`
	RetentionDays int       `db:"retention_days"`

	StartTs time.Time `db:"start_ts"`
	EndTs   time.Time `db:"end_ts"`
}

func (e *Evidence) Scan(value any) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("evidence: expected []byte, got %T", value)
	}

	return json.Unmarshal(b, e)
}

func (e Evidence) Value() (driver.Value, error) {
	return json.Marshal(e)
}
