package store_test

import (
	"context"
	"testing"

	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

func Test_OccupancyEligibleStore_Upsert(t *testing.T) {
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

	occupancyEligibleStore := store.NewOccupancyEligible(db)

	type inputParams struct {
		occupancy models.OccupancyEligibility
	}

	type testSetup struct {
		description string
		input       inputParams
		output      error
	}

	testCases := []testSetup{
		{
			description: "should upsert a occupancy eligible row with occupancy-id-1",
			input: inputParams{
				occupancy: models.OccupancyEligibility{
					OccupancyID: "occupancy-id-1",
					Reference:   "reference-1",
				},
			},
			output: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			occupancyEligibleStore.Begin()

			occupancyEligibleStore.Upsert(tc.input.occupancy)

			err := occupancyEligibleStore.Commit(ctx)

			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}
		})

	}
}

func Test_OccupancyEligibleStore_Delete(t *testing.T) {
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

	occupancyEligibleStore := store.NewOccupancyEligible(db)

	type inputParams struct {
		occupancy models.OccupancyEligibility
	}

	type testSetup struct {
		description string
		input       inputParams
		output      error
	}

	testCases := []testSetup{
		{
			description: "should mark the occupancy eligible as deleted for occupancy-id-1",
			input: inputParams{
				occupancy: models.OccupancyEligibility{
					OccupancyID: "occupancy-id-1",
				},
			},
			output: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			occupancyEligibleStore.Begin()

			occupancyEligibleStore.Delete(tc.input.occupancy)

			err := occupancyEligibleStore.Commit(ctx)

			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}
		})

	}
}
