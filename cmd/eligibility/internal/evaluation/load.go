package evaluation

import (
	"context"
	"errors"

	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
)

func (e *Evaluator) LoadOccupancy(ctx context.Context, id string) (*domain.Occupancy, error) {
	dbOccupancy, err := e.occupancyStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	occupancy := domain.Occupancy{
		ID: id,
	}

	// load services for given occupancy
	dbServices, err := e.serviceStore.GetLiveServicesByOccupancyID(ctx, id)
	if err != nil {
		return nil, err
	}
	// load meterpoint
	for _, s := range dbServices {
		// load meterpoint
		service := domain.Service{
			ID: s.ID,
			Mpxn: s.Mpxn,
		}

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
				Mpxn:       dbMeter.Mpxn,
				MSN:        dbMeter.Msn,
				SupplyType: dbMeter.SupplyType,
				Capacity:   dbMeter.Capacity,
				MeterType:  dbMeter.MeterType,
			}
		}

		ref, err := e.bookingRefStore.GetReference(ctx, s.Mpxn)
		if err != nil && !errors.Is(err, store.ErrBookingReferenceNotFound) {
			return nil, err
		}
		service.BookingReference = ref

		occupancy.Services = append(occupancy.Services, service)
	}

	dbAccount, err := e.accountStore.GetAccount(ctx, dbOccupancy.AccountID)
	if err != nil && !errors.Is(err, store.ErrAccountNotFound) {
		return nil, err
	}
	occupancy.Account = domain.Account{
		ID:       dbOccupancy.AccountID,
		OptOut:   dbAccount.OptOut,
		PSRCodes: dbAccount.PSRCodes,
	}

	dbSite, err := e.siteStore.Get(ctx, dbOccupancy.SiteID)
	if err != nil && !errors.Is(err, store.ErrSiteNotFound) {
		return nil, err
	}
	if err == nil {
		covered, err := e.postcodeStore.GetWanCoverage(ctx, dbSite.PostCode)
		if err != nil && !errors.Is(err, store.ErrPostCodeNotFound) {
			return nil, err
		}
		occupancy.Site = &domain.Site{
			ID:          dbSite.ID,
			Postcode:    dbSite.PostCode,
			WanCoverage: covered,
		}
	}

	return &occupancy, nil
}
