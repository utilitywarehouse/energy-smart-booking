package domain

import (
	"database/sql/driver"
	"errors"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
)

var ErrInvalidReason = errors.New("invalid ineligible reason")

type IneligibleReason uint8

const (
	IneligibleReasonUnknown IneligibleReason = iota
	IneligibleReasonNoWanCoverage
	IneligibleReasonNoActiveService
	IneligibleReasonMeterLargeCapacity
	IneligibleReasonPSRVulnerabilities
	IneligibleReasonAbortedBookings
	IneligibleReasonBookingScheduled
	IneligibleReasonBookingCompleted
	IneligibleReasonComplexTariff
	IneligibleReasonBookingReferenceMissing
	IneligibleReasonAlreadySmart
	IneligibleReasonMissingData
	IneligibleReasonAltHan
	IneligibleReasonBookingOptOut
	IneligibleReasonGasServiceOnly
)

type IneligibleReasons []IneligibleReason

func (r IneligibleReason) String() string {
	switch r {
	default:
		return "unknown"
	case IneligibleReasonNoWanCoverage:
		return "NoWanCoverage"
	case IneligibleReasonNoActiveService:
		return "NoActiveService"
	case IneligibleReasonMeterLargeCapacity:
		return "LargeCapacityMeter"
	case IneligibleReasonPSRVulnerabilities:
		return "AccountPSRVulnerabilities"
	case IneligibleReasonAbortedBookings:
		return "AbortedBookings"
	case IneligibleReasonBookingScheduled:
		return "BookingScheduled"
	case IneligibleReasonBookingCompleted:
		return "BookingCompleted"
	case IneligibleReasonComplexTariff:
		return "ComplexTariff"
	case IneligibleReasonBookingReferenceMissing:
		return "BookingReferenceMissing"
	case IneligibleReasonAlreadySmart:
		return "AlreadySmart"
	case IneligibleReasonMissingData:
		return "MissingData"
	case IneligibleReasonAltHan:
		return "AltHan"
	case IneligibleReasonBookingOptOut:
		return "OptOut"
	case IneligibleReasonGasServiceOnly:
		return "GasServiceOnly"
	}
}

func fromString(str string) (IneligibleReason, error) {
	switch str {
	default:
		return 0, ErrInvalidReason
	case "unknown":
		return IneligibleReasonUnknown, nil
	case "NoWanCoverage":
		return IneligibleReasonNoWanCoverage, nil
	case "NoActiveService":
		return IneligibleReasonNoActiveService, nil
	case "LargeCapacityMeter":
		return IneligibleReasonMeterLargeCapacity, nil
	case "AccountPSRVulnerabilities":
		return IneligibleReasonPSRVulnerabilities, nil
	case "AbortedBookings":
		return IneligibleReasonAbortedBookings, nil
	case "BookingScheduled":
		return IneligibleReasonBookingScheduled, nil
	case "ComplexTariff":
		return IneligibleReasonComplexTariff, nil
	case "BookingReferenceMissing":
		return IneligibleReasonBookingReferenceMissing, nil
	case "AlreadySmart":
		return IneligibleReasonAlreadySmart, nil
	case "MissingData":
		return IneligibleReasonMissingData, nil
	case "AltHan":
		return IneligibleReasonAltHan, nil
	case "OptOut":
		return IneligibleReasonBookingOptOut, nil
	case "GasServiceOnly":
		return IneligibleReasonGasServiceOnly, nil
	}
}

func (r IneligibleReasons) Value() (driver.Value, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	strReasons := make([]string, 0)
	for _, reason := range r {
		strReasons = append(strReasons, reason.String())
	}
	return json.Marshal(strReasons)
}

func (r *IneligibleReasons) Scan(value interface{}) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	switch x := value.(type) {
	case nil:
		return nil
	case string:
		var strReasons []string
		err := json.Unmarshal([]byte(x), &strReasons)
		if err != nil {
			return err
		}
		if r == nil {
			*r = make([]IneligibleReason, 0)
		}
		for _, str := range strReasons {
			reason, err := fromString(str)
			if err != nil {
				return err
			}
			*r = append(*r, reason)
		}
	}
	return nil
}

