-- +migrate Up
CREATE TABLE IF NOT EXISTS opt_out_account (
   id TEXT NOT NULL PRIMARY KEY,
   number TEXT NOT NULL,
   added_by TEXT,
   created_at TIMESTAMP NOT NULL
);

CREATE INDEX account_id_idx ON opt_out_account (id);

-- +migrate Down
DROP TABLE IF EXISTS opt_out_account;