package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

var ErrBookingReferenceNotFound = errors.New("booking reference not found")

type BookingReferenceStore struct {
	pool *pgxpool.Pool
}

func NewBookingReference(pool *pgxpool.Pool) *BookingReferenceStore {
	return &BookingReferenceStore{pool: pool}
}

func (s *BookingReferenceStore) Upsert(ctx context.Context, bookingReference models.BookingReference) error {
	q := `
	INSERT INTO booking_reference (mpxn, reference)
	VALUES ($1, $2)
	ON CONFLICT (mpxn)
	DO UPDATE 
	SET reference = $2, updated_at = now();`

	_, err := s.pool.Exec(ctx, q, bookingReference.MPXN, bookingReference.Reference)

	return err
}

func (s *BookingReferenceStore) GetReferenceByMPXN(ctx context.Context, mpxn string) (string, error) {
	var reference sql.NullString

	q := `SELECT reference FROM booking_reference WHERE mpxn = $1;`
	if err := s.pool.QueryRow(ctx, q, mpxn).
		Scan(&reference); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrBookingReferenceNotFound
		}
		return "", err
	}

	return reference.String, nil
}
