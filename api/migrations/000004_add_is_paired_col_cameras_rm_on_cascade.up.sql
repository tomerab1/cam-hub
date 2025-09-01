ALTER TABLE cameras
ADD COLUMN isPaired BOOLEAN DEFAULT true;

ALTER TABLE camera_creds 
DROP CONSTRAINT camera_creds_id_fkey;

ALTER TABLE camera_creds 
ADD CONSTRAINT camera_creds_id_fkey 
FOREIGN KEY (id) REFERENCES cameras(id);