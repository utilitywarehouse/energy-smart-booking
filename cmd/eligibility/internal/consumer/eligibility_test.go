package consumer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	smart "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/consumer"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/uw-labs/substrate"
)

func TestEligibilityConsumer(t *testing.T) {
	s := store.NewEligibility(pool)
	defer truncateDB(t)

	handler := consumer.HandleEligibility(s)

	eligibilityEv1, err := makeMessage(&smart.EligibleOccupancyAddedEvent{
		OccupancyId: "occupancyID",
		AccountId:   "accountID",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{eligibilityEv1})
	assert.NoError(t, err, "failed to handle eligibility event")

	eligibility, err := s.Get(t.Context(), "occupancyID", "accountID")
	assert.NoError(t, err, "failed to get eligibility")
	expected := store.Eligibility{
		OccupancyID: "occupancyID",
		AccountID:   "accountID",
		Reasons:     nil,
	}
	assert.Equal(t, expected, eligibility, "eligibility mismatch")

	eligibilityEv2, err := makeMessage(&smart.EligibleOccupancyRemovedEvent{
		OccupancyId: "occupancyID",
		AccountId:   "accountID",
		Reasons: []smart.IneligibleReason{
			smart.IneligibleReason_INELIGIBLE_REASON_ALREADY_SMART,
			smart.IneligibleReason_INELIGIBLE_REASON_COMPLEX_TARIFF,
		},
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{eligibilityEv2})
	assert.NoError(t, err, "failed to handle eligibility removed event")

	eligibility, err = s.Get(t.Context(), "occupancyID", "accountID")
	assert.NoError(t, err, "failed to get eligibility")
	expected.Reasons = domain.IneligibleReasons{domain.IneligibleReasonAlreadySmart, domain.IneligibleReasonComplexTariff}
	assert.Equal(t, expected, eligibility, "eligibility mismatch")
}
