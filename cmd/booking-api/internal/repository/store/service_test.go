package store_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

func Test_ServiceStore_Upsert(t *testing.T) {
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

	serviceStore := store.NewService(db)

	type inputParams struct {
		service models.Service
	}

	type testSetup struct {
		description string
		input       inputParams
		output      error
	}

	testCases := []testSetup{
		{
			description: "should upsert a service with service-id-1",
			input: inputParams{
				service: models.Service{
					ServiceID:   "service-id-1",
					Mpxn:        "mpxn-1",
					OccupancyID: "occupancy-id-1",
					SupplyType:  domain.SupplyTypeElectricity,
					AccountID:   "account-id-1",
					StartDate:   nil,
					EndDate:     nil,
					IsLive:      true,
				},
			},
			output: nil,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.description, func(t *testing.T) {

			serviceStore.Begin()

			serviceStore.Upsert(tc.input.service)

			serviceStore.Commit(ctx)

			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}
		})

	}
}

func Test_ServiceStore_GetServiceMPXNByOccupancyID(t *testing.T) {
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

	serviceStore := store.NewService(db)

	type inputParams struct {
		occupancyID string
	}

	type outputParams struct {
		reference string
		err       error
	}

	type testSetup struct {
		description string
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the mpxn of the service associated with the occupancy ID",
			input: inputParams{
				occupancyID: "occupancy-id-ref-test",
			},
			output: outputParams{
				reference: "ref#1",
				err:       nil,
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.description, func(t *testing.T) {

			actualReference, err := serviceStore.GetReferenceByOccupancyID(ctx, tc.input.occupancyID)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(actualReference, tc.output.reference); diff != "" {
				t.Fatal(diff)
			}

		})

	}
}
