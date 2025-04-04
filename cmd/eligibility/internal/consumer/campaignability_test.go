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

func TestCampaignabilityConsumer(t *testing.T) {
	s := store.NewCampaignability(pool)
	defer truncateDB(t)

	handler := consumer.HandleCampaignability(s)

	campaignabilityEv1, err := makeMessage(&smart.CampaignableOccupancyAddedEvent{
		OccupancyId: "occupancyID",
		AccountId:   "accountID",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{campaignabilityEv1})
	assert.NoError(t, err, "failed to handle campaignability event")

	eligibility, err := s.Get(t.Context(), "occupancyID", "accountID")
	assert.NoError(t, err, "failed to get campaignability")
	expected := store.Campaignability{
		OccupancyID: "occupancyID",
		AccountID:   "accountID",
		Reasons:     nil,
	}
	assert.Equal(t, expected, eligibility, "campaignability mismatch")

	campaignabilityEv2, err := makeMessage(&smart.CampaignableOccupancyRemovedEvent{
		OccupancyId: "occupancyID",
		AccountId:   "accountID",
		Reasons: []smart.IneligibleReason{
			smart.IneligibleReason_INELIGIBLE_REASON_ALREADY_SMART,
			smart.IneligibleReason_INELIGIBLE_REASON_COMPLEX_TARIFF,
		},
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{campaignabilityEv2})
	assert.NoError(t, err, "failed to handle campaignability removed event")

	eligibility, err = s.Get(t.Context(), "occupancyID", "accountID")
	assert.NoError(t, err, "failed to get campaignability")
	expected.Reasons = domain.IneligibleReasons{domain.IneligibleReasonAlreadySmart, domain.IneligibleReasonComplexTariff}
	assert.Equal(t, expected, eligibility, "campaignability mismatch")
}
