package consumer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store/migrations"
	"github.com/utilitywarehouse/energy-smart-booking/internal/testcommon"
	"github.com/uw-labs/substrate"
)

func TestSuppliabilityConsumer(t *testing.T) {
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
	s := store.NewSuppliability(pool)

	handler := HandleSuppliability(s)

	suppEv1, err := testcommon.MakeMessage(&smart.SuppliableOccupancyAddedEvent{
		OccupancyId: "occupancyID",
		AccountId:   "accountID",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{suppEv1})
	assert.NoError(err, "failed to handle suppliability event")

	suppliability, err := s.Get(ctx, "occupancyID", "accountID")
	assert.NoError(err, "failed to get suppliability")
	expected := store.Suppliability{
		OccupancyID: "occupancyID",
		AccountID:   "accountID",
		Reasons:     nil,
	}
	assert.Equal(expected, suppliability, "suppliability mismatch")

	suppEv2, err := testcommon.MakeMessage(&smart.SuppliableOccupancyRemovedEvent{
		OccupancyId: "occupancyID",
		AccountId:   "accountID",
		Reasons: []smart.IneligibleReason{
			smart.IneligibleReason_INELIGIBLE_REASON_ALREADY_SMART,
			smart.IneligibleReason_INELIGIBLE_REASON_NOT_ACTIVE,
		},
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{suppEv2})
	assert.NoError(err, "failed to handle suppliability removed event")

	suppliability, err = s.Get(ctx, "occupancyID", "accountID")
	assert.NoError(err, "failed to get suppliability")
	expected.Reasons = domain.IneligibleReasons{domain.IneligibleReasonAlreadySmart, domain.IneligibleReasonNoActiveService}
	assert.Equal(expected, suppliability, "suppliability mismatch")
}
