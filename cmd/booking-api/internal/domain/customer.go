package domain

import (
	"context"
	"errors"
	"fmt"

	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
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

type CustomerDomain struct {
	accounts       AccountGateway
	eligibilityGw  EligibilityGateway
	occupancyStore OccupancyStore
	siteStore      SiteStore
}

func NewCustomerDomain(accounts AccountGateway,
	eligibilityGw EligibilityGateway,
	occupancyStore OccupancyStore,
	siteStore SiteStore) CustomerDomain {
	return CustomerDomain{
		accounts,
		eligibilityGw,
		occupancyStore,
		siteStore,
	}
}

func (d CustomerDomain) GetCustomerContactDetails(ctx context.Context, accountID string) (models.Account, error) {

	account, err := d.accounts.GetAccountByAccountID(ctx, accountID)
	if err != nil {
		return models.Account{}, err
	}

	return account, nil
}

func (d CustomerDomain) GetAccountAddressByAccountID(ctx context.Context, accountID string) (models.AccountAddress, error) {

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
