package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBookingRef(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	store := NewBookingRef(connect(ctx))
	defer store.pool.Close()

	err := store.Add(ctx, "mpxn1", "ref1")
	assert.NoError(err, "failed to add booking ref")

	ref, err := store.GetReference(ctx, "mpxn1")
	assert.NoError(err, "failed to retrieve booking ref")
	assert.Equal("ref1", ref, "mismatch")

	err = store.Add(ctx, "mpxn1", "ref2")
	ref, err = store.GetReference(ctx, "mpxn1")
	assert.NoError(err, "failed to retrieve booking ref")
	assert.Equal("ref2", ref, "mismatch")

	ref, err = store.GetReference(ctx, "mpxn2")
	assert.ErrorIs(err, ErrBookingReferenceNotFound)
}
