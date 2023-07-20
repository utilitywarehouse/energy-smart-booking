package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountPSR(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	store := NewAccountPSR(connect(ctx))
	defer store.pool.Close()

	err := store.Add(ctx, "accountID", []string{"14", "31"})
	assert.NoError(err, "failed to add account psr codes")

	codes, err := store.GetPSRCodes(ctx, "accountID")
	assert.NoError(err, "failed to retrieve account psr codes")
	assert.Equal([]string{"14", "31"}, codes, "mismatch")

	err = store.Add(ctx, "accountID", []string{"27"})
	codes, err = store.GetPSRCodes(ctx, "accountID")
	assert.NoError(err, "failed to retrieve account psr codes")
	assert.Equal([]string{"27"}, codes, "mismatch")

	err = store.Remove(ctx, "accountID")
	assert.NoError(err, "failed to remove account psr codes")

	_, err = store.GetPSRCodes(ctx, "accountID")
	assert.ErrorIs(err, ErrAccountPSRCodesNotFound)
}
