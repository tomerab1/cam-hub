-- Create cameras table
CREATE TABLE IF NOT EXISTS cameras (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    manufacturer TEXT NOT NULL,
    model TEXT NOT NULL,
    firmware_version TEXT NOT NULL,
    serial_number TEXT NOT NULL,
    hardware_id TEXT NOT NULL,
    addr TEXT NOT NULL,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create camera creds table
CREATE TABLE IF NOT EXISTS camera_creds (
    id uuid PRIMARY KEY REFERENCES cameras(id) ON DELETE CASCADE,
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create camera ptz tokens table
CREATE TABLE IF NOT EXISTS ptz_tokens(
    id uuid PRIMARY KEY REFERENCES cameras(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Add uuid extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create recordings table
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

-- Add constraints
ALTER TABLE recordings
  ADD CONSTRAINT chk_score_0_1 CHECK (score >= 0.0 AND score <= 1.0);

ALTER TABLE recordings
  ADD CONSTRAINT chk_time_order
  CHECK (start_ts IS NULL OR end_ts IS NULL OR start_ts <= end_ts);

ALTER TABLE recordings
  ADD CONSTRAINT chk_retention_pos
  CHECK (retention_days > 0);

-- Add indexes
CREATE INDEX IF NOT EXISTS ix_recordings_cam_time
  ON recordings (cam_id, promoted_at DESC);

CREATE INDEX IF NOT EXISTS ix_recordings_due
  ON recordings ((promoted_at + (retention_days * INTERVAL '1 day')));