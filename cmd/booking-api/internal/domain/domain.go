package domain

import (
	"context"

	addressv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/address/v1"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
)

type AccountGateway interface {
	GetAccountByAccountID(ctx context.Context, accountID string) (models.Account, error)
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
	CreateBooking(ctx context.Context, postcode, reference string, slot models.BookingSlot, accountDetails models.AccountDetails, vulnerabilities []lowribeckv1.Vulnerability, other string) (gateway.CreateBookingResponse, error)
	GetAvailableSlotsPointOfSale(ctx context.Context, postcode, mpan, mprn string, tariffElectricity, tariffGas lowribeckv1.TariffType) (gateway.AvailableSlotsResponse, error)
	CreateBookingPointOfSale(ctx context.Context, postcode, mpan, mprn string, tariffElectricity, tariffGas lowribeckv1.TariffType, slot models.BookingSlot, accountDetails models.AccountDetails, vulnerabilities []lowribeckv1.Vulnerability, other string) (gateway.CreateBookingPointOfSaleResponse, error)
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

type BookingDomain struct {
	accounts                        AccountGateway
	lowribeckGw                     LowriBeckGateway
	occupancyStore                  OccupancyStore
	siteStore                       SiteStore
	bookingStore                    BookingStore
	partialBookingStore             PartialBookingStore
	pointOfSaleCustomerDetailsStore PointOfSaleCustomerDetailsStore
	eligibilityGw                   EligibilityGateway
	clickGw                         ClickGateway
	useTracing                      bool
}

func NewBookingDomain(accounts AccountGateway,
	lowribeckGw LowriBeckGateway,
	occupancyStore OccupancyStore,
	siteStore SiteStore,
	bookingStore BookingStore,
	partialBookingStore PartialBookingStore,
	pointOfSaleCustomerDetailsStore PointOfSaleCustomerDetailsStore,
	eligibilityGw EligibilityGateway,
	clickGw ClickGateway,
	useTracing bool,
) BookingDomain {
	return BookingDomain{
		accounts,
		lowribeckGw,
		occupancyStore,
		siteStore,
		bookingStore,
		partialBookingStore,
		pointOfSaleCustomerDetailsStore,
		eligibilityGw,
		clickGw,
		useTracing,
	}
}

func toAddress(address models.AccountAddress) *addressv1.Address {
	return &addressv1.Address{
		Uprn: address.UPRN,
		Paf: &addressv1.Address_PAF{
			Organisation:            address.PAF.Organisation,
			Department:              address.PAF.Department,
			SubBuilding:             address.PAF.SubBuilding,
			BuildingName:            address.PAF.BuildingName,
			BuildingNumber:          address.PAF.BuildingNumber,
			DependentThoroughfare:   address.PAF.DependentThoroughfare,
			Thoroughfare:            address.PAF.Thoroughfare,
			DoubleDependentLocality: address.PAF.DoubleDependentLocality,
			DependentLocality:       address.PAF.DependentLocality,
			PostTown:                address.PAF.PostTown,
			Postcode:                address.PAF.Postcode,
		},
	}
}
