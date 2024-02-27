package workers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"google.golang.org/protobuf/proto"
)

const twentyOneDaysInHours float64 = 504.0

var pendingPartialBookingsMetric = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "pending_partial_bookings",
	Help: "the count of pending partial bookings",
}, []string{"type"})

var pendingPartialBookingsByAgeMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "pending_partial_bookings_age",
	Help: "total partial bookings held by age",
}, []string{"age"})

var expiredPartialBookingsMetric = promauto.NewCounter(prometheus.CounterOpts{
	Name: "expired_partial_bookings",
	Help: "the count of partial bookings marked as deleted due to the lack of occupancy",
})

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
	MarkAsDeleted(ctx context.Context, bookingID string, reason models.DeletionReason) error
}

type PartialBookingWorker struct {
	pbStore        PartialBookingStore
	occupancyStore OccupancyStore
	publisher      BookingPublisher
	alertThreshold time.Duration
}

func NewPartialBookingWorker(pbStore PartialBookingStore, occupancyStore OccupancyStore, publisher BookingPublisher, alertThreshold time.Duration) *PartialBookingWorker {
	return &PartialBookingWorker{pbStore, occupancyStore, publisher, alertThreshold}
}

func (w PartialBookingWorker) Run(ctx context.Context) error {

	pendingPartialBookings, err := w.pbStore.GetPending(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending partial bookings, %w", err)
	}

	longRetainedBookingsNr := 0

	pendingPartialBookingsMetric.WithLabelValues(PendingBookings).Add(float64(len(pendingPartialBookings)))

	for _, elem := range pendingPartialBookings {

		event := elem.Event.(*bookingv1.BookingCreatedEvent)

		occupancy, err := w.occupancyStore.GetOccupancyByAccountID(ctx, event.Details.AccountId)
		if err != nil {
			if errors.Is(err, store.ErrOccupancyNotFound) {
				err = w.pbStore.UpdateRetries(ctx, elem.BookingID, elem.Retries)
				if err != nil {
					return fmt.Errorf("failed to update retries for bookingID: %s, %w", elem.BookingID, err)
				}

				if time.Since(elem.CreatedAt).Hours() > twentyOneDaysInHours {
					err := w.pbStore.MarkAsDeleted(ctx, elem.BookingID, models.DeletionReasonBookingExpired)
					if err != nil {
						return fmt.Errorf("failed to mark bookingID: %s as deleted due to expiration, %w", elem.BookingID, err)
					}

					expiredPartialBookingsMetric.Inc()
				}

				if time.Now().Sub(elem.CreatedAt) > w.alertThreshold {
					longRetainedBookingsNr++
				}
				continue
			}

			return fmt.Errorf("failed to get occupancy by account id: %s, %w", event.Details.AccountId, err)
		}

		event.OccupancyId = occupancy.OccupancyID

		if err := w.publisher.Sink(ctx, event, time.Now()); err != nil {
			return fmt.Errorf("failed to publish booking %s, %w", elem.BookingID, err)
		}

		err = w.pbStore.MarkAsDeleted(ctx, elem.BookingID, models.DeletionReasonBookingCompleted)
		if err != nil {
			return fmt.Errorf("failed to mark bookingID: %s as deleted, %w", elem.BookingID, err)
		}

		pendingPartialBookingsMetric.WithLabelValues(ProcessedBookings).Inc()
	}
	pendingPartialBookingsByAgeMetric.WithLabelValues(w.alertThreshold.String()).Set(float64(longRetainedBookingsNr))

	return nil
}
