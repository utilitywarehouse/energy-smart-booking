package evaluation

import (
	"strings"

	energy_domain "github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
)

type evaluation struct {
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

func evaluateSuppliability(o *domain.Occupancy) domain.IneligibleReasons {
	result := evaluation{reason: make(map[domain.IneligibleReason]struct{}, 0)}

	if o.Site == nil {
		result.addReason(domain.IneligibleReasonMissingSiteData)
	} else if !o.Site.WanCoverage {
		result.addReason(domain.IneligibleReasonNoWanCoverage)
	}

	if len(o.Services) == 0 {
		result.addReason(domain.IneligibleReasonNoActiveService)
	}

	for _, s := range o.Services {
		// we only project meterpoints for electricity ssc or profile class changes or if a meterpoint is alt han
		// if there is no entry, it means the meterpoint is not alt han
		// profile class and ssc are evaluated in eligibility criteria
		if s.Meterpoint != nil && s.Meterpoint.AltHan {
			result.addReason(domain.IneligibleReasonAltHan)
		}

		// we need to evaluate the meter capacity only if it's a gas service
		if s.SupplyType == energy_domain.SupplyTypeGas {
			if s.Meter == nil {
				result.addReason(domain.IneligibleReasonMissingMeterData)
			} else if s.Meter.Capacity != nil && (domain.IsLargeCapacity(s.Meter)) {
				result.addReason(domain.IneligibleReasonMeterLargeCapacity)
			}
		}
	}

	return result.status()
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
		result.addReason(domain.IneligibleReasonMissingSiteData)
	} else if !o.Site.WanCoverage {
		result.addReason(domain.IneligibleReasonNoWanCoverage)
	}

	if len(o.Services) == 0 {
		result.addReason(domain.IneligibleReasonNoActiveService)
	}

	for _, s := range o.Services {
		// complex tariff needs to be evaluated for electricity service only
		if s.SupplyType == energy_domain.SupplyTypeElectricity {
			if s.Meterpoint == nil {
				result.addReason(domain.IneligibleReasonMissingMeterpointData)
			} else if domain.HasComplexSSC(s.Meterpoint) {
				result.addReason(domain.IneligibleReasonComplexTariff)
			}
		}

		if s.Meter == nil {
			result.addReason(domain.IneligibleReasonMissingMeterData)
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
