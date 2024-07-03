package domain

import (
	addressv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/address/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

type BookingDomain struct {
	accounts                        AccountGateway
	accountNumber                   AccountNumberGateway
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
	accountNumberGateway AccountNumberGateway,
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
		accountNumberGateway,
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
