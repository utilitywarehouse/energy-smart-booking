package consumer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	smart "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/consumer"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/uw-labs/substrate"
)

func TestAccountOptOutConsumer(t *testing.T) {
	s := store.NewAccount(pool)
	defer truncateDB(t)

	handler := consumer.HandleAccountOptOut(s, nil, nil, true)

	ev1, err := makeMessage(&smart.AccountBookingOptOutAddedEvent{
		AccountId: "accountID",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{ev1})
	assert.NoError(t, err, "failed to handle account opt out event")

	account, err := s.GetAccount(t.Context(), "accountID")
	assert.NoError(t, err, "failed to get account")
	expected := store.Account{
		ID:       "accountID",
		PSRCodes: nil,
		OptOut:   true,
	}
	assert.Equal(t, expected, account, "mismatch")

	ev2, err := makeMessage(&smart.AccountBookingOptOutRemovedEvent{
		AccountId: "accountID",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{ev2})
	assert.NoError(t, err, "failed to handle account opt out event")

	account, err = s.GetAccount(t.Context(), "accountID")
	assert.NoError(t, err, "failed to get account")
	expected.OptOut = false
	assert.Equal(t, expected, account, "mismatch")
}
