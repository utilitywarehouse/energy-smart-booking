package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSite(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	store := NewSite(connect(ctx))
	defer store.pool.Close()

	err := store.Add(ctx, "site1", "postcode1", time.Now())
	assert.NoError(err, "failed to add site")

	site, err := store.Get(ctx, "site1")
	assert.NoError(err, "failed to retrieve site")
	assert.Equal(Site{ID: "site1", PostCode: "postcode1"}, site, "mismatch")

	err = store.Add(ctx, "site1", "postcode2", time.Now())
	assert.NoError(err)

	site, err = store.Get(ctx, "site1")
	assert.NoError(err, "failed to retrieve site")
	assert.Equal(Site{ID: "site1", PostCode: "postcode2"}, site, "mismatch")

	_, err = store.Get(ctx, "site2")
	assert.ErrorIs(err, ErrSiteNotFound)
}
