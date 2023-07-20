package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

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
	}
}

func (r IneligibleReasons) Value() (driver.Value, error) {
	strReasons := make([]string, 0)
	for _, reason := range r {
		strReasons = append(strReasons, reason.String())
	}
	return json.Marshal(strReasons)
}

func (r *IneligibleReasons) Scan(value interface{}) error {
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
	}

	return IneligibleReasonUnknown, fmt.Errorf("reason not mapped")
}
