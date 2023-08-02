package evaluation

import (
	"context"
	"errors"

	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
)

func (e *Evaluator) LoadOccupancy(ctx context.Context, id string) (*domain.Occupancy, error) {
	occupancy, err := e.occupancyStore.LoadOccupancy(ctx, id)
	if err != nil {
		return nil, err
	}

	// load services for given occupancy
	services, err := e.serviceStore.LoadLiveServicesByOccupancyID(ctx, id)
	if err != nil {
		return nil, err
	}

	for i, s := range services {
		dbMeter, err := e.meterStore.Get(ctx, s.Mpxn)
		if err != nil && !errors.Is(err, store.ErrMeterNotFound) {
			return nil, err
		}
		if err == nil {
			services[i].Meter = &domain.Meter{
				Mpxn:       dbMeter.Mpxn,
				MSN:        dbMeter.Msn,
				SupplyType: dbMeter.SupplyType,
				Capacity:   dbMeter.Capacity,
				MeterType:  dbMeter.MeterType,
			}
		}

		occupancy.Services = append(occupancy.Services, services[i])
	}

	return &occupancy, nil
}
