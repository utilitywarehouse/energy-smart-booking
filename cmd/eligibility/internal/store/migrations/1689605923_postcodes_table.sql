-- +migrate Up
CREATE TABLE IF NOT EXISTS postcodes (
    post_code TEXT PRIMARY KEY,
    wan_coverage boolean not null,

    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(post_code)
);

-- +migrate Down
DROP TABLE IF EXISTS postcodes;