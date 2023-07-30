package evaluation

import (
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
)

func (e *Evaluator) LoadOccupancy(ctx context.Context, id string) (*domain.Occupancy, error) {
	now := time.Now()
	occupancy, err := e.occupancyStore.LoadOccupancy(ctx, id)
	if err != nil {
		return nil, err
	}
	log.WithField("elapsed_load_occupancy", time.Since(now)).Info("time to load occ")

	// load services for given occupancy
	now = time.Now()
	services, err := e.serviceStore.LoadLiveServicesByOccupancyID(ctx, id)
	if err != nil {
		return nil, err
	}
	log.WithField("elapsed_load_services", time.Since(now)).Info("time to load services")

	// load meterpoint
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
