package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"google.golang.org/protobuf/proto"
)

var PendingPartialBookings = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "pending_partial_bookings",
	Help: "the count of pending partial bookings",
}, []string{"type"})

const (
	PendingBookings   = "pending_bookings"
	ProcessedBookings = "processed_bookings"
)

type BookingPublisher interface {
	Sink(ctx context.Context, proto proto.Message, at time.Time) error
}

type OccupancyStore interface {
	GetOccupancyByAccountID(context.Context, string) (*models.Occupancy, error)
}

type PartialBookingStore interface {
	Upsert(ctx context.Context, bookingID string, event *bookingv1.BookingCreatedEvent) error
	GetPending(ctx context.Context) ([]*models.PartialBooking, error)
	UpdateRetries(ctx context.Context, bookingID string, retries int) error
	MarkAsDeleted(ctx context.Context, bookingID string) error
}

type PartialBookingWorker struct {
	pbStore        PartialBookingStore
	occupancyStore OccupancyStore
	publisher      BookingPublisher
}

func NewPartialBookingWorker(pbStore PartialBookingStore, occupancyStore OccupancyStore, publisher BookingPublisher) *PartialBookingWorker {
	return &PartialBookingWorker{pbStore, occupancyStore, publisher}
}

func (w PartialBookingWorker) Run(ctx context.Context) error {

	pendingPartialBookings, err := w.pbStore.GetPending(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending partial bookings, %w", err)
	}

	PendingPartialBookings.WithLabelValues(PendingBookings).Add(float64(len(pendingPartialBookings)))

	for _, elem := range pendingPartialBookings {

		event := elem.Event.(*bookingv1.BookingCreatedEvent)

		occupancy, err := w.occupancyStore.GetOccupancyByAccountID(ctx, event.Details.AccountId)
		if err != nil {
			return fmt.Errorf("failed to get occupancy by account id: %s, %w", event.Details.AccountId, err)
		}

		event.OccupancyId = occupancy.OccupancyID

		if err := w.publisher.Sink(ctx, event, time.Now()); err != nil {
			err = w.pbStore.UpdateRetries(ctx, elem.BookingID, elem.Retries)
			if err != nil {
				return fmt.Errorf("failed to update retries for bookingID: %s, %w", elem.BookingID, err)
			}

			return fmt.Errorf("failed to publish occupancy, %w", err)
		}

		err = w.pbStore.MarkAsDeleted(ctx, elem.BookingID)
		if err != nil {
			return fmt.Errorf("failed to mark bookingID: %s as deleted, %w", elem.BookingID, err)
		}

		PendingPartialBookings.WithLabelValues(ProcessedBookings).Inc()
	}

	return nil
}
