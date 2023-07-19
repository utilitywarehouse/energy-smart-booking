-- +migrate Up
CREATE TABLE IF NOT EXISTS meterpoints (
    mpxn TEXT PRIMARY KEY,
    supply_type TEXT NOT NULL,
    alt_han BOOLEAN default false,
    profile_class TEXT,
    ssc TEXT,

    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE,
    UNIQUE(mpxn)
);

CREATE INDEX meterpoints_idx ON meterpoints (mpxn);

-- +migrate Down
DROP TABLE IF EXISTS meterpoints;