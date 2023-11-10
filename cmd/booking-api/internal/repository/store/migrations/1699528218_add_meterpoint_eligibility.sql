-- +migrate Up
CREATE TABLE IF NOT EXISTS meterpoint_eligible (
    mpxn                  TEXT PRIMARY KEY,
    eligible              TIMESTAMP WITHOUT TIME ZONE,

    updated_at            TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at            TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS meterpoint_eligible;
