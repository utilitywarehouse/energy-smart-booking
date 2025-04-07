package store_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
)

var (
	dsn  string
	pool *pgxpool.Pool
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.SetupTestContainer(ctx)
	if err != nil {
		logrus.WithError(err).Panic("unable to create postgres test container")
	}
	defer func() {
		err := container.Terminate(ctx)
		if err != nil {
			logrus.WithError(err).Panic("unable to terminate test container")
		}
	}()

	dsn, err = postgres.GetTestContainerDSN(container)
	if err != nil {
		logrus.WithError(err).Panic("unable to get dsn from test container")
	}

	pool, err = store.Setup(ctx, dsn)
	if err != nil {
		logrus.WithError(err).Panic("unable to connect to test container postgres")
	}

	m.Run()
}

func truncateDB(t *testing.T) {
	_, err := pool.Exec(t.Context(), "TRUNCATE TABLE service, occupancy, site, booking, occupancy_eligible, partial_booking, smart_meter_interest CASCADE")
	if err != nil {
		t.Fatalf("unable to truncate seeded data: %s", err)
	}
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

		-- SETUP FOR GETTING LB KEYS --
		INSERT INTO occupancy (occupancy_id, site_id, account_id, created_at)
		VALUES ('occupancy-id-#1', 'site-id-a', 'account-id-#1', NOW());

		INSERT INTO occupancy_eligible (occupancy_id, reference)
		VALUES ('occupancy-id-#1', 'ref##1');

		INSERT INTO occupancy (occupancy_id, site_id, account_id, created_at)
		VALUES 
			('occupancy-id-A', 'site-id', 'account-id-sorted', '2023-01-01 00:00:00'),
			('occupancy-id-B', 'site-id', 'account-id-sorted', '2023-01-02 00:00:00'),
			('occupancy-id-C', 'site-id', 'account-id-sorted', '2023-01-03 00:00:00'),
			('occupancy-id-D', 'site-id', 'account-id-sorted', '2023-01-04 00:00:00');

		INSERT INTO service (service_id, mpxn, occupancy_id, supply_type, account_id, start_date, end_date, is_live)
		VALUES
			('service-id-1', 'mpxn-1', 'occupancy-id-A', 'gas', 'account-id', NOW(), NOW(), true),
			('service-id-2', 'mpxn-1', 'occupancy-id-B', 'gas', 'account-id', NOW(), NOW(), true),
			('service-id-3', 'mpxn-1', 'occupancy-id-C', 'gas', 'account-id', NOW(), NOW(), true),
			('service-id-4', 'mpxn-1', 'occupancy-id-D', 'gas', 'account-id', NOW(), NOW(), false);

		INSERT INTO service (service_id, mpxn, occupancy_id, supply_type, account_id, start_date, end_date, is_live)
		VALUES
			('service-id-001G', 'mpxn-ref#1', 'occupancy-id-ref-test', 'gas', 'account-id', NOW(), NOW(), true),
			('service-id-001E', 'mpxn-ref#2', 'occupancy-id-ref-test', 'electricity', 'account-id', NOW(), NOW(), true);

		INSERT INTO booking (
			booking_id,
			account_id,
			status,
	
			occupancy_id,
	
			contact_title,
			contact_first_name,
			contact_last_name,
			contact_phone,
			contact_email,
	
			booking_date,
			booking_start_time,
			booking_end_time,
	
			vulnerabilities_list,
			vulnerabilities_other,
			external_reference,
	
			booking_type
		) VALUES ('booking-id-1', 'account-id-1', 1, 'occupancy-id-1', 'Mr', 'John', 'Doe', '333-100', 'jdoe@example.com', NOW(), 10, 14, '{}', 'nothing', 'lbg100', 1),
				('booking-id-2', 'account-id-2', 1, 'occupancy-id-2', 'Mr', 'John', 'Doe', '400-500', 'jdoe@example.com', ('2024-04-01 00:00:00'), 10, 16, '{}', 'weak jaw', 'booking-reference-1', 2)
		ON CONFLICT (booking_id)
		DO NOTHING;
	`,
	)

	return err
}
