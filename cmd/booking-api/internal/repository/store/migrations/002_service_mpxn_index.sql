-- +migrate Up
CREATE INDEX service_mpxn_idx ON service (mpxn);

-- +migrate Down
DROP INDEX IF EXISTS service_mpxn_idx;
