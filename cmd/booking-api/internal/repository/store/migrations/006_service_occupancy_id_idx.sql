-- +migrate Up
CREATE INDEX IF NOT EXISTS service_occupancy_id_index ON service(occupancy_id);

-- +migrate Down
DROP INDEX IF EXISTS service_occupancy_id_index;
