-- +migrate Up
ALTER TABLE IF EXISTS booking_references ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITHOUT TIME ZONE;

-- +migrate Down
ALTER TABLE IF EXISTS booking_references DROP COLUMN IF EXISTS deleted_at;
