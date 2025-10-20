-- Drop indexes
DROP INDEX IF EXISTS ix_recordings_due;
DROP INDEX IF EXISTS ix_recordings_cam_time;

-- Drop constraints
ALTER TABLE recordings DROP CONSTRAINT IF EXISTS chk_retention_pos;
ALTER TABLE recordings DROP CONSTRAINT IF EXISTS chk_time_order;
ALTER TABLE recordings DROP CONSTRAINT IF EXISTS chk_score_0_1;

-- Drop tables
DROP TABLE IF EXISTS ptz_tokens;
DROP TABLE IF EXISTS camera_creds;
DROP TABLE IF EXISTS recordings;
DROP TABLE IF EXISTS cameras;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
