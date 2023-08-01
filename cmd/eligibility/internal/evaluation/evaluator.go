package evaluation

import (
	"context"

	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/publisher"
)

type AccountStore interface {
	GetAccount(ctx context.Context, accountID string) (store.Account, error)
}

type MeterpointStore interface {
	Get(ctx context.Context, mpxn string) (store.Meterpoint, error)
}

type MeterStore interface {
	Get(ctx context.Context, mpxn string) (store.Meter, error)
}

type OccupancyStore interface {
	Get(ctx context.Context, id string) (store.Occupancy, error)
}

type PostcodeStore interface {
	GetWanCoverage(ctx context.Context, code string) (bool, error)
}

type SiteStore interface {
	Get(ctx context.Context, id string) (store.Site, error)
}

type ServiceStore interface {
	GetLiveServicesByOccupancyID(ctx context.Context, occupancyID string) ([]store.Service, error)
}

type BookingReferenceStore interface {
	GetReference(ctx context.Context, mpxn string) (string, error)
}

type CampaignabilityStore interface {
	Get(ctx context.Context, occupancyID, accountID string) (store.Campaignability, error)
}

type SuppliabilityStore interface {
	Get(ctx context.Context, occupancyID, accountID string) (store.Suppliability, error)
}

type EligibilityStore interface {
	Get(ctx context.Context, occupancyID, accountID string) (store.Eligibility, error)
}

type Evaluator struct {
	accountStore         AccountStore
	meterpointStore      MeterpointStore
	meterStore           MeterStore
	occupancyStore       OccupancyStore
	postcodeStore        PostcodeStore
	serviceStore         ServiceStore
	siteStore            SiteStore
	bookingRefStore      BookingReferenceStore
	campaignabilityStore CampaignabilityStore
	suppliabilityStore   SuppliabilityStore
	eligibilityStore     EligibilityStore
	eligibilitySync      publisher.SyncPublisher
	suppliabilitySync    publisher.SyncPublisher
	campaignabilitySync  publisher.SyncPublisher
}

func NewEvaluator(accountsStore AccountStore, meterpointsStore MeterpointStore,
	metersStore MeterStore, occupanciesStore OccupancyStore, postcodesStore PostcodeStore,
	serviceStore ServiceStore, siteStore SiteStore, bookingRefStore BookingReferenceStore,
	suppliabilityStore SuppliabilityStore, campaignabilityStore CampaignabilityStore, eligibilityStore EligibilityStore,
	eligibilitySync publisher.SyncPublisher, suppliabilitySync publisher.SyncPublisher, campaignabilitySync publisher.SyncPublisher) *Evaluator {
	return &Evaluator{
		accountStore:         accountsStore,
		meterpointStore:      meterpointsStore,
		meterStore:           metersStore,
		occupancyStore:       occupanciesStore,
		postcodeStore:        postcodesStore,
		serviceStore:         serviceStore,
		siteStore:            siteStore,
		bookingRefStore:      bookingRefStore,
		suppliabilityStore:   suppliabilityStore,
		eligibilityStore:     eligibilityStore,
		campaignabilityStore: campaignabilityStore,
		eligibilitySync:      eligibilitySync,
		suppliabilitySync:    suppliabilitySync,
		campaignabilitySync:  campaignabilitySync,
	}
}
