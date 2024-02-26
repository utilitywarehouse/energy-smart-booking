package models

import (
	"time"

	"google.golang.org/protobuf/proto"
)

type DeletionReason int32

const (
	DeletionReason_Unknown DeletionReason = 0
	// BookingCreated marks that the booking was marked as deleted due to the creation of full booking
	DeletionReason_BookingCreated DeletionReason = 1
	// BookingExpired marks that the booking was marked as deleted due to the lack of occupancy for more than three weeks
	DeletionReason_BookingExpired DeletionReason = 2
)

func MapIntToDeletionReason(reason int32) DeletionReason {
	switch reason {
	case 1:
		return DeletionReason_BookingCreated
	case 2:
		return DeletionReason_BookingExpired
	}

	return DeletionReason_Unknown
}

type PartialBooking struct {
	BookingID      string
	Event          proto.Message
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	DeletedAt      *time.Time
	Retries        int
	DeletionReason *DeletionReason
}
