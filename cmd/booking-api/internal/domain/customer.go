package domain

import (
	"context"
	"errors"
	"fmt"

	addressv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/address/v1"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"google.golang.org/genproto/googleapis/type/date"
)

var (
	ErrNoEligibleOccupanciesFound = errors.New("no eligible occupancies were found")
	ErrPOSCustomerDetailsNotFound = errors.New("point of sale customer details not found")
)

func (d BookingDomain) GetCustomerContactDetails(ctx context.Context, accountID string) (models.Account, error) {
	return d.getCustomerContactDetails(ctx, accountID)
}

func (d BookingDomain) GetAccountAddressByAccountID(ctx context.Context, accountID string) (models.AccountAddress, error) {

	site, _, err := d.occupancyStore.GetSiteExternalReferenceByAccountID(ctx, accountID)
	if err != nil {
		if errors.Is(err, store.ErrNoEligibleOccupancyFound) {
			return models.AccountAddress{}, ErrNoEligibleOccupanciesFound
		}
		return models.AccountAddress{}, fmt.Errorf("failed to get occupancies by account id, %w", err)
	}

	address := models.AccountAddress{
		UPRN: site.UPRN,
		PAF: models.PAF{
			BuildingName:            site.BuildingNameNumber,
			BuildingNumber:          site.BuildingNameNumber,
			SubBuilding:             site.SubBuildingNameNumber,
			Department:              site.Department,
			DependentLocality:       site.DependentLocality,
			DependentThoroughfare:   site.DependentThoroughfare,
			DoubleDependentLocality: site.DoubleDependentLocality,
			Organisation:            site.Organisation,
			Postcode:                site.Postcode,
			Thoroughfare:            site.Thoroughfare,
			PostTown:                site.Town,
		},
	}

	return address, nil
}

func (d BookingDomain) GetCustomerBookings(ctx context.Context, accountID string) ([]*bookingv1.Booking, error) {
	bookingModels, err := d.bookingStore.GetBookingsByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	addresses, err := d.getAddresses(ctx, getUniqueOccupancyIDs(bookingModels))
	if err != nil {
		return nil, err
	}

	contractBookings := make([]*bookingv1.Booking, 0, len(bookingModels))
	for _, bm := range bookingModels {
		y, m, d := bm.Slot.Date.Date()
		gdate := &date.Date{
			Year:  int32(y),
			Month: int32(m),
			Day:   int32(d),
		}
		contractBookings = append(contractBookings, &bookingv1.Booking{
			Id:          bm.BookingID,
			AccountId:   accountID,
			SiteAddress: addresses[bm.OccupancyID],
			ContactDetails: &bookingv1.ContactDetails{
				Title:     bm.Contact.Title,
				FirstName: bm.Contact.FirstName,
				LastName:  bm.Contact.LastName,
				Phone:     bm.Contact.Mobile,
				Email:     bm.Contact.Email,
			},
			Slot: &bookingv1.BookingSlot{
				Date:      gdate,
				StartTime: int32(bm.Slot.StartTime),
				EndTime:   int32(bm.Slot.EndTime),
			},
			VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
				Vulnerabilities: bm.VulnerabilityDetails.Vulnerabilities,
				Other:           bm.VulnerabilityDetails.Other,
			},
			Status: bm.Status,
		})
	}

	return contractBookings, nil
}

func (d BookingDomain) GetCustomerDetailsPointOfSale(ctx context.Context, accountNumber string) (*models.PointOfSaleCustomerDetails, error) {
	return d.getCustomerDetailsPointOfSale(ctx, accountNumber)
}

func getUniqueOccupancyIDs(bookings []models.Booking) map[string]struct{} {
	idSet := make(map[string]struct{})
	for _, b := range bookings {
		idSet[b.OccupancyID] = struct{}{}
	}
	return idSet
}

func (d BookingDomain) getAddresses(ctx context.Context, occupancyIDs map[string]struct{}) (map[string]*addressv1.Address, error) {
	addresses := make(map[string]*addressv1.Address)
	for occID := range occupancyIDs {
		sm, err := d.siteStore.GetSiteByOccupancyID(ctx, occID)
		if err != nil {
			return nil, err
		}
		address := &addressv1.Address{
			Uprn: sm.UPRN,
			Paf: &addressv1.Address_PAF{
				Organisation:            sm.Organisation,
				Department:              sm.Department,
				SubBuilding:             sm.SubBuildingNameNumber,
				BuildingName:            sm.BuildingNameNumber,
				BuildingNumber:          sm.BuildingNameNumber,
				DependentThoroughfare:   sm.DependentThoroughfare,
				Thoroughfare:            sm.Thoroughfare,
				DoubleDependentLocality: sm.DoubleDependentLocality,
				DependentLocality:       sm.DependentLocality,
				PostTown:                sm.Town,
				Postcode:                sm.Postcode,
			},
		}
		addresses[occID] = address
	}
	return addresses, nil
}

func (d BookingDomain) getCustomerDetailsPointOfSale(ctx context.Context, accountNumber string) (*models.PointOfSaleCustomerDetails, error) {
	customerDetails, err := d.pointOfSaleCustomerDetailsStore.GetByAccountNumber(ctx, accountNumber)
	if err != nil {
		switch err {
		case store.ErrPOSCustomerDetailsNotFound:
			return nil, fmt.Errorf("%w, %w", ErrPOSCustomerDetailsNotFound, err)
		}

		return nil, fmt.Errorf("failed to get customer details, %w", err)
	}

	return customerDetails, nil
}

func (d BookingDomain) getCustomerContactDetails(ctx context.Context, accountID string) (models.Account, error) {
	account, err := d.accounts.GetAccountByAccountID(ctx, accountID)
	if err != nil {
		return models.Account{}, err
	}

	return account, nil
}
