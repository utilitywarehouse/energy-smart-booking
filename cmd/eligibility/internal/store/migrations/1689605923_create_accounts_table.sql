-- +migrate Up
CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    psr_codes text[],
    opt_out boolean default FALSE,

    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE
);

-- +migrate Down
DROP TABLE IF EXISTS accounts;
