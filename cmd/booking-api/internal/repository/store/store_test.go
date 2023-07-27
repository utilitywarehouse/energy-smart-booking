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
			address)
		VALUES ('site-id-a', '{"postcode":"post-code-1","uprn":"uprn","building_name_number":"building-name-number","sub_building_name_number":"sub-building-name-number","dependent_thoroughfare":"dependent-thoroughfare","thoroughfare":"thoroughfare","double_dependent_locality":"double-dependent-locality","dependent_locality":"dependent-locality","locality":"locality","county":"county","town":"town","department":"department","organisation":"organisation","po_box":"po-box","delivery_point_suffix":"deliver-point-suffix"}');

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
