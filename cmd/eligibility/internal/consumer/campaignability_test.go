package consumer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	smart "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store/migrations"
	"github.com/utilitywarehouse/energy-smart-booking/internal/testcommon"
	"github.com/uw-labs/substrate"
)

func TestCampaignabilityConsumer(t *testing.T) {
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
	s := store.NewCampaignability(pool)

	handler := HandleCampaignability(s)

	campaignabilityEv1, err := testcommon.MakeMessage(&smart.CampaignableOccupancyAddedEvent{
		OccupancyId: "occupancyID",
		AccountId:   "accountID",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{campaignabilityEv1})
	assert.NoError(err, "failed to handle campaignability event")

	eligibility, err := s.Get(ctx, "occupancyID", "accountID")
	assert.NoError(err, "failed to get campaignability")
	expected := store.Campaignability{
		OccupancyID: "occupancyID",
		AccountID:   "accountID",
		Reasons:     nil,
	}
	assert.Equal(expected, eligibility, "campaignability mismatch")

	campaignabilityEv2, err := testcommon.MakeMessage(&smart.CampaignableOccupancyRemovedEvent{
		OccupancyId: "occupancyID",
		AccountId:   "accountID",
		Reasons: []smart.IneligibleReason{
			smart.IneligibleReason_INELIGIBLE_REASON_ALREADY_SMART,
			smart.IneligibleReason_INELIGIBLE_REASON_COMPLEX_TARIFF,
		},
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{campaignabilityEv2})
	assert.NoError(err, "failed to handle campaignability removed event")

	eligibility, err = s.Get(ctx, "occupancyID", "accountID")
	assert.NoError(err, "failed to get campaignability")
	expected.Reasons = domain.IneligibleReasons{domain.IneligibleReasonAlreadySmart, domain.IneligibleReasonComplexTariff}
	assert.Equal(expected, eligibility, "campaignability mismatch")
}
