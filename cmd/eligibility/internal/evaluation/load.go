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
	dbServices, err := e.serviceStore.GetLiveServicesByOccupancyID(ctx, id)
	if err != nil {
		return nil, err
	}

	// load meterpoint
	for _, s := range dbServices {
		service := domain.Service{
			ID:         s.ID,
			Mpxn:       s.Mpxn,
			SupplyType: s.SupplyType,
		}

		// load meterpoint
		dbMeterpoint, err := e.meterpointStore.Get(ctx, s.Mpxn)
		if err != nil && !errors.Is(err, store.ErrMeterpointNotFound) {
			return nil, err
		}
		if err == nil {
			service.Meterpoint = &domain.Meterpoint{
				Mpxn:         dbMeterpoint.Mpxn,
				AltHan:       dbMeterpoint.AltHan,
				ProfileClass: dbMeterpoint.ProfileClass,
				SSC:          dbMeterpoint.SSC,
			}
		}

		dbMeter, err := e.meterStore.Get(ctx, s.Mpxn)
		if err != nil && !errors.Is(err, store.ErrMeterNotFound) {
			return nil, err
		}
		if err == nil {
			service.Meter = &domain.Meter{
				Mpxn:      dbMeter.Mpxn,
				MSN:       dbMeter.Msn,
				Capacity:  dbMeter.Capacity,
				MeterType: dbMeter.MeterType,
			}
		}

		ref, err := e.bookingRefStore.GetReference(ctx, s.Mpxn)
		if err != nil && !errors.Is(err, store.ErrBookingReferenceNotFound) {
			return nil, err
		}
		service.BookingReference = ref

		occupancy.Services = append(occupancy.Services, service)
	}

	return &occupancy, nil
}
