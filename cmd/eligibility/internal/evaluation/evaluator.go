package evaluation

import (
	"context"

	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/internal/publisher"
)

type OccupancyStore interface {
	LoadOccupancy(ctx context.Context, occupancyID string) (domain.Occupancy, error)
}

type ServiceStore interface {
	LoadLiveServicesByOccupancyID(ctx context.Context, occupancyID string) ([]domain.Service, error)
}

type Evaluator struct {
	occupancyStore      OccupancyStore
	serviceStore        ServiceStore
	eligibilitySync     publisher.SyncPublisher
	suppliabilitySync   publisher.SyncPublisher
	campaignabilitySync publisher.SyncPublisher
}

func NewEvaluator(occupanciesStore OccupancyStore, serviceStore ServiceStore,
	eligibilitySync publisher.SyncPublisher, suppliabilitySync publisher.SyncPublisher, campaignabilitySync publisher.SyncPublisher) *Evaluator {
	return &Evaluator{
		occupancyStore:      occupanciesStore,
		serviceStore:        serviceStore,
		eligibilitySync:     eligibilitySync,
		suppliabilitySync:   suppliabilitySync,
		campaignabilitySync: campaignabilitySync,
	}
}
