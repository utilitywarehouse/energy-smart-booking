-- +migrate Up
CREATE TABLE IF NOT EXISTS point_of_sale_customer_details (
    account_number          TEXT PRIMARY KEY,
    details                 JSONB,

    created_at              TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    updated_at              TIMESTAMP WITHOUT TIME ZONE,
    deleted_at              TIMESTAMP WITHOUT TIME ZONE
);

-- +migrate Down
DROP TABLE IF EXISTS partial_booking;
