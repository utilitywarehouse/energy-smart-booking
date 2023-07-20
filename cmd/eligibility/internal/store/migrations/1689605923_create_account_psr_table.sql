-- +migrate Up
CREATE TABLE IF NOT EXISTS account_psr (
    id TEXT PRIMARY KEY,
    psr_codes text[],

    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE
);

-- +migrate Down
DROP TABLE IF EXISTS account_psr;
