package consumer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	smart "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/consumer"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/uw-labs/substrate"
)

func TestAccountPSRConsumer(t *testing.T) {
	s := store.NewAccount(pool)
	defer truncateDB(t)

	handler := consumer.HandleAccountPSR(s, nil, nil, true)

	ev1, err := makeMessage(&smart.AccountPSRCodesChangedEvent{
		AccountId: "accountID",
		Codes:     []string{"12", "45"},
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{ev1})
	assert.NoError(t, err, "failed to handle account psr event")

	account, err := s.GetAccount(t.Context(), "accountID")
	assert.NoError(t, err, "failed to get account")
	expected := store.Account{
		ID:       "accountID",
		PSRCodes: []string{"12", "45"},
	}
	assert.Equal(t, expected, account, "mismatch")

	ev2, err := makeMessage(&smart.AccountPSRCodesRemovedEvent{
		AccountId: "accountID",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{ev2})
	assert.NoError(t, err, "failed to handle account psr event")

	account, err = s.GetAccount(t.Context(), "accountID")
	assert.NoError(t, err, "failed to get account")
	expected.PSRCodes = nil
	assert.Equal(t, expected, account, "mismatch")
}
