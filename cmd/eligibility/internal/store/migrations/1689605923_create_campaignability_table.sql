-- +migrate Up
CREATE TABLE IF NOT EXISTS campaignability (
    occupancy_id TEXT,
    account_id TEXT,
    reasons json,

    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE,
    UNIQUE(occupancy_id, account_id)
);

-- +migrate Down
DROP TABLE IF EXISTS campaignability;