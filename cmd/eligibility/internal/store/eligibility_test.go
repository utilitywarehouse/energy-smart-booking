package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
)

func TestEligibility(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	s := NewEligibility(connect(ctx))
	defer s.pool.Close()

	err := s.Add(ctx, "occupancy", "account", domain.IneligibleReasons{domain.IneligibleReasonAlreadySmart, domain.IneligibleReasonComplexTariff})
	assert.NoError(err, "failed to add eligibility")

	eligibility, err := s.Get(ctx, "occupancy", "account")
	assert.NoError(err, "failed to get eligibility")

	expected := Eligibility{
		OccupancyID: "occupancy",
		AccountID:   "account",
		Reasons:     domain.IneligibleReasons{domain.IneligibleReasonAlreadySmart, domain.IneligibleReasonComplexTariff},
	}
	assert.Equal(expected, eligibility, "mismatch")

	err = s.Add(ctx, "occupancy", "account", nil)
	assert.NoError(err, "failed to remove reasons from occupancy")

	eligibility, err = s.Get(ctx, "occupancy", "account")
	assert.NoError(err, "failed to get eligibility")
	expected.Reasons = nil
	assert.Equal(expected, eligibility, "mismatch")

	_, err = s.Get(ctx, "occupancy", "account1")
	assert.ErrorIs(err, ErrEligibilityNotFound)
}
