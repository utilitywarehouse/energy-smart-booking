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
	pool  *pgxpool.Pool
	batch *pgx.Batch
}

func NewBookingReference(pool *pgxpool.Pool) *BookingReferenceStore {
	return &BookingReferenceStore{pool: pool}
}

func (s *BookingReferenceStore) Begin() {
	s.batch = &pgx.Batch{}
}

func (s *BookingReferenceStore) Commit(ctx context.Context) error {
	res := s.pool.SendBatch(ctx, s.batch)

	s.batch = nil
	return res.Close()
}

func (s *BookingReferenceStore) Upsert(bookingReference models.BookingReference) {
	q := `
	INSERT INTO booking_reference (mpxn, reference)
	VALUES ($1, $2)
	ON CONFLICT (mpxn)
	DO UPDATE 
	SET reference = $2, updated_at = now(), deleted_at = NULL;`

	s.batch.Queue(q, bookingReference.MPXN, bookingReference.Reference)
}

func (s *BookingReferenceStore) Remove(mpxn string) {
	q := `UPDATE booking_reference SET deleted_at = now() WHERE mpxn = $1`

	s.batch.Queue(q, mpxn)
}

func (s *BookingReferenceStore) GetReferenceByMPXN(ctx context.Context, mpxn string) (string, error) {
	var reference sql.NullString

	q := `SELECT reference FROM booking_reference WHERE mpxn = $1 AND deleted_at IS NULL;`
	if err := s.pool.QueryRow(ctx, q, mpxn).
		Scan(&reference); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrBookingReferenceNotFound
		}
		return "", err
	}

	return reference.String, nil
}
