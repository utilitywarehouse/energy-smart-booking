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
	Link     string
}

func (d BookingDomain) ProcessEligibility(ctx context.Context, params ProcessEligibilityParams) (ProcessEligibilityResult, error) {

	err := d.pointOfSaleCustomerDetailsStore.Upsert(ctx, params.AccountNumber, params.Details)
	if err != nil {
		return ProcessEligibilityResult{}, fmt.Errorf("failed to upsert customer details, %w", err)
	}

	elecMeterpoint, gasMeterpoint, err := deduceMeterpoints(params.Details.Meterpoints)
	if err != nil {
		return ProcessEligibilityResult{}, fmt.Errorf("failed to deduce meterpoints, %w", err)
	}

	eligible, err := d.eligibilityGw.GetMeterpointEligibility(ctx, params.AccountNumber, elecMeterpoint.MPXN, gasMeterpoint.MPXN, params.Details.Address.PAF.Postcode)
	if err != nil {
		return ProcessEligibilityResult{}, fmt.Errorf("failed to get meterpoint eligibility, %w", err)
	}

	if !eligible {
		return ProcessEligibilityResult{
			Eligible: eligible,
			Link:     "",
		}, nil
	}

	link, err := d.clickGw.GenerateAuthenticated(ctx, params.AccountNumber)
	if err != nil {
		return ProcessEligibilityResult{}, fmt.Errorf("failed to generate authenticated link for account number: %s, %w", params.AccountNumber, err)
	}

	return ProcessEligibilityResult{
		Eligible: eligible,
		Link:     fmt.Sprintf("%s&journey_type=point_of_sale&account_number=%s", link, params.AccountNumber),
	}, nil
}

func deduceMeterpoints(meterpoints []models.Meterpoint) (models.Meterpoint, models.Meterpoint, error) {
	var mpan, mprn models.Meterpoint
	for i, meterpoint := range meterpoints {
		mpxn, err := energy.NewMeterPointNumber(meterpoint.MPXN)
		if err != nil {
			return models.Meterpoint{}, models.Meterpoint{}, fmt.Errorf("invalid meterpoint number (%s): %v", meterpoint.MPXN, err)
		}
		// We want the first electricity MPAN
		if mpxn.SupplyType() == energy.SupplyTypeElectricity && mpan.IsEmpty() {
			mpan = meterpoints[i]
		} else if mpxn.SupplyType() == energy.SupplyTypeGas {
			mprn = meterpoints[i]
		}
	}
	return mpan, mprn, nil
}
