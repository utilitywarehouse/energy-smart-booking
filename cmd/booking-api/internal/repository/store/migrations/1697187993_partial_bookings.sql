-- +migrate Up
CREATE TABLE IF NOT EXISTS partial_booking (
    booking_id              TEXT PRIMARY KEY,
    event                   JSONB,

    created_at              TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    updated_at              TIMESTAMP WITHOUT TIME ZONE,
    deleted_at              TIMESTAMP WITHOUT TIME ZONE,
    retries                 INT NOT NULL DEFAULT 0
);

-- +migrate Down
DROP TABLE IF EXISTS partial_booking;
