CREATE TABLE IF NOT EXISTS cameras (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    manufacturer TEXT NOT NULL,
    model TEXT NOT NULL,
    firmwareVersion TEXT NOT NULL,
    serialNumber TEXT NOT NULL,
    hardwareId TEXT NOT NULL,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
)
