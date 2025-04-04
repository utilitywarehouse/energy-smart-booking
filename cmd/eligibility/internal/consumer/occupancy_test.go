package consumer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/consumer"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/uw-labs/substrate"
)

func TestOccupancyConsumer(t *testing.T) {
	s := store.NewOccupancy(pool)
	defer truncateDB(t)

	handler := consumer.HandleOccupancy(s, nil, true)

	occEv1, err := makeMessage(&platform.OccupancyStartedEvent{
		OccupancyId:       "occupancyID",
		SiteId:            "siteID",
		CustomerAccountId: "customerAccID",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{occEv1})
	assert.NoError(t, err, "failed to handle occupancy discovered event")

	occ, err := s.Get(t.Context(), "occupancyID")
	assert.NoError(t, err, "failed to get occupancy")
	expected := store.Occupancy{
		ID:        "occupancyID",
		SiteID:    "siteID",
		AccountID: "customerAccID",
	}
	assert.Equal(t, expected, occ, "occupancy mismatch")

	occEv2, err := makeMessage(&platform.OccupancySiteCorrectedEvent{
		OccupancyId: "occupancyID",
		SiteId:      "siteID1",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{occEv2})
	assert.NoError(t, err, "failed to handle occupancy site corrected event")

	occ, err = s.Get(t.Context(), "occupancyID")
	assert.NoError(t, err, "failed to get occupancy")
	expected.SiteID = "siteID1"
	assert.Equal(t, expected, occ, "occupancy mismatch")
}
