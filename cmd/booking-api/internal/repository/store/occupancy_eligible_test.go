package store_test

import (
	"testing"

	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

func Test_OccupancyEligibleStore_Upsert(t *testing.T) {
	occupancyEligibleStore := store.NewOccupancyEligible(pool)
	defer truncateDB(t)

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

			err := occupancyEligibleStore.Commit(t.Context())

			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}
		})

	}
}

func Test_OccupancyEligibleStore_Delete(t *testing.T) {
	occupancyEligibleStore := store.NewOccupancyEligible(pool)
	defer truncateDB(t)

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

			err := occupancyEligibleStore.Commit(t.Context())

			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}
		})

	}
}
