package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
)

func TestSuppliability(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	s := NewSuppliability(connect(ctx))
	defer s.pool.Close()

	err := s.Add(ctx, "occupancy", "account", domain.IneligibleReasons{domain.IneligibleReasonAlreadySmart, domain.IneligibleReasonComplexTariff})
	assert.NoError(err, "failed to add suppliability")

	suppliability, err := s.Get(ctx, "occupancy", "account")
	assert.NoError(err, "failed to get suppliability")

	expected := Suppliability{
		OccupancyID: "occupancy",
		AccountID:   "account",
		Reasons:     domain.IneligibleReasons{domain.IneligibleReasonAlreadySmart, domain.IneligibleReasonComplexTariff},
	}
	assert.Equal(expected, suppliability, "mismatch")

	err = s.Add(ctx, "occupancy", "account", nil)
	assert.NoError(err, "failed to remove reasons from occupancy")

	suppliability, err = s.Get(ctx, "occupancy", "account")
	assert.NoError(err, "failed to get suppliability")
	expected.Reasons = nil
	assert.Equal(expected, suppliability, "mismatch")

	_, err = s.Get(ctx, "occupancy", "account1")
	assert.ErrorIs(err, ErrSuppliabilityNotFound)
}
