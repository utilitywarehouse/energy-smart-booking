package domain

import (
	"context"

	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
)

type AccountGateway interface {
	GetAccountByAccountID(ctx context.Context, accountID string) (models.Account, error)
}

type EligibilityGateway interface {
	GetEligibility(ctx context.Context, accountID, occupancyID string) (bool, error)
}

type OccupancyStore interface {
	GetLiveOccupanciesByAccountID(ctx context.Context, accountID string) ([]models.Occupancy, error)
	GetSiteExternalReferenceByAccountID(ctx context.Context, accountID string) (*models.Site, *models.OccupancyEligibility, error)
}

type SiteStore interface {
	GetSiteByOccupancyID(ctx context.Context, occupancyID string) (*models.Site, error)
	GetSiteBySiteID(ctx context.Context, siteID string) (*models.Site, error)
}

type LowriBeckGateway interface {
	GetAvailableSlots(ctx context.Context, postcode, reference string) (gateway.AvailableSlotsResponse, error)
	CreateBooking(ctx context.Context, postcode, reference string, slot models.BookingSlot, accountDetails models.AccountDetails, vulnerabilities []lowribeckv1.Vulnerability, other string) (gateway.CreateBookingResponse, error)
}

type BookingStore interface {
	GetBookingByBookingID(ctx context.Context, bookingID string) (models.Booking, error)
	GetBookingsByAccountID(ctx context.Context, accountID string) ([]models.Booking, error)
}

type BookingDomain struct {
	accounts       AccountGateway
	eligibilityGw  EligibilityGateway
	lowribeckGw    LowriBeckGateway
	occupancyStore OccupancyStore
	siteStore      SiteStore
	bookingStore   BookingStore
}

func NewBookingDomain(accounts AccountGateway,
	eligibilityGw EligibilityGateway,
	lowribeckGw LowriBeckGateway,
	occupancyStore OccupancyStore,
	siteStore SiteStore,
	bookingStore BookingStore,
) BookingDomain {
	return BookingDomain{
		accounts,
		eligibilityGw,
		lowribeckGw,
		occupancyStore,
		siteStore,
		bookingStore,
	}
}
