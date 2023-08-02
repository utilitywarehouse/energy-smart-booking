package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/click-generator/internal/store"
)

func TestEvaluationStore(t *testing.T) {
	ctx := context.Background()

	testContainer, err := postgres.SetupTestContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	dsn, err := postgres.GetTestContainerDSN(testContainer)
	if err != nil {
		t.Fatal(err)
	}

	pool, err := store.Setup(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	db := store.NewSmartBookingEvaluation(pool)

	_, err = db.Get(ctx, "evaluation-account-A", "evaluation-occupancy-A")
	assert.ErrorIs(t, err, store.ErrEvaluationNotFound)

	err = db.UpsertEligibility(ctx, "evaluation-account-A", "evaluation-occupancy-A", false)
	assert.NoError(t, err)
	err = db.UpsertSuppliability(ctx, "evaluation-account-A", "evaluation-occupancy-A", true)
	assert.NoError(t, err)

	evaluation, err := db.Get(ctx, "evaluation-account-A", "evaluation-occupancy-A")
	assert.NoError(t, err)
	assert.Equal(t, store.Evaluation{
		AccountID:   "evaluation-account-A",
		OccupancyID: "evaluation-occupancy-A",
		Eligible:    false,
		Suppliable:  true,
	}, evaluation)

	err = db.UpsertEligibility(ctx, "evaluation-account-A", "evaluation-occupancy-A", true)
	assert.NoError(t, err)
	err = db.UpsertSuppliability(ctx, "evaluation-account-A", "evaluation-occupancy-A", false)
	assert.NoError(t, err)
	evaluation, err = db.Get(ctx, "evaluation-account-A", "evaluation-occupancy-A")
	assert.NoError(t, err)
	assert.Equal(t, store.Evaluation{
		AccountID:   "evaluation-account-A",
		OccupancyID: "evaluation-occupancy-A",
		Eligible:    true,
		Suppliable:  false,
	}, evaluation)
}
