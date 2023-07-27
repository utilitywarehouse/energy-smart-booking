package evaluation

import (
	"fmt"
	"strings"

	energy_domain "github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
)

type evaluation struct {
	result bool
	reason map[domain.IneligibleReason]struct{}
}

func (e *evaluation) addReason(r domain.IneligibleReason) {
	e.reason[r] = struct{}{}
}

func (e evaluation) status() domain.IneligibleReasons {
	if len(e.reason) == 0 {
		return nil
	}
	reasons := make(domain.IneligibleReasons, 0, len(e.reason))
	for k := range e.reason {
		reasons = append(reasons, k)
	}
	return reasons
}

func evaluateSuppliability(o *domain.Occupancy) (domain.IneligibleReasons, error) {
	result := evaluation{reason: make(map[domain.IneligibleReason]struct{}, 0)}

	if len(o.Services) == 0 {
		result.addReason(domain.IneligibleReasonNoActiveService)
	}

	if len(o.Services) == 1 {
		meterpoint, err := energy_domain.NewMeterPointNumber(o.Services[0].Mpxn)
		if err != nil {
			return nil, fmt.Errorf("failed to build meterpoint from mpxn %s: %w", o.Services[0].Mpxn, err)
		}
		if meterpoint.SupplyType() == energy_domain.SupplyTypeGas {
			result.addReason(domain.IneligibleReasonGasServiceOnly)
			return result.status(), nil
		}
	}

	for _, s := range o.Services {
		if s.Meterpoint == nil {
			result.addReason(domain.IneligibleReasonMissingData)
		} else if s.Meterpoint.AltHan {
			result.addReason(domain.IneligibleReasonAltHan)
		}

		if s.Meter == nil {
			result.addReason(domain.IneligibleReasonMissingData)
		} else {
			if s.Meter.Capacity == nil {
				result.addReason(domain.IneligibleReasonMissingData)
			} else if *s.Meter.Capacity != 6 && *s.Meter.Capacity != 212 {
				result.addReason(domain.IneligibleReasonMeterLargeCapacity)
			}
		}
	}

	return result.status(), nil
}

func evaluateEligibility(o *domain.Occupancy) domain.IneligibleReasons {
	result := evaluation{reason: make(map[domain.IneligibleReason]struct{}, 0)}

	if len(o.Account.PSRCodes) != 0 {
		psrCodes := strings.Join(o.Account.PSRCodes, ",")
		if strings.Contains(psrCodes, "10") ||
			strings.Contains(psrCodes, "35") ||
			strings.Contains(psrCodes, "36") {
			result.addReason(domain.IneligibleReasonPSRVulnerabilities)
		}
	}

	if o.Site == nil {
		result.addReason(domain.IneligibleReasonMissingData)
	} else if !o.Site.WanCoverage {
		result.addReason(domain.IneligibleReasonNoWanCoverage)
	}

	if len(o.Services) == 0 {
		result.addReason(domain.IneligibleReasonNoActiveService)
	}

	for _, s := range o.Services {
		if s.Meterpoint == nil {
			result.addReason(domain.IneligibleReasonMissingData)
		} else if s.Meterpoint.HasComplexTariff() {
			result.addReason(domain.IneligibleReasonComplexTariff)
		}

		if s.Meter == nil {
			result.addReason(domain.IneligibleReasonMissingData)
		} else if s.Meter.IsSmart() {
			result.addReason(domain.IneligibleReasonAlreadySmart)
		}
	}

	return result.status()
}

func evaluateCampaignability(o *domain.Occupancy) domain.IneligibleReasons {
	result := evaluation{reason: make(map[domain.IneligibleReason]struct{}, 0)}

	if o.Account.OptOut {
		result.addReason(domain.IneligibleReasonBookingOptOut)
	}

	if len(o.Services) == 0 {
		result.addReason(domain.IneligibleReasonNoActiveService)
	}

	return result.status()
}
