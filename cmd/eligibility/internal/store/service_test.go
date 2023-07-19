package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-pkg/domain"
)

func TestService(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	s := NewService(connect(ctx))

	err := s.Add(ctx, &Service{
		ID:          "service1",
		Mpxn:        "mpxn1",
		OccupancyID: "occupancy1",
		SupplyType:  domain.SupplyTypeGas,
		IsLive:      false,
	})
	startDate := date(2020, 2, 1)
	assert.NoError(err, "failed to add service")

	err = s.AddStatDate(ctx, "service1", startDate)
	assert.NoError(err, "failed to add start date to service")

	service, err := s.Get(ctx, "service1")
	assert.NoError(err, "failed to get service")
	expected := Service{
		ID:          "service1",
		Mpxn:        "mpxn1",
		OccupancyID: "occupancy1",
		SupplyType:  domain.SupplyTypeGas,
		IsLive:      false,
		StartDate:   &startDate,
		EndDate:     nil,
	}
	assert.Equal(expected, service, "mismatch")

	liveServices, err := s.GetLiveServicesByOccupancyID(ctx, "occupancy1")
	assert.NoError(err, "failed to get live services")
	assert.Equal(0, len(liveServices), "mismatch: should have 0 live services")

	err = s.Add(ctx, &Service{
		ID:          "service1",
		Mpxn:        "mpxn1",
		OccupancyID: "occupancy1",
		SupplyType:  domain.SupplyTypeGas,
		IsLive:      true,
	})
	assert.NoError(err, "failed to update service")

	liveServices, err = s.GetLiveServicesByOccupancyID(ctx, "occupancy1")
	assert.NoError(err, "failed to get live services")
	assert.Equal(1, len(liveServices), "mismatch: should have 1 live service")
}
