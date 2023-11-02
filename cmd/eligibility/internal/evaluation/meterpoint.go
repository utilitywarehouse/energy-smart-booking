package evaluation

import (
	"context"

	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

type WanCoverageStore interface {
	GetWanCoverage(ctx context.Context, postcode string) (bool, error)
}

type AltHanStore interface {
	GetAltHan(ctx context.Context, mpxn string) (bool, error)
}

type MeterpointEvaluator struct {
	WanCoverageStore
	AltHanStore
}

func NewMeterpointEvaluator(w WanCoverageStore, a AltHanStore) *MeterpointEvaluator {
	return &MeterpointEvaluator{
		WanCoverageStore: w,
		AltHanStore:      a,
	}
}

func (e *MeterpointEvaluator) GetElectricityMeterpointEligibility(ctx context.Context, meters *models.ElectricityMeterTechnicalDetails, postcode string) (bool, error) {
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
