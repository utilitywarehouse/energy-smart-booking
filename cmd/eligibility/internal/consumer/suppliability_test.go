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

func TestSuppliabilityConsumer(t *testing.T) {
	s := store.NewSuppliability(pool)
	defer truncateDB(t)

	handler := consumer.HandleSuppliability(s)

	suppEv1, err := makeMessage(&smart.SuppliableOccupancyAddedEvent{
		OccupancyId: "occupancyID",
		AccountId:   "accountID",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{suppEv1})
	assert.NoError(t, err, "failed to handle suppliability event")

	suppliability, err := s.Get(t.Context(), "occupancyID", "accountID")
	assert.NoError(t, err, "failed to get suppliability")
	expected := store.Suppliability{
		OccupancyID: "occupancyID",
		AccountID:   "accountID",
		Reasons:     nil,
	}
	assert.Equal(t, expected, suppliability, "suppliability mismatch")

	suppEv2, err := makeMessage(&smart.SuppliableOccupancyRemovedEvent{
		OccupancyId: "occupancyID",
		AccountId:   "accountID",
		Reasons: []smart.IneligibleReason{
			smart.IneligibleReason_INELIGIBLE_REASON_ALREADY_SMART,
			smart.IneligibleReason_INELIGIBLE_REASON_NOT_ACTIVE,
		},
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{suppEv2})
	assert.NoError(t, err, "failed to handle suppliability removed event")

	suppliability, err = s.Get(t.Context(), "occupancyID", "accountID")
	assert.NoError(t, err, "failed to get suppliability")
	expected.Reasons = domain.IneligibleReasons{domain.IneligibleReasonAlreadySmart, domain.IneligibleReasonNoActiveService}
	assert.Equal(t, expected, suppliability, "suppliability mismatch")
}
