-- +migrate Up
CREATE INDEX IF NOT EXISTS eligibility_occupancy_id_idx ON eligibility (occupancy_id);
CREATE INDEX IF NOT EXISTS suppliability_occupancy_id_idx ON suppliability (occupancy_id);
CREATE INDEX IF NOT EXISTS suppliability_occupancy_id_idx ON campaignability (occupancy_id);

-- +migrate Down
DROP INDEX IF EXISTS eligibility_occupancy_id_idx;
DROP INDEX IF EXISTS suppliability_occupancy_id_idx;
DROP INDEX IF EXISTS suppliability_occupancy_id_idx;

