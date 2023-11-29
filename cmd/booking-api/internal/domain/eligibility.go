package domain

import (
	"context"
	"fmt"

	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

const posJourneyType = "point_of_sale"

type ProcessEligibilityParams struct {
	AccountNumber     string
	Postcode          string
	ElecOrderSupplies models.OrderSupply
	GasOrderSupplies  models.OrderSupply
}

type ProcessEligibilityResult struct {
	Eligible bool
}

func (d BookingDomain) ProcessEligibility(ctx context.Context, params ProcessEligibilityParams) (ProcessEligibilityResult, error) {

	eligible, err := d.eligibilityGw.GetMeterpointEligibility(ctx, params.ElecOrderSupplies.MPXN, params.GasOrderSupplies.MPXN, params.Postcode)
	if err != nil {
		return ProcessEligibilityResult{}, fmt.Errorf("failed to get meterpoint eligibility, %w", err)
	}

	return ProcessEligibilityResult{
		Eligible: eligible,
	}, nil
}

type GetClickLinkParams struct {
	AccountNumber string
	Details       models.PointOfSaleCustomerDetails
}

type GetClickLinkResult struct {
	Eligible bool
	Link     string
}

func (d BookingDomain) GetClickLink(ctx context.Context, params GetClickLinkParams) (GetClickLinkResult, error) {

	eligible, err := d.eligibilityGw.GetMeterpointEligibility(ctx, params.Details.ElecOrderSupplies.MPXN, params.Details.GasOrderSupplies.MPXN, params.Details.Address.PAF.Postcode)
	if err != nil {
		return GetClickLinkResult{}, fmt.Errorf("failed to get meterpoint eligibility for mpan/mprn: (%s/%s), %w", params.Details.ElecOrderSupplies.MPXN, params.Details.GasOrderSupplies.MPXN, err)
	}

	if !eligible {
		return GetClickLinkResult{
			Eligible: eligible,
			Link:     "",
		}, nil
	}

	err = d.pointOfSaleCustomerDetailsStore.SetAccountDetails(ctx, params.Details)
	if err != nil {
		return GetClickLinkResult{}, fmt.Errorf("failed to upsert customer details for account number: (%s), %w", params.AccountNumber, err)
	}

	attributes := map[string]string{
		"journey_type":   posJourneyType,
		"account_number": params.AccountNumber,
	}
	link, err := d.clickGw.GenerateAuthenticated(ctx, params.AccountNumber, attributes)
	if err != nil {
		return GetClickLinkResult{}, fmt.Errorf("failed to generate authenticated link for account number: %s, %w", params.AccountNumber, err)
	}

	return GetClickLinkResult{
		Eligible: eligible,
		Link:     link,
	}, nil
}
