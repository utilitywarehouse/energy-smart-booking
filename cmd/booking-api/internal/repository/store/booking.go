package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

var (
	ErrBookingNotFound = errors.New("no booking found")
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

		occupancy_id,

		contact_title,
		contact_first_name,
		contact_last_name,
		contact_phone,
		contact_email,

		booking_date,
		booking_start_time,
		booking_end_time,

		vulnerabilities_list,
		vulnerabilities_other,
		external_reference,

		booking_type
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	ON CONFLICT (booking_id)
	DO NOTHING;
	`
	vulnerabilitiesList := booking.VulnerabilityDetails.Vulnerabilities
	if vulnerabilitiesList.IsEmpty() {
		vulnerabilitiesList = models.Vulnerabilities{}
	}
	s.batch.Queue(q,
		booking.BookingID,
		booking.AccountID,
		booking.Status,
		booking.OccupancyID,
		booking.Contact.Title,
		booking.Contact.FirstName,
		booking.Contact.LastName,
		booking.Contact.Mobile,
		booking.Contact.Email,
		booking.Slot.Date,
		booking.Slot.StartTime,
		booking.Slot.EndTime,
		vulnerabilitiesList,
		booking.VulnerabilityDetails.Other,
		booking.BookingReference,
		booking.BookingType)
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

func (s *BookingStore) UpdateSchedule(bookingID string, newSchedule *time.Time) {
	q := `
	UPDATE booking
	SET booking_date = $2,
		updated_at = now()
	WHERE booking_id = $1;
	`
	s.batch.Queue(q, bookingID, *newSchedule)
}

func (s *BookingStore) UpdateBookingOnReschedule(bookingID string, contactDetails models.AccountDetails, bookingSlot models.BookingSlot, vulnerabilityDetails models.VulnerabilityDetails) {
	q := `
	UPDATE booking
	SET booking_date = $2,
		booking_start_time = $3,
		booking_end_time = $4,
		vulnerabilities_list = $5,
		vulnerabilities_other = $6,
		contact_title = $7,
		contact_first_name = $8,
		contact_last_name = $9,
		contact_phone = $10,
		contact_email = $11,
		updated_at = now()
	WHERE booking_id = $1;
	`

	vulnerabilitiesList := vulnerabilityDetails.Vulnerabilities
	if vulnerabilitiesList.IsEmpty() {
		vulnerabilitiesList = models.Vulnerabilities{}
	}
	s.batch.Queue(q, bookingID,
		bookingSlot.Date,
		bookingSlot.StartTime,
		bookingSlot.EndTime,
		vulnerabilitiesList,
		vulnerabilityDetails.Other,
		contactDetails.Title,
		contactDetails.FirstName,
		contactDetails.LastName,
		contactDetails.Mobile,
		contactDetails.Email,
	)
}

func (s *BookingStore) GetBookingsByAccountID(ctx context.Context, accountID string) ([]models.Booking, error) {
	q := `
	SELECT
		booking_id,
		account_id,
		status,

		occupancy_id,

		contact_title,
		contact_first_name,
		contact_last_name,
		contact_phone,
		contact_email,

		booking_date,
		booking_start_time,
		booking_end_time,

		vulnerabilities_list,
		vulnerabilities_other,

		external_reference,
		booking_type

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
			&booking.OccupancyID,
			&booking.Contact.Title,
			&booking.Contact.FirstName,
			&booking.Contact.LastName,
			&booking.Contact.Mobile,
			&booking.Contact.Email,
			&booking.Slot.Date,
			&booking.Slot.StartTime,
			&booking.Slot.EndTime,
			&booking.VulnerabilityDetails.Vulnerabilities,
			&booking.VulnerabilityDetails.Other,
			&booking.BookingReference,
			&booking.BookingType,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}
	return bookings, nil
}

func (s *BookingStore) GetBookingByBookingID(ctx context.Context, bookingID string) (models.Booking, error) {

	q := `
	SELECT
		booking_id,
		account_id,
		status,

		occupancy_id,

		contact_title,
		contact_first_name,
		contact_last_name,
		contact_phone,
		contact_email,

		booking_date,
		booking_start_time,
		booking_end_time,

		vulnerabilities_list,
		vulnerabilities_other,

		external_reference,

		booking_type

	FROM booking
	WHERE booking_id = $1; 
	`
	row := s.pool.QueryRow(ctx, q, bookingID)

	booking := models.Booking{}
	err := row.Scan(
		&booking.BookingID,
		&booking.AccountID,
		&booking.Status,
		&booking.OccupancyID,
		&booking.Contact.Title,
		&booking.Contact.FirstName,
		&booking.Contact.LastName,
		&booking.Contact.Mobile,
		&booking.Contact.Email,
		&booking.Slot.Date,
		&booking.Slot.StartTime,
		&booking.Slot.EndTime,
		&booking.VulnerabilityDetails.Vulnerabilities,
		&booking.VulnerabilityDetails.Other,
		&booking.BookingReference,
		&booking.BookingType,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Booking{}, ErrBookingNotFound
		}

		return models.Booking{}, fmt.Errorf("failed to scan row, %w", err)
	}

	return booking, nil
}
