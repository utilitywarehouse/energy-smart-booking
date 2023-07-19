-- +migrate Up
CREATE TABLE IF NOT EXISTS sites (
    id TEXT PRIMARY KEY,
    post_code TEXT,

    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE
);

CREATE INDEX sites_idx ON sites (id);

-- +migrate Down
DROP TABLE IF EXISTS sites;