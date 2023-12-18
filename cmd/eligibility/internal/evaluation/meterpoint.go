package evaluation

import (
	"context"
	"errors"
	"fmt"

	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

var (
	ErrThirdPartyMeterpointError = errors.New("could not retrieve meterpoint details")
)

type MeterpointIneligibleReason string

const (
	noneMeterpointIneligibleReason                  MeterpointIneligibleReason = ""
	alreadyMeterpointIneligibleReason               MeterpointIneligibleReason = "already_a_smart_meter"
	notWanMeterpointIneligibleReason                MeterpointIneligibleReason = "not_WAN"
	isAltHanlreadyMeterpointIneligibleReason        MeterpointIneligibleReason = "Alt_HAN"
	hasRelatedMeterpointsMeterpointIneligibleReason MeterpointIneligibleReason = "related_meterpoints_present"
	hasComplexSSCMeterpointIneligibleReason         MeterpointIneligibleReason = "complex_SSC"
	notLargeCapacityMeterpointIneligibleReason      MeterpointIneligibleReason = "not_large_capacity"
)

type WanCoverageStore interface {
	GetWanCoverage(ctx context.Context, postcode string) (bool, error)
}

type AltHanStore interface {
	GetAltHan(ctx context.Context, mpxn string) (bool, error)
}

type EcoesAPI interface {
	GetMPANTechnicalDetails(ctx context.Context, mpan string) (*models.ElectricityMeterTechnicalDetails, error)
	HasRelatedMPAN(ctx context.Context, mpan string) (bool, error)
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

func (e *MeterpointEvaluator) GetElectricityMeterpointEligibility(ctx context.Context, mpan string, postcode string) (bool, MeterpointIneligibleReason, error) {
	// None of the meters at the meters points can be smart (SMETS1 or SMETS2)
	// to use exactly the same logic as in https://github.com/utilitywarehouse/energy-smart-booking/blob/master/cmd/eligibility/internal/domain/entities.go#L388

	meters, err := e.ecoesAPI.GetMPANTechnicalDetails(ctx, mpan)
	if err != nil {
		return false, noneMeterpointIneligibleReason, fmt.Errorf("%w: %w", ErrThirdPartyMeterpointError, err)
	}

	for _, meter := range meters.Meters {
		if domain.IsElectricitySmartMeter(meter.MeterType.String()) {
			return false, alreadyMeterpointIneligibleReason, nil
		}
	}

	// Property must have WAN
	isWan, err := e.GetWanCoverage(ctx, postcode)
	if err != nil {
		return false, noneMeterpointIneligibleReason, err
	}
	if !isWan {
		return false, notWanMeterpointIneligibleReason, nil
	}

	// Property must not require ALT-HAN
	isAltHan, err := e.GetAltHan(ctx, mpan)
	if err != nil {
		return false, noneMeterpointIneligibleReason, err
	}
	if isAltHan {
		return false, isAltHanlreadyMeterpointIneligibleReason, err
	}

	// Electricity must not have a related MPAN Set-up
	// We should not receive a related MPAN from a GetRelatedMPANs call
	HasRelatedMPAN, err := e.ecoesAPI.HasRelatedMPAN(ctx, mpan)
	if err != nil {
		return false, noneMeterpointIneligibleReason, err
	}
	if HasRelatedMPAN {
		return false, hasRelatedMeterpointsMeterpointIneligibleReason, nil
	}

	// Electricity must not have “complex tariff”
	// Similar to the current logic in the normal eligibilty check for "complex tariff"
	if domain.HasComplexSSC(meters) {
		return false, hasComplexSSCMeterpointIneligibleReason, nil
	}

	return true, noneMeterpointIneligibleReason, nil
}

func (e *MeterpointEvaluator) GetGasMeterpointEligibility(ctx context.Context, mprn string) (bool, MeterpointIneligibleReason, error) {
	meters, err := e.xoserveAPI.GetMPRNTechnicalDetails(ctx, mprn)
	if err != nil {
		return false, noneMeterpointIneligibleReason, fmt.Errorf("%w: %w", ErrThirdPartyMeterpointError, err)
	}

	// Gas meter at property must not be “large capacity”
	// Large Capacity means the meter's capacity is different than 6 or 212
	if domain.IsLargeCapacity(meters) {
		return false, notLargeCapacityMeterpointIneligibleReason, nil
	}

	return true, noneMeterpointIneligibleReason, nil
}
