-- +migrate Up
CREATE TABLE IF NOT EXISTS smart_meter_interest (
    registration_id        TEXT PRIMARY KEY,
    account_id             TEXT NOT NULL,
    interested             BOOLEAN,
    reason                 TEXT,
    created_at             TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE IF EXISTS smart_meter_interest;

