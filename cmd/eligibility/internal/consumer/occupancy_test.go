package consumer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store/migrations"
	"github.com/utilitywarehouse/energy-smart-booking/internal/test_common"
	"github.com/uw-labs/substrate"
)

func TestOccupancyConsumer(t *testing.T) {
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
	s := store.NewOccupancy(pool)

	handler := HandleOccupancy(s, nil, true)

	occEv1, err := test_common.MakeMessage(&platform.OccupancyStartedEvent{
		OccupancyId:       "occupancyID",
		SiteId:            "siteID",
		CustomerAccountId: "customerAccID",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{occEv1})
	assert.NoError(err, "failed to handle occupancy discovered event")

	occ, err := s.Get(ctx, "occupancyID")
	assert.NoError(err, "failed to get occupancy")
	expected := store.Occupancy{
		ID:        "occupancyID",
		SiteID:    "siteID",
		AccountID: "customerAccID",
	}
	assert.Equal(expected, occ, "occupancy mismatch")

	occEv2, err := test_common.MakeMessage(&platform.OccupancySiteCorrectedEvent{
		OccupancyId: "occupancyID",
		SiteId:      "siteID1",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{occEv2})
	assert.NoError(err, "failed to handle occupancy site corrected event")

	occ, err = s.Get(ctx, "occupancyID")
	assert.NoError(err, "failed to get occupancy")
	expected.SiteID = "siteID1"
	assert.Equal(expected, occ, "occupancy mismatch")
}
