-- +migrate Up
ALTER TABLE booking
ALTER COLUMN vulnerabilities_list SET DEFAULT '{}'::INTEGER[];

-- +migrate Down
ALTER TABLE booking
ALTER COLUMN vulnerabilities_list DROP DEFAULT;
