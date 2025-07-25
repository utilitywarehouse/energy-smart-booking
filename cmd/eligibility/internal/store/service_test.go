package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	energy_domain "github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
)

func TestService(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	s := NewService(connect(ctx))

	const mpxn = "service_mpxn"
	err := s.Add(ctx, &Service{
		ID:          "service1",
		Mpxn:        mpxn,
		OccupancyID: "service-occupancy-1",
		SupplyType:  energy_domain.SupplyTypeGas,
		IsLive:      false,
	})
	startDate := date(2020, 2, 1)
	assert.NoError(err, "failed to add service")

	err = s.AddStartDate(ctx, "service1", startDate)
	assert.NoError(err, "failed to add start date to service")

	service, err := s.Get(ctx, "service1")
	assert.NoError(err, "failed to get service")
	expected := Service{
		ID:          "service1",
		Mpxn:        mpxn,
		OccupancyID: "service-occupancy-1",
		SupplyType:  energy_domain.SupplyTypeGas,
		IsLive:      false,
		StartDate:   &startDate,
		EndDate:     nil,
	}
	assert.Equal(expected, service, "mismatch")

	liveServices, err := s.LoadLiveServicesByOccupancyID(ctx, "service-occupancy-1")
	assert.NoError(err, "failed to get live services")
	assert.Equal(0, len(liveServices), "mismatch: should have 0 live services")

	err = s.Add(ctx, &Service{
		ID:          "service1",
		Mpxn:        mpxn,
		OccupancyID: "service-occupancy-1",
		SupplyType:  energy_domain.SupplyTypeGas,
		IsLive:      true,
	})
	assert.NoError(err, "failed to update service")

	liveServices, err = s.LoadLiveServicesByOccupancyID(ctx, "service-occupancy-1")
	assert.NoError(err, "failed to get live services")
	assert.Equal(1, len(liveServices), "mismatch: should have 1 live service")
	expectedSv := domain.Service{
		ID:         "service1",
		Mpxn:       mpxn,
		SupplyType: energy_domain.SupplyTypeGas,
	}
	assert.Equal(expectedSv, liveServices[0], "mismatch")

	_, err = s.pool.Exec(ctx, `INSERT INTO meterpoints(mpxn, supply_type, alt_han, profile_class, ssc) VALUES('service_mpxn', 'gas', true, 'PROFILE_CLASS_01', 'ssc');`)
	assert.NoError(err)

	liveServices, err = s.LoadLiveServicesByOccupancyID(ctx, "service-occupancy-1")
	assert.NoError(err, "failed to get live services")
	assert.Equal(1, len(liveServices), "mismatch: should have 1 live service")

	expectedSv.Meterpoint = &domain.Meterpoint{
		Mpxn:         mpxn,
		AltHan:       true,
		ProfileClass: platform.ProfileClass(1),
		SSC:          "ssc",
	}
	assert.Equal(expectedSv, liveServices[0], "mismatch")
}

func TestGetServicesWithBookingRef(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	s := NewService(connect(ctx))
	tmstp := time.Date(2023, time.January, 12, 10, 0, 0, 0, time.UTC)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO services(id, mpxn, occupancy_id, supply_type, is_live)
		VALUES 
		    ('service_id_1', 'mpxn_1', 'occupancy_id_1', 'gas', true),
		    ('service_id_2', 'mpxn_2', 'occupancy_id_1', 'elec', true);
		INSERT INTO booking_references(mpxn, reference) VALUES
		    ('mpxn_1', 'ref');`)
	assert.NoError(err)

	_, err = s.pool.Exec(ctx, `
		INSERT INTO booking_references(mpxn, reference, deleted_at) VALUES ('mpxn_2', 'ref2', $1);`, tmstp)
	assert.NoError(err)

	serviceBookingRef, err := s.GetLiveServicesWithBookingRef(ctx, "occupancy_id_1")
	assert.NoError(err)
	expected := []ServiceBookingRef{
		{
			ServiceID:  "service_id_1",
			BookingRef: "ref",
			DeletedAt:  nil,
		},
		{
			ServiceID:  "service_id_2",
			BookingRef: "ref2",
			DeletedAt:  &tmstp,
		},
	}
	assert.Equal(expected, serviceBookingRef)
}
