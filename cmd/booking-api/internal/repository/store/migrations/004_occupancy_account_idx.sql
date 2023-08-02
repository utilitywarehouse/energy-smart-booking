-- +migrate Up
CREATE INDEX IF NOT EXISTS occupancy_account_id_idx ON occupancy(account_id);

-- +migrate Down
DROP INDEX IF EXISTS occupancy_account_id_idx;
