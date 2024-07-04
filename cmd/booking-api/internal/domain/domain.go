package domain

import (
	"context"

	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
)

type AccountGateway interface {
	GetAccountByAccountID(ctx context.Context, accountID string) (models.Account, error)
}

type AccountNumberGateway interface {
	Get(ctx context.Context, accountID string) (string, error)
}

type OccupancyStore interface {
	GetSiteExternalReferenceByAccountID(ctx context.Context, accountID string) (*models.Site, *models.OccupancyEligibility, error)
	GetOccupancyByAccountID(context.Context, string) (*models.Occupancy, error)
}

type SiteStore interface {
	GetSiteByOccupancyID(ctx context.Context, occupancyID string) (*models.Site, error)
}

type LowriBeckGateway interface {
	GetAvailableSlots(ctx context.Context, postcode, reference string) (gateway.AvailableSlotsResponse, error)
	CreateBooking(ctx context.Context, postcode, reference string, slot models.BookingSlot, contactDetails models.AccountDetails, vulnerabilities []lowribeckv1.Vulnerability, other string) (gateway.CreateBookingResponse, error)
	GetAvailableSlotsPointOfSale(ctx context.Context, postcode, mpan, mprn string, tariffElectricity, tariffGas lowribeckv1.TariffType) (gateway.AvailableSlotsResponse, error)
	CreateBookingPointOfSale(ctx context.Context, mpan, mprn string, tariffElectricity, tariffGas lowribeckv1.TariffType, slot models.BookingSlot, contactDetails models.AccountDetails, vulnerabilities []lowribeckv1.Vulnerability, other string, siteAddress models.AccountAddress) (gateway.CreateBookingPointOfSaleResponse, error)
}

type EligibilityGateway interface {
	GetMeterpointEligibility(ctx context.Context, mpan, mprn, postcode string) (bool, error)
}

type ClickGateway interface {
	GenerateAuthenticated(ctx context.Context, accountNo string, attributes map[string]string) (string, error)
}

type BookingStore interface {
	GetBookingByBookingID(ctx context.Context, bookingID string) (models.Booking, error)
	GetBookingsByAccountID(ctx context.Context, accountID string) ([]models.Booking, error)
}

type PartialBookingStore interface {
	Upsert(ctx context.Context, bookingID string, event *bookingv1.BookingCreatedEvent) error
}

type PointOfSaleCustomerDetailsStore interface {
	GetByAccountNumber(context.Context, string) (*models.PointOfSaleCustomerDetails, error)
	Upsert(context.Context, string, models.PointOfSaleCustomerDetails) error
}

type SmartMeterInterestStore interface {
	Insert(ctx context.Context, smartMeterInterest models.SmartMeterInterest) error
}
