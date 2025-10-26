package models

import (
	"time"
)

// Recordings represents a video recording entry in the system with its associated metadata.
// It contains information about the storage location, evidence data, scoring, and timing details.
//
// Fields:
//   - Id: Unique identifier for the recording
//   - CamUUID: Camera identifier that created the recording
//   - BucketName: Name of the storage bucket containing the recording
//   - VidBucketKey: Key to access the video file in the bucket
//   - BestFrameBucketKey: Key to access the best frame image in the bucket
//   - Evidence: Structured evidence data associated with the recording
//   - EvidenceRaw: Raw binary evidence data stored in the database (used because i had trouble with jsonb and pgx)
//   - Score: Numerical score/confidence level of the recording
//   - State: Current state of the recording in the system
//   - NeedsPublish: Flag indicating if the recording needs to be published
//   - PromotedAt: Timestamp when the recording was promoted
//   - RetentionDays: Number of days to retain the recording
//   - StartTs: Timestamp when the recording started
//   - EndTs: Timestamp when the recording ended
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

// Evidence represents a bounding box detection with confidence score.
// It contains coordinates for a rectangular region (x_min, y_min, x_max, y_max)
// and a confidence value indicating the detection certainty.
//
// Fields:
//   - Conf: Confidence score of the detection (0.0 to 1.0)
//   - Xmin: X coordinate of the top-left corner
//   - Ymin: Y coordinate of the top-left corner
//   - Xmax: X coordinate of the bottom-right corner
//   - Ymax: Y coordinate of the bottom-right corner
type Evidence struct {
	Conf float32 `json:"conf"`
	Xmin float32 `json:"x_min"`
	Ymin float32 `json:"y_min"`
	Xmax float32 `json:"x_max"`
	Ymax float32 `json:"y_max"`
}
