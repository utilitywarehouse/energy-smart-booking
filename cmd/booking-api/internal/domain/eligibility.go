package domain

import (
	"context"
	"fmt"

	energy "github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

type ProcessEligibilityParams struct {
	AccountNumber string
	Details       models.PointOfSaleCustomerDetails
}

type ProcessEligibilityResult struct {
	Eligible bool
}

func (d BookingDomain) ProcessEligibility(ctx context.Context, params ProcessEligibilityParams) (ProcessEligibilityResult, error) {

	elecOrderSupply, gasOrderSupply, err := deduceOrderSupplies(params.Details.OrderSupplies)
	if err != nil {
		return ProcessEligibilityResult{}, fmt.Errorf("failed to deduce order supplies, %w", err)
	}

	eligible, err := d.eligibilityGw.GetMeterpointEligibility(ctx, elecOrderSupply.MPXN, gasOrderSupply.MPXN, params.Details.Address.PAF.Postcode)
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

	elecOrderSupply, gasOrderSupply, err := deduceOrderSupplies(params.Details.OrderSupplies)
	if err != nil {
		return GetClickLinkResult{}, fmt.Errorf("failed to deduce order supplies, %w", err)
	}

	eligible, err := d.eligibilityGw.GetMeterpointEligibility(ctx, elecOrderSupply.MPXN, gasOrderSupply.MPXN, params.Details.Address.PAF.Postcode)
	if err != nil {
		return GetClickLinkResult{}, fmt.Errorf("failed to get meterpoint eligibility for mpan/mprn: (%s/%s), %w", elecOrderSupply.MPXN, gasOrderSupply.MPXN, err)
	}

	if !eligible {
		return GetClickLinkResult{
			Eligible: eligible,
			Link:     "",
		}, nil
	}

	err = d.pointOfSaleCustomerDetailsStore.Upsert(ctx, params.AccountNumber, params.Details)
	if err != nil {
		return GetClickLinkResult{}, fmt.Errorf("failed to upsert customer details for account number: (%s), %w", params.AccountNumber, err)
	}

	link, err := d.clickGw.GenerateAuthenticated(ctx, params.AccountNumber)
	if err != nil {
		return GetClickLinkResult{}, fmt.Errorf("failed to generate authenticated link for account number: %s, %w", params.AccountNumber, err)
	}

	return GetClickLinkResult{
		Eligible: eligible,
		Link:     fmt.Sprintf("%s&journey_type=point_of_sale&account_number=%s", link, params.AccountNumber),
	}, nil
}

func deduceOrderSupplies(orderSupplies []models.OrderSupply) (models.OrderSupply, models.OrderSupply, error) {
	var mpan, mprn models.OrderSupply
	for i, orderSupply := range orderSupplies {
		mpxn, err := energy.NewMeterPointNumber(orderSupply.MPXN)
		if err != nil {
			return models.OrderSupply{}, models.OrderSupply{}, fmt.Errorf("invalid meterpoint number (%s): %v", orderSupply.MPXN, err)
		}
		// We want the first electricity MPAN
		if mpxn.SupplyType() == energy.SupplyTypeElectricity && mpan.IsEmpty() {
			mpan = orderSupplies[i]
		} else if mpxn.SupplyType() == energy.SupplyTypeGas {
			mprn = orderSupplies[i]
		}
	}
	return mpan, mprn, nil
}
