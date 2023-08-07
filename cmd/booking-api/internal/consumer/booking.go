package consumer

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/uw-labs/substrate"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/protobuf/proto"
)

type BookingStore interface {
	Upsert(models.Booking)
	UpdateStatus(bookingID string, newStatus bookingv1.BookingStatus)
	UpdateSchedule(bookingID string, newSchedule *time.Time)

	Begin()
	Commit(context.Context) error
}

type OccupancyReadOnlyStore interface {
	GetOccupancyByID(ctx context.Context, occupancyID string) (*models.Occupancy, error)
}

type BookingHandler struct {
	bookingStore   BookingStore
	occupancyStore OccupancyReadOnlyStore
}

func HandleBooking(bookings BookingStore, occupancies OccupancyReadOnlyStore) *BookingHandler {
	return &BookingHandler{bookingStore: bookings, occupancyStore: occupancies}
}

func (h *BookingHandler) PreHandle(_ context.Context) error {
	h.bookingStore.Begin()
	return nil
}

func (h *BookingHandler) PostHandle(ctx context.Context) error {
	return h.bookingStore.Commit(ctx)
}

func dateIntoTime(d *date.Date) (*time.Time, error) {
	t, err := time.ParseInLocation(
		time.DateOnly,
		fmt.Sprintf("%d-%02d-%02d", d.Year, d.Month, d.Day),
		time.UTC,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (h *BookingHandler) Handle(ctx context.Context, message substrate.Message) error {
	var env generated.Envelope
	if err := proto.Unmarshal(message.Data(), &env); err != nil {
		return err
	}

	eventUuid := env.Uuid
	if env.Message == nil {
		log.WithField("event-uuid", eventUuid).Info("skipping empty message", eventUuid)
		metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
		return nil
	}

	payload, err := env.Message.UnmarshalNew()
	if err != nil {
		return fmt.Errorf("failed to unmarshall event in booking topic [%s|%s]: %w", eventUuid, env.Message.TypeUrl, err)
	}
	switch ev := payload.(type) {
	case *bookingv1.BookingCreatedEvent:
		details := ev.GetDetails()
		contactDetails := details.GetContactDetails()
		slot := details.GetSlot()
		dateTime, err := dateIntoTime(slot.GetDate())
		if err != nil {
			return err
		}
		vulns := details.GetVulnerabilityDetails()

		h.bookingStore.Upsert(models.Booking{
			BookingID:   ev.GetBookingId(),
			AccountID:   details.GetAccountId(),
			Status:      details.GetStatus(),
			OccupancyID: ev.GetOccupancyId(),
			Contact: models.AccountDetails{
				Title:     contactDetails.GetTitle(),
				FirstName: contactDetails.GetFirstName(),
				LastName:  contactDetails.GetLastName(),
				Email:     contactDetails.GetEmail(),
				Mobile:    contactDetails.GetPhone(),
			},
			Slot: models.BookingSlot{
				Date:      *dateTime,
				StartTime: int(slot.GetStartTime()),
				EndTime:   int(slot.GetEndTime()),
			},
			VulnerabilityDetails: models.VulnerabilityDetails{
				Vulnerabilities: vulns.GetVulnerabilities(),
				Other:           vulns.GetOther(),
			},
		})
	case *bookingv1.BookingRescheduledEvent:
		bookingID := ev.GetBookingId()
		dt, err := dateIntoTime(ev.GetSlot().GetDate())
		if err != nil {
			return err
		}
		h.bookingStore.UpdateSchedule(bookingID, dt)
	}

	return nil
}
