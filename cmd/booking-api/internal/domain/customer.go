package domain

import (
	"context"
	"errors"
	"fmt"

	addressv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/address/v1"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"google.golang.org/genproto/googleapis/type/date"
)

var (
	ErrNoEligibleOccupanciesFound = errors.New("no eligible occupancies were found")
	ErrNoOccupanciesFound         = errors.New("no occupancies were found")
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
	GetSiteBySiteID(ctx context.Context, siteID string) (*models.Site, error)
}

type BookingStore interface {
	GetBookingsByAccountID(ctx context.Context, accountID string) ([]models.Booking, error)
}

type CustomerDomain struct {
	accounts       AccountGateway
	eligibilityGw  EligibilityGateway
	occupancyStore OccupancyStore
	siteStore      SiteStore
	bookingStore   BookingStore
}

func NewCustomerDomain(accounts AccountGateway,
	eligibilityGw EligibilityGateway,
	occupancyStore OccupancyStore,
	siteStore SiteStore,
	bookingStore BookingStore) *CustomerDomain {
	return &CustomerDomain{
		accounts,
		eligibilityGw,
		occupancyStore,
		siteStore,
		bookingStore,
	}
}

func (d *CustomerDomain) GetCustomerContactDetails(ctx context.Context, accountID string) (models.Account, error) {

	account, err := d.accounts.GetAccountByAccountID(ctx, accountID)
	if err != nil {
		return models.Account{}, err
	}

	return account, nil
}

func (d *CustomerDomain) GetAccountAddressByAccountID(ctx context.Context, accountID string) (models.AccountAddress, error) {

	var targetOccupancy models.Occupancy = models.Occupancy{}

	occupancies, err := d.occupancyStore.GetLiveOccupanciesByAccountID(ctx, accountID)
	if err != nil {
		return models.AccountAddress{}, fmt.Errorf("failed to get occupancies by account id, %w", err)
	}

	if len(occupancies) == 0 {
		return models.AccountAddress{}, ErrNoOccupanciesFound
	}

	for _, occupancy := range occupancies {
		isEligible, err := d.eligibilityGw.GetEligibility(ctx, accountID, occupancy.OccupancyID)
		if err != nil {
			return models.AccountAddress{}, fmt.Errorf("failed to get eligibility for accountId: %s, occupancyId: %s, %w", accountID, occupancy.OccupancyID, err)
		}

		if isEligible {
			targetOccupancy = occupancy
			break
		}
	}

	if targetOccupancy.IsEmpty() {
		return models.AccountAddress{}, ErrNoEligibleOccupanciesFound
	}

	site, err := d.siteStore.GetSiteBySiteID(ctx, targetOccupancy.SiteID)
	if err != nil {
		return models.AccountAddress{}, fmt.Errorf("failed to get site with site_id :%s, %w", targetOccupancy.SiteID, err)
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

func getUniqueSiteIDs(bookings []models.Booking) map[string]struct{} {
	idSet := make(map[string]struct{})
	for _, b := range bookings {
		idSet[b.SiteID] = struct{}{}
	}
	return idSet
}

func (d *CustomerDomain) getAddresses(ctx context.Context, siteIDs map[string]struct{}) (map[string]*addressv1.Address, error) {
	addresses := make(map[string]*addressv1.Address)
	for siteID := range siteIDs {
		sm, err := d.siteStore.GetSiteBySiteID(ctx, siteID)
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
		addresses[siteID] = address
	}
	return addresses, nil
}

func (d *CustomerDomain) GetCustomerBookings(ctx context.Context, accountID string) ([]*bookingv1.Booking, error) {
	bookingModels, err := d.bookingStore.GetBookingsByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	addresses, err := d.getAddresses(ctx, getUniqueSiteIDs(bookingModels))
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
			SiteAddress: addresses[bm.SiteID],
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
