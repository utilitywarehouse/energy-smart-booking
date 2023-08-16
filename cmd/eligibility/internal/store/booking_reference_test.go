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

	const mpxn = "booking_mpxn"

	err := store.Add(ctx, mpxn, "ref1")
	assert.NoError(err, "failed to add booking ref")

	ref, err := store.GetReference(ctx, mpxn)
	assert.NoError(err, "failed to retrieve booking ref")
	assert.Equal("ref1", ref, "mismatch")

	err = store.Add(ctx, mpxn, "ref2")
	assert.NoError(err)

	ref, err = store.GetReference(ctx, mpxn)
	assert.NoError(err, "failed to retrieve booking ref")
	assert.Equal("ref2", ref, "mismatch")

	err = store.Remove(ctx, mpxn)
	assert.NoError(err, "failed to remove booking ref")

	_, err = store.GetReference(ctx, mpxn)
	assert.ErrorIs(err, ErrBookingReferenceNotFound)
}