func MapIneligibleProtoToDomainReason(reason smart.IneligibleReason) (IneligibleReason, error) {
	switch reason {
	case smart.IneligibleReason_INELIGIBLE_REASON_UNKNOWN:
		return IneligibleReasonUnknown, nil
	case smart.IneligibleReason_INELIGIBLE_REASON_NO_WAN_COVERAGE:
		return IneligibleReasonNoWanCoverage, nil
	case smart.IneligibleReason_INELIGIBLE_REASON_NOT_ACTIVE:
		return IneligibleReasonNoActiveService, nil
	case smart.IneligibleReason_INELIGIBLE_REASON_METER_LARGE_CAPACITY:
		return IneligibleReasonMeterLargeCapacity, nil
	case smart.IneligibleReason_INELIGIBLE_REASON_ACCOUNT_PSR_VULNERABILITIES:
		return IneligibleReasonPSRVulnerabilities, nil
	case smart.IneligibleReason_INELIGIBLE_REASON_ABORTED_BOOKINGS:
		return IneligibleReasonAbortedBookings, nil
	case smart.IneligibleReason_INELIGIBLE_REASON_BOOKING_SCHEDULED:
		return IneligibleReasonBookingScheduled, nil
	case smart.IneligibleReason_INELIGIBLE_REASON_BOOKING_COMPLETED:
		return IneligibleReasonBookingCompleted, nil
	case smart.IneligibleReason_INELIGIBLE_REASON_COMPLEX_TARIFF:
		return IneligibleReasonComplexTariff, nil
	case smart.IneligibleReason_INELIGIBLE_REASON_BOOKING_REFERENCE_MISSING:
		return IneligibleReasonBookingReferenceMissing, nil
	case smart.IneligibleReason_INELIGIBLE_REASON_ALREADY_SMART:
		return IneligibleReasonAlreadySmart, nil
	case smart.IneligibleReason_INELIGIBLE_REASON_MISSING_DATA:
		return IneligibleReasonMissingData, nil
	case smart.IneligibleReason_INELIGIBLE_REASON_ALT_HAN:
		return IneligibleReasonAltHan, nil
	case smart.IneligibleReason_INELIGIBLE_REASON_SMART_BOOKING_OPT_OUT:
		return IneligibleReasonBookingOptOut, nil
	case smart.IneligibleReason_INELIGIBLE_REASON_GAS_SERVICE_ONLY:
		return IneligibleReasonGasServiceOnly, nil
	}

	return IneligibleReasonUnknown, fmt.Errorf("reason not mapped")
}

func (r IneligibleReasons) MapToProto() ([]smart.IneligibleReason, error) {
	reasons := make([]smart.IneligibleReason, 0, len(r))
	for _, reason := range r {
		pReason, err := mapDomainToProtoReason(reason)
		if err != nil {
			return nil, err
		}
		reasons = append(reasons, pReason)
	}

	return reasons, nil
}

func (r IneligibleReasons) ToString() []string {
	values := make([]string, 0, len(r))

	for _, reason := range r {
		values = append(values, reason.String())
	}

	return values
}

func (r IneligibleReasons) Contains(reason IneligibleReason) bool {
	for _, rr := range r {
		if rr == reason {
			return true
		}
	}
	return false
}

func mapDomainToProtoReason(reason IneligibleReason) (smart.IneligibleReason, error) {
	switch reason {
	case IneligibleReasonUnknown:
		return smart.IneligibleReason_INELIGIBLE_REASON_UNKNOWN, nil
	case IneligibleReasonNoWanCoverage:
		return smart.IneligibleReason_INELIGIBLE_REASON_NO_WAN_COVERAGE, nil
	case IneligibleReasonNoActiveService:
		return smart.IneligibleReason_INELIGIBLE_REASON_NOT_ACTIVE, nil
	case IneligibleReasonMeterLargeCapacity:
		return smart.IneligibleReason_INELIGIBLE_REASON_METER_LARGE_CAPACITY, nil
	case IneligibleReasonPSRVulnerabilities:
		return smart.IneligibleReason_INELIGIBLE_REASON_ACCOUNT_PSR_VULNERABILITIES, nil
	case IneligibleReasonAbortedBookings:
		return smart.IneligibleReason_INELIGIBLE_REASON_ABORTED_BOOKINGS, nil
	case IneligibleReasonBookingScheduled:
		return smart.IneligibleReason_INELIGIBLE_REASON_BOOKING_SCHEDULED, nil
	case IneligibleReasonBookingCompleted:
		return smart.IneligibleReason_INELIGIBLE_REASON_BOOKING_COMPLETED, nil
	case IneligibleReasonComplexTariff:
		return smart.IneligibleReason_INELIGIBLE_REASON_COMPLEX_TARIFF, nil
	case IneligibleReasonBookingReferenceMissing:
		return smart.IneligibleReason_INELIGIBLE_REASON_BOOKING_REFERENCE_MISSING, nil
	case IneligibleReasonAlreadySmart:
		return smart.IneligibleReason_INELIGIBLE_REASON_ALREADY_SMART, nil
	case IneligibleReasonMissingData:
		return smart.IneligibleReason_INELIGIBLE_REASON_MISSING_DATA, nil
	case IneligibleReasonAltHan:
		return smart.IneligibleReason_INELIGIBLE_REASON_ALT_HAN, nil
	case IneligibleReasonBookingOptOut:
		return smart.IneligibleReason_INELIGIBLE_REASON_SMART_BOOKING_OPT_OUT, nil
	case IneligibleReasonGasServiceOnly:
		return smart.IneligibleReason_INELIGIBLE_REASON_GAS_SERVICE_ONLY, nil
	default:
		return smart.IneligibleReason_INELIGIBLE_REASON_UNKNOWN, fmt.Errorf("reason not mapped")
	}
}
