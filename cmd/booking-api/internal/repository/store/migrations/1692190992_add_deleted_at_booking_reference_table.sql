-- +migrate Up
ALTER TABLE IF EXISTS booking_reference ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITHOUT TIME ZONE;

-- +migrate Down
ALTER TABLE IF EXISTS booking_reference DROP COLUMN IF EXISTS deleted_at;
