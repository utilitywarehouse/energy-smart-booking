-- +migrate Up
CREATE INDEX IF NOT EXISTS occupancies_account_id_idx ON occupancies(account_id);

-- +migrate Down
DROP INDEX IF EXISTS occupancies_account_id_idx;
