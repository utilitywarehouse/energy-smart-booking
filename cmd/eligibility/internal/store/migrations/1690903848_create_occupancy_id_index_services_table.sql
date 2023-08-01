-- +migrate Up
CREATE INDEX IF NOT EXISTS services_occupancy_id_idx ON services (occupancy_id);

-- +migrate Down
DROP INDEX IF EXISTS services_occupancy_id_idx;
