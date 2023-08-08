-- +migrate Up
CREATE INDEX IF NOT EXISTS sites_post_code_idx ON sites(post_code);
CREATE INDEX IF NOT EXISTS services_mpxn_idx ON services(mpxn);

-- +migrate Down
DROP INDEX IF EXISTS sites_post_code_idx;
DROP INDEX IF EXISTS services_mpxn_idx;
