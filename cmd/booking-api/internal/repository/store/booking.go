package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

type BookingStore struct {
	pool  *pgxpool.Pool
	batch pgx.Batch
}

func NewBooking(pool *pgxpool.Pool) *BookingStore {
	return &BookingStore{pool: pool}
}

func (s *BookingStore) Begin() {
	s.batch = pgx.Batch{}
}

func (s *BookingStore) Commit(ctx context.Context) error {
	res := s.pool.SendBatch(ctx, &s.batch)
	s.batch = pgx.Batch{}
	return res.Close()
}

func (s *BookingStore) Upsert(booking models.Booking) {
	q := `
	INSERT INTO booking (
		booking_id,
		account_id,
		status,

		site_id,

		contact_title,
		contact_first_name,
		contact_last_name,
		contact_phone,
		contact_email,

		booking_date,
		booking_start_time,
		booking_end_time,

		vulnerabilities_list,
		vulnerabilities_other
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	ON CONFLICT (booking_id)
	DO NOTHING;
	`
	s.batch.Queue(q,
		booking.BookingID,
		booking.AccountID,
		booking.Status,
		booking.SiteID,
		booking.Contact.Title,
		booking.Contact.FirstName,
		booking.Contact.LastName,
		booking.Contact.Phone,
		booking.Contact.Email,
		booking.Slot.Date,
		booking.Slot.StartTime,
		booking.Slot.EndTime,
		booking.VulnerabilityDetails.Vulnerabilities,
		booking.VulnerabilityDetails.Other)
}

func (s *BookingStore) UpdateStatus(bookingID string, newStatus bookingv1.BookingStatus) {
	q := `
	UPDATE booking
	SET status = $2,
		updated_at = now()
	WHERE booking_id = $1;
	`
	s.batch.Queue(q, bookingID, newStatus)
}

func (s *BookingStore) GetBookingsByAccountID(ctx context.Context, accountID string) ([]models.Booking, error) {
	q := `
	SELECT
		booking_id,
		account_id,
		status,

		site_id,

		contact_title,
		contact_first_name,
		contact_last_name,
		contact_phone,
		contact_email,

		booking_date,
		booking_start_time,
		booking_end_time,

		vulnerabilities_list,
		vulnerabilities_other
	FROM booking
	WHERE account_id = $1; 
	`
	rows, err := s.pool.Query(ctx, q, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := make([]models.Booking, 0)
	for rows.Next() {
		booking := models.Booking{}
		err := rows.Scan(
			&booking.BookingID,
			&booking.AccountID,
			&booking.Status,
			&booking.SiteID,
			&booking.Contact.Title,
			&booking.Contact.FirstName,
			&booking.Contact.LastName,
			&booking.Contact.Phone,
			&booking.Contact.Email,
			&booking.Slot.Date,
			&booking.Slot.StartTime,
			&booking.Slot.EndTime,
			&booking.VulnerabilityDetails.Vulnerabilities,
			&booking.VulnerabilityDetails.Other,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}
	return bookings, nil
}
