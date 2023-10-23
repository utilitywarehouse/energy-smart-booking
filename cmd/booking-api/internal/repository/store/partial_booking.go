package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	ErrPartialBookingNotFound = errors.New("partial booking was not found")
)

type PartialBookingStore struct {
	pool *pgxpool.Pool
}

func NewPartialBooking(pool *pgxpool.Pool) *PartialBookingStore {
	return &PartialBookingStore{pool: pool}
}

func (s *PartialBookingStore) Upsert(ctx context.Context, bookingID string, event *bookingv1.BookingCreatedEvent) error {

	marshalledEvent, err := protojson.Marshal(event)
	if err != nil {
		return err
	}

	q := `
	INSERT INTO partial_booking (booking_id, event)
	VALUES ($1, $2)
	ON CONFLICT (booking_id)
	DO UPDATE SET event = $2;`

	_, err = s.pool.Exec(ctx, q, bookingID, marshalledEvent)
	if err != nil {
		return fmt.Errorf("failed to insert partial booking: %+v, %w", event, err)
	}

	return nil
}

func (s *PartialBookingStore) Get(ctx context.Context, bookingID string) (*models.PartialBooking, error) {

	var bID string
	var updatedAt, deletedAt, createdAt sql.NullTime
	var retries int
	var event []byte

	q := `
	SELECT booking_id, event, created_at, updated_at, deleted_at, retries
	FROM partial_booking
	WHERE booking_id = $1;`

	err := s.pool.QueryRow(ctx, q, bookingID).Scan(&bID, &event, &createdAt, &updatedAt, &deletedAt, &retries)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPartialBookingNotFound
		}
		return nil, fmt.Errorf("failed to get partial booking, %w", err)
	}

	e := &bookingv1.BookingCreatedEvent{}
	if err := protojson.Unmarshal(event, e); err != nil {
		return nil, fmt.Errorf("failed to unmarshal partial booking, %v, %w", string(event), err)
	}

	partialBooking := &models.PartialBooking{
		BookingID: bID,
		Event:     e,
		CreatedAt: createdAt.Time,
		UpdatedAt: nil,
		DeletedAt: nil,
		Retries:   retries,
	}

	if updatedAt.Valid {
		partialBooking.UpdatedAt = &updatedAt.Time
	}

	if deletedAt.Valid {
		partialBooking.DeletedAt = &deletedAt.Time
	}

	return partialBooking, nil
}

func (s *PartialBookingStore) UpdateRetries(ctx context.Context, bookingID string, retries int) error {
	q := `UPDATE partial_booking SET retries = $2 + 1, updated_at = NOW() WHERE booking_id = $1;`

	_, err := s.pool.Exec(ctx, q, bookingID, retries)

	if err != nil {
		return fmt.Errorf("failed to update retries for booking id: %s, %w", bookingID, err)
	}

	return nil
}

func (s *PartialBookingStore) MarkAsDeleted(ctx context.Context, bookingID string) error {
	q := `UPDATE partial_booking SET deleted_at = NOW() WHERE booking_id = $1;`

	_, err := s.pool.Exec(ctx, q, bookingID)

	if err != nil {
		return fmt.Errorf("failed to mark partial booking as deleted for booking id: %s, %w", bookingID, err)
	}

	return nil
}
