package consumer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/click-generator/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/click-generator/internal/store/migrations"
	"github.com/utilitywarehouse/energy-smart-booking/internal/test_common"
)

func TestEligibilityConsumer(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)
	container, err := postgres.SetupTestContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(ctx)

	postgresURL, err := postgres.GetTestContainerDSN(container)
	if err != nil {
		t.Fatal(err)
	}

	pool, err := postgres.Setup(ctx, postgresURL, migrations.Source)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = postgres.Teardown(pool, migrations.Source); err != nil {
			t.Fatal(err)
		}
	}()

	s := store.NewSmartBookingEvaluation(pool)
	linkS := store.NewLink(pool)

	handler := EligibilityHandler{store: s, linkStore: linkS}
	ev1, err := test_common.MakeMessage(&smart.EligibleOccupancyRemovedEvent{
		AccountId:   "account-E",
		OccupancyId: "occupancy-E",
	})
	assert.NoError(err)
	ev2, err := test_common.MakeMessage(&smart.EligibleOccupancyAddedEvent{
		AccountId:   "account-E",
		OccupancyId: "occupancy-E",
	})

	err = handler.Handle(ctx, ev1)
	assert.NoError(err)
	err = handler.Handle(ctx, ev2)
	assert.NoError(err)

	evaluation, err := s.Get(ctx, "account-E", "occupancy-E")
	assert.NoError(err)
	assert.Equal(store.Evaluation{
		AccountID:   "account-E",
		OccupancyID: "occupancy-E",
		Eligible:    true,
		Suppliable:  false,
	}, evaluation)
}
