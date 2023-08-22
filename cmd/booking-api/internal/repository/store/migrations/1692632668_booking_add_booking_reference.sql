-- +migrate Up
ALTER TABLE IF EXISTS booking ADD COLUMN IF NOT EXISTS external_reference TEXT;

-- +migrate Down
ALTER TABLE IF EXISTS booking DROP COLUMN IF EXISTS external_reference;
