package domain

import (
	"context"
	"fmt"

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

	elecMeterpoint, gasMeterpoint, err := models.DeduceOrderSupplies(params.Details.OrderSupplies)
	if err != nil {
		return ProcessEligibilityResult{}, fmt.Errorf("failed to deduce meterpoints, %w", err)
	}

	eligible, err := d.eligibilityGw.GetMeterpointEligibility(ctx, elecMeterpoint.MPXN, gasMeterpoint.MPXN, params.Details.Address.PAF.Postcode)
	if err != nil {
		return ProcessEligibilityResult{}, fmt.Errorf("failed to get meterpoint eligibility, %w", err)
	}

	if !eligible {
		return ProcessEligibilityResult{
			Eligible: eligible,
			Link:     "",
		}, nil
	}

	err = d.pointOfSaleCustomerDetailsStore.Upsert(ctx, params.AccountNumber, params.Details)
	if err != nil {
		return ProcessEligibilityResult{}, fmt.Errorf("failed to upsert customer details, %w", err)
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
