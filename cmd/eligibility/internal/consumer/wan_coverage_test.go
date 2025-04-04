package consumer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	smart "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/consumer"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/uw-labs/substrate"
)

func TestWanCoverageConsumer(t *testing.T) {
	s := store.NewPostCode(pool)
	defer truncateDB(t)

	handler := consumer.HandleWanCoverage(s, nil, nil, true)

	wanCoverageEv1, err := makeMessage(&smart.WanCoverageAtPostcodeStartedEvent{
		Postcode: "postCode",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{wanCoverageEv1})
	assert.NoError(t, err, "failed to handle wan coverage event")

	covered, err := s.GetWanCoverage(t.Context(), "postCode")
	assert.NoError(t, err, "failed to get wan coverage for post code")
	assert.True(t, covered)

	wanCoverageEv2, err := makeMessage(&smart.WanCoverageAtPostcodeEndedEvent{
		Postcode: "postCode",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{wanCoverageEv2})
	assert.NoError(t, err, "failed to handle wan coverage removed event")

	covered, err = s.GetWanCoverage(t.Context(), "postCode")
	assert.NoError(t, err, "failed to get wan coverage for post code")
	assert.False(t, covered)
}
