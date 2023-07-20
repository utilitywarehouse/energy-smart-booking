-- +migrate Up
CREATE TABLE IF NOT EXISTS booking_references (
    mpxn TEXT PRIMARY KEY,
    reference TEXT NOT NULL ,

    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE
);

-- +migrate Down
DROP TABLE IF EXISTS booking_references;