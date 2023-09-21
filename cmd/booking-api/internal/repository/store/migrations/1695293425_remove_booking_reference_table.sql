-- +migrate Up
DROP TABLE IF EXISTS booking_reference;

-- +migrate Down
CREATE TABLE IF NOT EXISTS booking_reference (
    mpxn                   TEXT PRIMARY KEY,
    reference              TEXT NOT NULL,

    updated_at             TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at             TIMESTAMP WITHOUT TIME ZONE
);