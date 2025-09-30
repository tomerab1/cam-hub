CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS recordings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cam_id UUID NOT NULL REFERENCES cameras(id) ON DELETE CASCADE,
    bucket_name TEXT NOT NULL,
    vid_key TEXT NOT NULL UNIQUE,
    best_frame_key TEXT,
    evidence JSONB NOT NULL,
    score REAL NOT NULL,
    state TEXT NOT NULL CHECK (state IN ('promoted','discarded')),
    needs_publish BOOLEAN NOT NULL,

    promoted_at TIMESTAMP NOT NULL DEFAULT (timezone('UTC', now()))::timestamp,
    retention_days INT NOT NULL DEFAULT 14,

    start_ts TIMESTAMP,
    end_ts TIMESTAMP
);

ALTER TABLE recordings
  ADD CONSTRAINT chk_score_0_1 CHECK (score >= 0.0 AND score <= 1.0);

ALTER TABLE recordings
  ADD CONSTRAINT chk_time_order
  CHECK (start_ts IS NULL OR end_ts IS NULL OR start_ts <= end_ts);

ALTER TABLE recordings
  ADD CONSTRAINT chk_retention_pos
  CHECK (retention_days > 0);

CREATE INDEX IF NOT EXISTS ix_recordings_cam_time
  ON recordings (cam_id, promoted_at DESC);

CREATE INDEX IF NOT EXISTS ix_recordings_due
  ON recordings ((promoted_at + (retention_days * INTERVAL '1 day')));