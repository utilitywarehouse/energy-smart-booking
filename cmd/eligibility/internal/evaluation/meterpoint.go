package evaluation

import (
	"context"
	"errors"

	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

var (
	ErrThirdPartyMeterpointError = errors.New("could not retrieve meterpoint details")
)

type WanCoverageStore interface {
	GetWanCoverage(ctx context.Context, postcode string) (bool, error)
}

type AltHanStore interface {
	GetAltHan(ctx context.Context, mpxn string) (bool, error)
}

type EcoesAPI interface {
	GetMPANTechnicalDetails(ctx context.Context, mpan string) (*models.ElectricityMeterTechnicalDetails, error)
}

type XoserveAPI interface {
	GetMPRNTechnicalDetails(ctx context.Context, mprn string) (*models.GasMeterTechnicalDetails, error)
}

type MeterpointEvaluator struct {
	WanCoverageStore
	AltHanStore
	ecoesAPI   EcoesAPI
	xoserveAPI XoserveAPI
}

func NewMeterpointEvaluator(w WanCoverageStore, a AltHanStore, ecoesAPI EcoesAPI, xoserveAPI XoserveAPI) *MeterpointEvaluator {
	return &MeterpointEvaluator{
		WanCoverageStore: w,
		AltHanStore:      a,
		ecoesAPI:         ecoesAPI,
		xoserveAPI:       xoserveAPI,
	}
}

func (e *MeterpointEvaluator) GetElectricityMeterpointEligibility(ctx context.Context, mpan string, postcode string) (bool, error) {
	// None of the meters at the meters points are smart (SMETS1 or SMETS2)
	// to use exactly the same logic as in https://github.com/utilitywarehouse/energy-smart-booking/blob/master/cmd/eligibility/internal/domain/entities.go#L388

	meters, err := e.ecoesAPI.GetMPANTechnicalDetails(ctx, mpan)
	if err != nil {
		return false, ErrThirdPartyMeterpointError
	}

	for _, meter := range meters.Meters {
		if domain.IsElectricitySmartMeter(meter.MeterType.String()) {
			return false, nil
		}
	}

	// Property has WAN
	isWan, err := e.GetWanCoverage(ctx, postcode)
	if err != nil {
		return false, err
	}
	if !isWan {
		return false, nil
	}

	// Property does not require ALT-HAN
	isAltHan, err := e.GetAltHan(ctx, postcode)
	if err != nil {
		return false, err
	}
	if isAltHan {
		return false, err
	}

	// Electricity is not a related MPAN Set-up
	// We should not receive a related MPAN from a GetRelatedMPANs call

	// Electricity does not have “complex tariff”
	// Similar to the current logic in the normal eligibilty check for "complex tariff"
	// same logic as https://github.com/utilitywarehouse/energy-smart-booking/blob/master/cmd/eligibility/internal/domain/entities.go#L377

	// Gas meter at property is not “large capacity”
	// Large Capacity means the meter's capacity is different than 6 or 212
	// same logic as https://github.com/utilitywarehouse/energy-smart-booking/blob/master/cmd/eligibility/internal/evaluation/rules.go#L55

	return true, nil
}
