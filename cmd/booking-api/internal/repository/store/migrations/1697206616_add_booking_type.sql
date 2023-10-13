-- +migrate Up
ALTER TABLE IF EXISTS booking ADD COLUMN IF NOT EXISTS booking_type INTEGER;

-- +migrate Down
ALTER TABLE IF EXISTS booking DROP COLUMN IF EXISTS booking_type;
