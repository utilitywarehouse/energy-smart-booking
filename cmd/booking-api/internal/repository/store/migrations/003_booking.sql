-- +migrate Up
CREATE TABLE IF NOT EXISTS booking (
    booking_id             TEXT PRIMARY KEY,
    account_id             TEXT,
    status                 INT,

    -- address (normalized)
    site_id                TEXT,
    
    -- contact details
    contact_title          TEXT,
    contact_first_name     TEXT,
    contact_last_name      TEXT,
    contact_phone          TEXT,
    contact_email          TEXT,

    -- booking slot
    booking_date           DATE,
    booking_start_time     INT,
    booking_end_time       INT,

    -- vulnerability details
    vulnerabilities_list   INT[],
    vulnerabilities_other  TEXT,

    updated_at             TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX booking_account_id_idx ON booking(account_id);

-- +migrate Down
DROP TABLE IF EXISTS booking;
DROP INDEX IF EXISTS booking_account_id_idx;
