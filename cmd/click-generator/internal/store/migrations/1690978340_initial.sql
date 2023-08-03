-- +migrate Up
CREATE TABLE IF NOT EXISTS smart_booking_evaluation (
    account_id TEXT NOT NULL,
    occupancy_id TEXT NOT NULL,
    eligible BOOLEAN DEFAULT FALSE,
    suppliable BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE,

    UNIQUE (account_id, occupancy_id)
);

CREATE INDEX IF NOT EXISTS smart_booking_evaluation_account_id_idx on smart_booking_evaluation(account_id);
CREATE INDEX IF NOT EXISTS smart_booking_evaluation_occupancy_id_idx on smart_booking_evaluation(occupancy_id);

CREATE TABLE IF NOT EXISTS account_links (
    account_id TEXT NOT NULL,
    occupancy_id TEXT NOT NULL,
    link TEXT NOT NULL,

    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE,

    UNIQUE (account_id, occupancy_id)
);

-- +migrate Down
DROP TABLE IF EXISTS smart_booking_evaluation;