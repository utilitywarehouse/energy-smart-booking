package store_test

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/utilitywarehouse/energy-pkg/postgres"
)

func setupTestContainer(ctx context.Context) (testcontainers.Container, error) {
	return postgres.SetupTestContainer(ctx)
}

func populateDB(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `	
		INSERT INTO site (site_id,
			postcode,
			uprn,
			building_name_number,
			dependent_thoroughfare,
			thoroughfare,
			double_dependent_locality,
			dependent_locality,
			locality,
			county,
			town,
			department,
			organisation,
			po_box,
			delivery_point_suffix,
			sub_building_name_number)
		VALUES ('site-id-a', 'post-code-1', 'uprn', 'building-name-number',
		'dependent-thoroughfare', 'thoroughfare', 'double-dependent-locality',
		'dependent-locality', 'locality', 'county', 'town', 'department',
		'organisation', 'po-box', 'deliver-point-suffix', 'sub-building-name-number');

		INSERT INTO occupancy (occupancy_id, site_id, account_id, created_at)
		VALUES ('occupancy-id', 'site-id', 'account-id', NOW());

		INSERT INTO occupancy (occupancy_id, site_id, account_id, created_at)
		VALUES 
			('occupancy-id-A', 'site-id', 'account-id-sorted', '2023-01-01 00:00:00'),
			('occupancy-id-B', 'site-id', 'account-id-sorted', '2023-01-02 00:00:00'),
			('occupancy-id-C', 'site-id', 'account-id-sorted', '2023-01-03 00:00:00');

		INSERT INTO booking_reference (mpxn, reference)
		VALUES ('mpxn', 'reference');
	`,
	)

	return err
}
