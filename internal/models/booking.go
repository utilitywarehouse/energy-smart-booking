package models

import (
	"time"

	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
)

type BookingSlot struct {
	Date      time.Time
	StartTime int
	EndTime   int
}

// DateFromString strict enforces the parsing of Date strings into UTC Time
// objects to hopefully avoid off-by-one errors present in other services in
// the absence of a simple Date class.
func DateFromString(value string) (time.Time, error) {
	return time.ParseInLocation(time.DateOnly, value, time.UTC)
}

type Vulnerabilities []bookingv1.Vulnerability

func (v *Vulnerabilities) IsEmpty() bool {
	return v == nil || len(*v) == 0
}

type VulnerabilityDetails struct {
	Vulnerabilities Vulnerabilities
	Other           string
}

type Booking struct {
	BookingID            string
	AccountID            string
	Status               bookingv1.BookingStatus
	OccupancyID          string
	Contact              AccountDetails
	Slot                 BookingSlot
	VulnerabilityDetails VulnerabilityDetails
	BookingReference     string
	BookingType          bookingv1.BookingType
}
