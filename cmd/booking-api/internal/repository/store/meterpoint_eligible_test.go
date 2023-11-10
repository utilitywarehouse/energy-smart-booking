package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
)

func Test_MeterpointEligibleStore_Upsert(t *testing.T) {
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

	meterpointEligibleStore := store.NewMeterpointEligible(db, 6*time.Hour)

	type inputParams struct {
		mpxn      string
		eligible  *time.Time
		expiresAt time.Time
	}

	type testSetup struct {
		description string
		input       inputParams
		output      error
	}

	now := time.Now()

	testCases := []testSetup{
		{
			description: "should upsert a meterpoint eligible record",
			input: inputParams{
				mpxn:      "mpxn-1",
				eligible:  &now,
				expiresAt: time.Now().Add(time.Hour),
			},
			output: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err := meterpointEligibleStore.Upsert(ctx, tc.input.mpxn, tc.input.eligible, tc.input.expiresAt)
			if err != tc.output {
				t.Fatalf("error output does not match, expected: %s | actual: %s", tc.output, err)
			}
		})
	}
}
