package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWanCoverage(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	s := NewPostCode(connect(ctx))
	defer s.pool.Close()

	err := s.AddWanCoverage(ctx, "postCode1", true)
	assert.NoError(err, "failed to add wan coverage for post code")

	covered, err := s.GetWanCoverage(ctx, "postCode1")
	assert.NoError(err, "failed to get wan coverage for post code")
	assert.True(covered)

	err = s.AddWanCoverage(ctx, "postCode1", false)
	assert.NoError(err, "failed to update wan coverage for post code")

	covered, err = s.GetWanCoverage(ctx, "postCode1")
	assert.NoError(err, "failed to get wan coverage for post code")
	assert.False(covered)

	_, err = s.GetWanCoverage(ctx, "postCode2")
	assert.ErrorIs(err, ErrPostCodeNotFound)
}
