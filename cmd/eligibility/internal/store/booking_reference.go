package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrBookingReferenceNotFound = errors.New("booking reference not found")

type BookingRefStore struct {
	pool *pgxpool.Pool
}

func NewBookingRef(pool *pgxpool.Pool) *BookingRefStore {
	return &BookingRefStore{pool: pool}
}

func (s *BookingRefStore) Add(ctx context.Context, mpxn string, bookingRef string) error {
	q := `
	INSERT INTO booking_references (mpxn, reference)
	VALUES ($1, $2)
	ON CONFLICT (mpxn)
	DO UPDATE 
	set reference = $2, updated_at = now(), deleted_at = NULL;`

	_, err := s.pool.Exec(ctx, q, mpxn, bookingRef)

	return err
}

func (s *BookingRefStore) Remove(ctx context.Context, mpxn string) error {
	q := `
	UPDATE booking_references SET deleted_at = now()
	WHERE mpxn = $1;`
	if _, err := s.pool.Exec(ctx, q, mpxn); err != nil {
		return err
	}

	return nil
}

func (s *BookingRefStore) GetReference(ctx context.Context, mpxn string) (string, error) {
	var ref string

	q := `
	SELECT reference FROM booking_references 
	WHERE mpxn = $1
	AND deleted_at IS NULL;`
	if err := s.pool.QueryRow(ctx, q, mpxn).
		Scan(&ref); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrBookingReferenceNotFound
		}
		return "", err
	}

	return ref, nil
}
