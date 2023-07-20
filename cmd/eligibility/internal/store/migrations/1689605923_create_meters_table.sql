-- +migrate Up
CREATE TABLE IF NOT EXISTS meters (
    id TEXT PRIMARY KEY,
    mpxn TEXT,
    msn TEXT,
    supply_type TEXT NOT NULL,
    capacity NUMERIC,
    meter_type TEXT NOT NULL,

    created_at TIMESTAMP WITHOUT TIME ZONE,
    installed_at TIMESTAMP WITHOUT TIME ZONE,
    uninstalled_at TIMESTAMP WITHOUT TIME ZONE,
    UNIQUE(id)
);

-- +migrate Down
DROP TABLE IF EXISTS meters;