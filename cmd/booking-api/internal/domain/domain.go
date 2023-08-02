package domain

import (
	"context"

	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

type AccountGateway interface {
	GetAccountByAccountID(ctx context.Context, accountID string) (models.Account, error)
}

type EligibilityGateway interface {
	GetEligibility(ctx context.Context, accountID, occupancyID string) (bool, error)
}

type OccupancyStore interface {
	GetLiveOccupanciesByAccountID(ctx context.Context, accountID string) ([]models.Occupancy, error)
}

type SiteStore interface {
	GetSiteByOccupancyID(ctx context.Context, occupancyID string) (*models.Site, error)
}

type LowriBeckGateway interface {
	GetAvailableSlots(ctx context.Context, postcode, reference string) ([]models.BookingSlot, error)
	CreateBooking(ctx context.Context, postcode, reference string, slot models.BookingSlot, accountDetails models.AccountDetails, vulnerabilities []lowribeckv1.Vulnerability, other string) (bool, error)
}

type ServiceStore interface {
	GetReferenceByOccupancyID(ctx context.Context, occupancyID string) (string, error)
}

type BookingReferenceStore interface {
	GetReferenceByMPXN(ctx context.Context, mpxn string) (string, error)
}

type BookingStore interface {
	GetBookingByBookingID(ctx context.Context, bookingID string) (models.Booking, error)
	GetBookingsByAccountID(ctx context.Context, accountID string) ([]models.Booking, error)
}

type BookingDomain struct {
	accounts              AccountGateway
	eligibilityGw         EligibilityGateway
	lowribeckGw           LowriBeckGateway
	occupancyStore        OccupancyStore
	siteStore             SiteStore
	serviceStore          ServiceStore
	bookingReferenceStore BookingReferenceStore
	bookingStore          BookingStore
}

func NewBookingDomain(accounts AccountGateway,
	eligibilityGw EligibilityGateway,
	lowribeckGw LowriBeckGateway,
	occupancyStore OccupancyStore,
	siteStore SiteStore,
	serviceStore ServiceStore,
	bookingReferenceStore BookingReferenceStore,
	bookingStore BookingStore,
) BookingDomain {
	return BookingDomain{
		accounts,
		eligibilityGw,
		lowribeckGw,
		occupancyStore,
		siteStore,
		serviceStore,
		bookingReferenceStore,
		bookingStore,
	}
}
