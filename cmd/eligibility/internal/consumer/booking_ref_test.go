package consumer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	smart "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/consumer"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/uw-labs/substrate"
)

func TestBookingRefConsumer(t *testing.T) {
	s := store.NewBookingRef(pool)
	occ := store.NewOccupancy(pool)
	defer truncateDB(t)

	handler := consumer.HandleBookingRef(s, occ, nil, false)

	ev1, err := makeMessage(&smart.BookingMpxnReferenceCreatedEvent{
		Mpxn:      "mpxn1",
		Reference: "ref1",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{ev1})
	assert.NoError(t, err, "failed to handle booking ref event")

	ref, err := s.GetReference(t.Context(), "mpxn1")
	assert.NoError(t, err, "failed to get booking reference")
	assert.Equal(t, "ref1", ref, "mismatch")
}
