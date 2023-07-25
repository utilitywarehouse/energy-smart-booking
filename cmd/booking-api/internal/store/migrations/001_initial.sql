-- +migrate Up
CREATE TABLE IF NOT EXISTS booking_reference (
    mpxn                   TEXT PRIMARY KEY,
    reference              TEXT NOT NULL,

    updated_at             TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS service (
    service_id             TEXT PRIMARY KEY,    
    mpxn                   TEXT NOT NULL,
    occupancy_id           TEXT NOT NULL,
    supply_type            TEXT NOT NULL,
    account_id             TEXT NOT NULL,
    start_date             TIMESTAMP WITHOUT TIME ZONE,
    end_date               TIMESTAMP WITHOUT TIME ZONE,
    is_live                BOOLEAN NOT NULL,

    updated_at             TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS occupancy (
    occupancy_id           TEXT PRIMARY KEY,
    site_id                TEXT NOT NULL,
    account_id             TEXT NOT NULL,

    updated_at             TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX occupancy_site_id_idx ON occupancy(site_id);

CREATE TABLE IF NOT EXISTS site (
    site_id                TEXT PRIMARY KEY,
    postcode               TEXT NOT NULL,

    updated_at             TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE IF EXISTS booking_reference;
DROP TABLE IF EXISTS service;
DROP TABLE IF EXISTS occupancy;
DROP TABLE IF EXISTS site;
