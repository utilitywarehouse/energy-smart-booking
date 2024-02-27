-- +migrate Up
ALTER TABLE IF EXISTS partial_booking ADD COLUMN IF NOT EXISTS deletion_reason INTEGER;

-- +migrate Down
ALTER TABLE IF EXISTS partial_booking DROP COLUMN IF EXISTS deletion_reason;
