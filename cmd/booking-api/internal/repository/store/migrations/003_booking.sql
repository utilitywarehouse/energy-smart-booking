-- +migrate Up
CREATE TABLE IF NOT EXISTS booking (
    booking_id             TEXT PRIMARY KEY,
    account_id             TEXT NOT NULL,
    status                 INT NOT NULL,

    -- address (normalized)
    occupancy_id           TEXT NOT NULL,
    
    -- contact details
    contact_title          TEXT NOT NULL,
    contact_first_name     TEXT NOT NULL,
    contact_last_name      TEXT NOT NULL,
    contact_phone          TEXT NOT NULL,
    contact_email          TEXT NOT NULL,

    -- booking slot
    booking_date           DATE NOT NULL,
    booking_start_time     INT NOT NULL,
    booking_end_time       INT NOT NULL,

    -- vulnerability details
    vulnerabilities_list   INT[] NOT NULL,
    vulnerabilities_other  TEXT NOT NULL,

    updated_at             TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX booking_account_id_idx ON booking(account_id);

-- +migrate Down
DROP TABLE IF EXISTS booking;
DROP INDEX IF EXISTS booking_account_id_idx;
