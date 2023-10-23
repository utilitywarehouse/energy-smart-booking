package models

import (
	"time"

	"google.golang.org/protobuf/proto"
)

type PartialBooking struct {
	BookingID string
	Event     proto.Message
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
	Retries   int
}
