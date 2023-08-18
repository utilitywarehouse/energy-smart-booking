package evaluation

import (
	"context"

	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/publisher"
)

type OccupancyStore interface {
	LoadOccupancy(ctx context.Context, occupancyID string) (domain.Occupancy, error)
}

type ServiceStore interface {
	LoadLiveServicesByOccupancyID(ctx context.Context, occupancyID string) ([]domain.Service, error)
}

type MeterStore interface {
	Get(ctx context.Context, mpxn string) (store.Meter, error)
}

type Evaluator struct {
	occupancyStore         OccupancyStore
	serviceStore           ServiceStore
	meterStore             MeterStore
	eligibilitySync        publisher.SyncPublisher
	suppliabilitySync      publisher.SyncPublisher
	campaignabilitySync    publisher.SyncPublisher
	bookingEligibilitySync publisher.SyncPublisher
}

func NewEvaluator(occupanciesStore OccupancyStore, serviceStore ServiceStore, meterStore MeterStore,
	eligibilitySync publisher.SyncPublisher, suppliabilitySync publisher.SyncPublisher, campaignabilitySync publisher.SyncPublisher,
	bookingEligibilitySync publisher.SyncPublisher) *Evaluator {
	return &Evaluator{
		occupancyStore:         occupanciesStore,
		serviceStore:           serviceStore,
		meterStore:             meterStore,
		eligibilitySync:        eligibilitySync,
		suppliabilitySync:      suppliabilitySync,
		campaignabilitySync:    campaignabilitySync,
		bookingEligibilitySync: bookingEligibilitySync,
	}
}
