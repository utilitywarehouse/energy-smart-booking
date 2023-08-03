-- +migrate Up

CREATE TABLE IF NOT EXISTS account_links (
    account_number TEXT NOT NULL,
    link TEXT NOT NULL,

    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE,

    UNIQUE (account_number)
);

-- +migrate Down
DROP TABLE IF EXISTS account_links;