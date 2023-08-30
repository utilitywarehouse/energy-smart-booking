package store_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

func Test_BookingReferenceStore_Upsert(t *testing.T) {
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

	bookingReferenceStore := store.NewBookingReference(db)

	type inputParams struct {
		bookingReference models.BookingReference
	}

	type testSetup struct {
		description string
		input       inputParams
		output      error
	}

	testCases := []testSetup{
		{
			description: "should upsert a booking reference with mpxn: mpxn-1",
			input: inputParams{
				bookingReference: models.BookingReference{
					Reference: "REF-01",
					MPXN:      "mpxn-1",
				},
			},
			output: nil,
		},
		{
			description: "should not have a conflict after an insert with a collision on primary key (update on conflict)",
			input: inputParams{
				bookingReference: models.BookingReference{
					Reference: "REF-02",
					MPXN:      "mpxn-1",
				},
			},
			output: nil,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.description, func(t *testing.T) {

			bookingReferenceStore.Begin()

			bookingReferenceStore.Upsert(tc.input.bookingReference)

			err := bookingReferenceStore.Commit(ctx)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}

		})
	}
}

func Test_BookingReference_Remove(t *testing.T) {
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

	bookingReferenceStore := store.NewBookingReference(db)

	type inputParams struct {
		mpxn string
	}

	type testSetup struct {
		description string
		input       inputParams
		output      error
	}

	testCases := []testSetup{
		{
			description: "should remove booking reference for given mpxn",
			input: inputParams{
				mpxn: "mpxn",
			},
			output: nil,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.description, func(t *testing.T) {
			bookingReferenceStore.Begin()
			bookingReferenceStore.Remove(tc.input.mpxn)
			err = bookingReferenceStore.Commit(ctx)

			if err != nil {
				if tc.output != nil {
					if diff := cmp.Diff(err, tc.output, cmpopts.EquateErrors()); diff != "" {
						t.Fatal(diff)
					}
				} else {
					t.Fatal(err)
				}
			}
		})
	}
}
