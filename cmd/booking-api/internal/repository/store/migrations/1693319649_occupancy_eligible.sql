-- +migrate Up
CREATE TABLE IF NOT EXISTS occupancy_eligible (
    occupancy_id           TEXT PRIMARY KEY,
    reference              TEXT,

    created_at             TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at             TIMESTAMP WITHOUT TIME ZONE,
    deleted_at             TIMESTAMP WITHOUT TIME ZONE
);

-- +migrate Down
DROP TABLE IF EXISTS occupancy_eligible;
