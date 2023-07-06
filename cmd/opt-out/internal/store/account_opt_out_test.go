package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccountOptOut(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	store := NewAccountOptOut(connect(ctx))
	defer store.pool.Close()

	err := store.Add(ctx, "id1", "account_no_1", "user", time.Now())
	assert.NoError(err, "failed to add opt out account")

	account, err := store.Get(ctx, "id1")
	assert.NoError(err, "failed to get opt out account")
	assert.Equal(account.ID, "id1")

	err = store.Remove(ctx, "id1")
	assert.NoError(err, "failed to remove opt out account")

	accounts, err := store.List(ctx)
	assert.NoError(err, "failed to list opt out accounts")
	assert.Equal(0, len(accounts))
}
