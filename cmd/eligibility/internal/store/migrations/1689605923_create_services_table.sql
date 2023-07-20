-- +migrate Up
CREATE TABLE IF NOT EXISTS services (
    id TEXT PRIMARY KEY,
    mpxn TEXT NOT NULL,
    occupancy_id TEXT NOT NULL,
    supply_type TEXT NOT NULL,
    is_live BOOLEAN NOT NULL,
    start_date TIMESTAMP WITHOUT TIME ZONE,
    end_date TIMESTAMP WITHOUT TIME ZONE,

    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(id)
);

-- +migrate Down
DROP TABLE IF EXISTS services;