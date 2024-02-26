package models

import (
	"time"

	"google.golang.org/protobuf/proto"
)

type DeletionReason int32

const (
	// BookingCreated marks that the booking was marked as deleted due to the creation of full booking
	BookingCreated DeletionReason = iota
	// BookingExpired marks that the booking was marked as deleted due to the lack of occupancy for more than three weeks
	BookingExpired
)

type PartialBooking struct {
	BookingID      string
	Event          proto.Message
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	DeletedAt      *time.Time
	Retries        int
	DeletionReason *DeletionReason
}
