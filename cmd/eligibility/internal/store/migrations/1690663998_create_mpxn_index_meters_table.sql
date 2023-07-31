-- +migrate Up
CREATE INDEX IF NOT EXISTS meters_idx ON meters (mpxn);

-- +migrate Down
DROP INDEX IF EXISTS meters_idx;
