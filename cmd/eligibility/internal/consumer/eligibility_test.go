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
	"github.com/utilitywarehouse/energy-smart-booking/internal/test_common"
	"github.com/uw-labs/substrate"
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
	s := store.NewEligibility(pool)

	handler := HandleEligibility(s)

	eligibilityEv1, err := test_common.MakeMessage(&smart.EligibleOccupancyAddedEvent{
		OccupancyId: "occupancyID",
		AccountId:   "accountID",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{eligibilityEv1})
	assert.NoError(err, "failed to handle eligibility event")

	eligibility, err := s.Get(ctx, "occupancyID", "accountID")
	assert.NoError(err, "failed to get eligibility")
	expected := store.Eligibility{
		OccupancyID: "occupancyID",
		AccountID:   "accountID",
		Reasons:     nil,
	}
	assert.Equal(expected, eligibility, "eligibility mismatch")

	eligibilityEv2, err := test_common.MakeMessage(&smart.EligibleOccupancyRemovedEvent{
		OccupancyId: "occupancyID",
		AccountId:   "accountID",
		Reasons: []smart.IneligibleReason{
			smart.IneligibleReason_INELIGIBLE_REASON_ALREADY_SMART,
			smart.IneligibleReason_INELIGIBLE_REASON_COMPLEX_TARIFF,
		},
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{eligibilityEv2})
	assert.NoError(err, "failed to handle eligibility removed event")

	eligibility, err = s.Get(ctx, "occupancyID", "accountID")
	assert.NoError(err, "failed to get eligibility")
	expected.Reasons = domain.IneligibleReasons{domain.IneligibleReasonAlreadySmart, domain.IneligibleReasonComplexTariff}
	assert.Equal(expected, eligibility, "eligibility mismatch")
}
