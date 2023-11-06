package evaluation

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
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
	GetRelatedMPAN(ctx context.Context, mpan string) (*models.ElectricityMeterRelatedMPAN, error)
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
	// None of the meters at the meters points can be smart (SMETS1 or SMETS2)
	// to use exactly the same logic as in https://github.com/utilitywarehouse/energy-smart-booking/blob/master/cmd/eligibility/internal/domain/entities.go#L388

	meters, err := e.ecoesAPI.GetMPANTechnicalDetails(ctx, mpan)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrThirdPartyMeterpointError, err)
	}

	for _, meter := range meters.Meters {
		if domain.IsElectricitySmartMeter(meter.MeterType.String()) {
			logrus.WithField("mpan", mpan).Info("ineligible point-of-sale booking: is already a smart meter")
			return false, nil
		}
	}

	// Property must have WAN
	isWan, err := e.GetWanCoverage(ctx, postcode)
	if err != nil {
		return false, err
	}
	if !isWan {
		logrus.WithField("mpan", mpan).Info("ineligible point-of-sale booking: is not WAN")
		return false, nil
	}

	// Property must not require ALT-HAN
	isAltHan, err := e.GetAltHan(ctx, mpan)
	if err != nil {
		return false, err
	}
	if isAltHan {
		logrus.WithField("mpan", mpan).Info("ineligible point-of-sale booking: is Alt-HAN")
		return false, err
	}

	// Electricity must not have a related MPAN Set-up
	// We should not receive a related MPAN from a GetRelatedMPANs call
	related, err := e.ecoesAPI.GetRelatedMPAN(ctx, mpan)
	if err != nil {
		return false, err
	}
	if related != nil && len(related.Relations) != 0 {
		logrus.WithField("mpan", mpan).Info("ineligible point-of-sale booking: related meterpoints present")
		return false, nil
	}

	// Electricity must not have “complex tariff”
	// Similar to the current logic in the normal eligibilty check for "complex tariff"
	if domain.HasComplexSSC(meters) {
		logrus.WithField("mpan", mpan).Info("ineligible point-of-sale booking: has complex SSC")
		return false, nil
	}

	return true, nil
}

func (e *MeterpointEvaluator) GetGasMeterpointEligibility(ctx context.Context, mprn string) (bool, error) {
	meters, err := e.xoserveAPI.GetMPRNTechnicalDetails(ctx, mprn)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrThirdPartyMeterpointError, err)
	}

	// Gas meter at property must not be “large capacity”
	// Large Capacity means the meter's capacity is different than 6 or 212
	if domain.IsLargeCapacity(meters) {
		logrus.WithField("mprn", mprn).Info("ineligible point-of-sale booking: gas meterpoint is not large capacity")
		return false, nil
	}

	return true, nil
}
