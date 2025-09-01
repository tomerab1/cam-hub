ALTER TABLE cameras
DROP COLUMN isPaired;

ALTER TABLE camera_creds
DROP CONSTRAINT camera_creds_id_fkey;

ALTER TABLE camera_creds
ADD CONSTRAINT camera_creds_id_fkey
FOREIGN KEY (id) REFERENCES cameras(id) ON DELETE CASCADE;