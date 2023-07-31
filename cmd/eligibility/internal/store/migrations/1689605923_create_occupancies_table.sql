-- +migrate Up
CREATE TABLE IF NOT EXISTS occupancies (
     id TEXT PRIMARY KEY,
     site_id TEXT,
     account_id TEXT,

     created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
     updated_at TIMESTAMP WITHOUT TIME ZONE
);


CREATE INDEX occupancy_site_id_idx ON occupancies (site_id);

-- +migrate Down
DROP TABLE IF EXISTS occupancies;