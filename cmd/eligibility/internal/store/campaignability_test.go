package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
)

func TestCampaignability(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	s := NewCampaignability(connect(ctx))
	defer s.pool.Close()

	err := s.Add(ctx, "occupancy", "account", domain.IneligibleReasons{domain.IneligibleReasonAlreadySmart, domain.IneligibleReasonComplexTariff})
	assert.NoError(err, "failed to add campaignability")

	campaignability, err := s.Get(ctx, "occupancy", "account")
	assert.NoError(err, "failed to get campaignability")

	expected := Campaignability{
		OccupancyID: "occupancy",
		AccountID:   "account",
		Reasons:     domain.IneligibleReasons{domain.IneligibleReasonAlreadySmart, domain.IneligibleReasonComplexTariff},
	}
	assert.Equal(expected, campaignability, "mismatch")

	err = s.Add(ctx, "occupancy", "account", nil)
	assert.NoError(err, "failed to remove reasons from occupancy")

	campaignability, err = s.Get(ctx, "occupancy", "account")
	assert.NoError(err, "failed to get campaignability")
	expected.Reasons = nil
	assert.Equal(expected, campaignability, "mismatch")

	_, err = s.Get(ctx, "occupancy", "account1")
	assert.ErrorIs(err, ErrCampaignabilityNotFound)
}
