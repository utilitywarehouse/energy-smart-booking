package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/store"
)

func Test_OccupancyStore_Insert(t *testing.T) {
	ctx := context.Background()

	testContainer, err := setupTestContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	dsn, err := postgres.GetTestContainerDSN(testContainer)
	if err != nil {
		t.Fatal(err)
	}

	db, err := store.Setup(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	occupancyStore := store.NewOccupancy(db)

	type inputParams struct {
		occupancy models.Occupancy
	}

	type testSetup struct {
		description string
		input       inputParams
		output      error
	}

	testCases := []testSetup{
		{
			description: "should upsert a occupancy with occupancy-id-1",
			input: inputParams{
				occupancy: models.Occupancy{
					OccupancyID: "occupancy-id-1",
					SiteID:      "site-id-1",
					AccountID:   "account-id-1",
					CreatedAt:   time.Now(),
				},
			},
			output: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			occupancyStore.Begin()

			occupancyStore.Insert(tc.input.occupancy)

			err := occupancyStore.Commit(ctx)

			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}
		})

	}
}

func Test_OccupancyStore_UpdateSiteID(t *testing.T) {
	ctx := context.Background()

	testContainer, err := setupTestContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	dsn, err := postgres.GetTestContainerDSN(testContainer)
	if err != nil {
		t.Fatal(err)
	}

	db, err := store.Setup(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	err = populateDB(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	occupancyStore := store.NewOccupancy(db)

	type inputParams struct {
		occupancyID string
		siteID      string
	}

	type testSetup struct {
		description string
		input       inputParams
		output      error
	}

	testCases := []testSetup{
		{
			description: "should update the site id of occupancy with occupancy-id",
			input: inputParams{
				occupancyID: "occupancy-id",
				siteID:      "new-site-id",
			},
			output: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			occupancyStore.Begin()

			occupancyStore.UpdateSiteID(tc.input.occupancyID, tc.input.siteID)

			err = occupancyStore.Commit(ctx)

			if err != nil {
				if tc.output != nil {
					if diff := cmp.Diff(err.Error(), tc.output.Error()); diff != "" {
						t.Fatal(diff)
					}
				} else {
					t.Fatal(err)
				}
			}
		})

	}
}

func Test_OccupancyStore_GetOccupancyByID(t *testing.T) {
	ctx := context.Background()

	testContainer, err := setupTestContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	dsn, err := postgres.GetTestContainerDSN(testContainer)
	if err != nil {
		t.Fatal(err)
	}

	db, err := store.Setup(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	err = populateDB(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	occupancyStore := store.NewOccupancy(db)

	type inputParams struct {
		occupancyID string
	}

	type outputParams struct {
		occupancy *models.Occupancy
		err       error
	}

	type testSetup struct {
		description string
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the occupancy by occupancy-id",
			input: inputParams{
				occupancyID: "occupancy-id",
			},
			output: outputParams{
				occupancy: &models.Occupancy{
					OccupancyID: "occupancy-id",
					SiteID:      "site-id",
					AccountID:   "account-id",
				},
				err: nil,
			},
		},
		{
			description: "should fail to find a occupancy id because it does not exist",
			input: inputParams{
				occupancyID: "i do not exist",
			},
			output: outputParams{
				occupancy: nil,
				err:       store.ErrOccupancyNotFound,
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.description, func(t *testing.T) {
			occupancy, err := occupancyStore.GetOccupancyByID(ctx, tc.input.occupancyID)

			if err != nil {
				if tc.output.err != nil {
					if diff := cmp.Diff(err, tc.output.err, cmpopts.EquateErrors()); diff != "" {
						t.Fatal(diff)
					}
				} else {
					t.Fatal(err)
				}
			}

			if diff := cmp.Diff(occupancy, tc.output.occupancy); diff != "" {
				t.Fatal(diff)
			}
		})

	}
}

func Test_OccupancyStore_GetOccupanciesByAccountID(t *testing.T) {
	ctx := context.Background()

	testContainer, err := setupTestContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	dsn, err := postgres.GetTestContainerDSN(testContainer)
	if err != nil {
		t.Fatal(err)
	}

	db, err := store.Setup(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	err = populateDB(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	occupancyStore := store.NewOccupancy(db)

	type inputParams struct {
		accountID string
	}

	type outputParams struct {
		occupancies []models.Occupancy
		err         error
	}

	type testSetup struct {
		description string
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the occupancy by occupancy-id",
			input: inputParams{
				accountID: "account-id-sorted",
			},
			output: outputParams{
				occupancies: []models.Occupancy{
					{
						OccupancyID: "occupancy-id-C",
						SiteID:      "site-id",
						AccountID:   "account-id-sorted",
						CreatedAt:   time.Date(2023, time.January, 3, 0, 0, 0, 0, time.UTC),
					},
					{
						OccupancyID: "occupancy-id-B",
						SiteID:      "site-id",
						AccountID:   "account-id-sorted",
						CreatedAt:   time.Date(2023, time.January, 2, 0, 0, 0, 0, time.UTC),
					},
					{
						OccupancyID: "occupancy-id-A",
						SiteID:      "site-id",
						AccountID:   "account-id-sorted",
						CreatedAt:   time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.description, func(t *testing.T) {
			occupancy, err := occupancyStore.GetOccupanciesByAccountID(ctx, tc.input.accountID)

			if err != nil {
				if tc.output.err != nil {
					if diff := cmp.Diff(err, tc.output.err, cmpopts.EquateErrors()); diff != "" {
						t.Fatal(diff)
					}
				} else {
					t.Fatal(err)
				}
			}

			if diff := cmp.Diff(occupancy, tc.output.occupancies); diff != "" {
				t.Fatal(diff)
			}
		})

	}
}
