package evaluation

import (
	"context"
	"time"

	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"google.golang.org/protobuf/proto"
)

type OccupancyStore interface {
	LoadOccupancy(ctx context.Context, occupancyID string) (domain.Occupancy, error)
}

type ServiceStore interface {
	LoadLiveServicesByOccupancyID(ctx context.Context, occupancyID string) ([]domain.Service, error)
	GetLiveServicesWithBookingRef(ctx context.Context, occupancyID string) ([]store.ServiceBookingRef, error)
}

type MeterStore interface {
	Get(ctx context.Context, mpxn string) (domain.Meter, error)
}

type SyncPublisher interface {
	Sink(ctx context.Context, payload proto.Message, occurredAt time.Time) error
}

type Evaluator struct {
	occupancyStore         OccupancyStore
	serviceStore           ServiceStore
	meterStore             MeterStore
	eligibilitySync        SyncPublisher
	suppliabilitySync      SyncPublisher
	campaignabilitySync    SyncPublisher
	bookingEligibilitySync SyncPublisher
}

func NewEvaluator(occupanciesStore OccupancyStore, serviceStore ServiceStore, meterStore MeterStore,
	eligibilitySync SyncPublisher, suppliabilitySync SyncPublisher, campaignabilitySync SyncPublisher,
	bookingEligibilitySync SyncPublisher) *Evaluator {
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
