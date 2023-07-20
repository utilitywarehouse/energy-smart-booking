package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountPSR(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	store := NewAccount(connect(ctx))
	defer store.pool.Close()

	err := store.AddPSRCodes(ctx, "accountID", []string{"14", "31"})
	assert.NoError(err, "failed to add account psr codes")

	err = store.AddOptOut(ctx, "accountID", true)
	assert.NoError(err, "failed to add account opt out")

	account, err := store.GetAccount(ctx, "accountID")
	assert.NoError(err, "failed to retrieve account")
	expected := Account{
		ID:       "accountID",
		PSRCodes: []string{"14", "31"},
		OptOut:   true,
	}
	assert.Equal(expected, account, "mismatch")

	err = store.AddPSRCodes(ctx, "accountID", []string{"27"})
	assert.NoError(err, "failed to update account")

	account, err = store.GetAccount(ctx, "accountID")
	assert.NoError(err, "failed to retrieve account")
	expected.PSRCodes = []string{"27"}
	assert.Equal(expected, account, "mismatch")

	err = store.AddPSRCodes(ctx, "accountID", nil)
	assert.NoError(err, "failed to update psr codes")

	account, err = store.GetAccount(ctx, "accountID")
	assert.NoError(err, "failed to get account")
	expected.PSRCodes = nil
	assert.Equal(expected, account, "mismatch")

	err = store.AddOptOut(ctx, "accountID", false)
	assert.NoError(err, "failed to add account opt out")

	account, err = store.GetAccount(ctx, "accountID")
	assert.NoError(err, "failed to get account")
	expected.OptOut = false
	assert.Equal(expected, account, "mismatch")
}
