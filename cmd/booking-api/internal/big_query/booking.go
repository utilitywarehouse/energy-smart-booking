package bq

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/sirupsen/logrus"
	energy_contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	utilities "github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/utils"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type RescheduleBooking struct {
	BookingID   string    `bigquery:"booking_id"`
	Source      string    `bigquery:"source"`
	BookingDate time.Time `bigquery:"booking_date"`
	StartTime   int32     `bigquery:"start_time"`
	EndTime     int32     `bigquery:"end_time"`

	CreatedAt time.Time `bigquery:"created_at"`
}

type Booking struct {
	BookingID          string    `bigquery:"booking_id"`
	AccountID          string    `bigquery:"account_id"`
	ExternalReference  string    `bigquery:"external_reference"`
	Source             string    `bigquery:"source"`
	Status             string    `bigquery:"status"`
	OccupancyID        string    `bigquery:"occupancy_id"`
	Title              string    `bigquery:"title"`
	FirstName          string    `bigquery:"first_name"`
	LastName           string    `bigquery:"last_name"`
	Phone              string    `bigquery:"phone"`
	Email              string    `bigquery:"email"`
	BookingDate        time.Time `bigquery:"booking_date"`
	BookingStartTime   int32     `bigquery:"start_time"`
	BookingEndTime     int32     `bigquery:"end_time"`
	VulnerabilityList  []string  `bigquery:"vulnerability_list"`
	VulnerabilityOther string    `bigquery:"vulnerability_other"`
	BookingType        string    `bigquery:"booking_type"`

	CreatedAt time.Time `bigquery:"created_at"`
}

type RescheduledBookingIndexer struct {
	client                  *bigquery.Client
	dataset                 string
	bookingsTable           string
	rescheduleBookingsTable string

	bookingBuffer           []Booking
	rescheduleBookingBuffer []RescheduleBooking
}

func NewRescheduledBookingIndexer(client *bigquery.Client, dataset, bookingsTable, rescheduleBookingsTable string) *RescheduledBookingIndexer {
	return &RescheduledBookingIndexer{
		client:                  client,
		dataset:                 dataset,
		bookingsTable:           bookingsTable,
		rescheduleBookingsTable: rescheduleBookingsTable,
	}
}

func (i *RescheduledBookingIndexer) PreHandle(_ context.Context) error {
	i.bookingBuffer = []Booking{}
	i.rescheduleBookingBuffer = []RescheduleBooking{}
	return nil
}

func (i *RescheduledBookingIndexer) PostHandle(ctx context.Context) error {
	err := i.client.Dataset(i.dataset).Table(i.rescheduleBookingsTable).Inserter().Put(ctx, i.rescheduleBookingBuffer)
	if err != nil {
		return fmt.Errorf("failed to put %d rows in reschedule bookings table, %w", len(i.rescheduleBookingBuffer), err)
	}

	err = i.client.Dataset(i.dataset).Table(i.bookingsTable).Inserter().Put(ctx, i.bookingBuffer)
	if err != nil {
		return fmt.Errorf("failed to put %d rows in bookings table, %w", len(i.bookingBuffer), err)
	}

	return nil
}

func (i *RescheduledBookingIndexer) Handle(_ context.Context, message substrate.Message) error {
	var env energy_contracts.Envelope
	if err := proto.Unmarshal(message.Data(), &env); err != nil {
		return err
	}

	if env.Message == nil {
		logrus.Info("skipping empty message")
		metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
		return nil
	}

	inner, err := env.Message.UnmarshalNew()
	if err != nil {
		return fmt.Errorf("error unmarshaling event [%s] %s: %w", env.GetUuid(), env.GetMessage().GetTypeUrl(), err)
	}
	switch x := inner.(type) {
	case *bookingv1.BookingCreatedEvent:

		bookingDate, err := utilities.DateIntoTime(x.GetDetails().GetSlot().GetDate())
		if err != nil {
			return fmt.Errorf("failed to convert date.Date type into time.Time, %w", err)
		}

		createdBooking := Booking{
			BookingID:          x.GetBookingId(),
			AccountID:          x.GetDetails().GetAccountId(),
			Status:             x.GetDetails().GetStatus().String(),
			Source:             x.BookingSource.String(),
			OccupancyID:        x.GetOccupancyId(),
			Title:              x.GetDetails().ContactDetails.Title,
			FirstName:          x.GetDetails().ContactDetails.FirstName,
			LastName:           x.GetDetails().ContactDetails.LastName,
			Phone:              x.GetDetails().ContactDetails.Phone,
			Email:              x.GetDetails().ContactDetails.Email,
			BookingDate:        *bookingDate,
			BookingStartTime:   x.GetDetails().Slot.StartTime,
			BookingEndTime:     x.GetDetails().Slot.EndTime,
			VulnerabilityList:  vulnerabilitiesAsStringSlice(x.GetDetails().VulnerabilityDetails.Vulnerabilities),
			VulnerabilityOther: x.GetDetails().VulnerabilityDetails.Other,
			ExternalReference:  x.GetDetails().ExternalReference,
			BookingType:        x.GetDetails().BookingType.String(),
			CreatedAt:          env.CreatedAt.AsTime(),
		}

		i.bookingBuffer = append(i.bookingBuffer, createdBooking)

	case *bookingv1.BookingRescheduledEvent:

		bookingDate, err := utilities.DateIntoTime(x.GetSlot().GetDate())
		if err != nil {
			return fmt.Errorf("failed to convert date.Date type into time.Time, %w", err)
		}

		rescheduleBooking := RescheduleBooking{
			BookingID:   x.GetBookingId(),
			Source:      x.BookingSource.String(),
			BookingDate: *bookingDate,
			StartTime:   x.GetSlot().StartTime,
			EndTime:     x.GetSlot().EndTime,
			CreatedAt:   env.CreatedAt.AsTime(),
		}

		i.rescheduleBookingBuffer = append(i.rescheduleBookingBuffer, rescheduleBooking)

	}
	if err != nil {
		return fmt.Errorf("failed to index campaignability event %s: %w", env.Uuid, err)
	}

	return nil
}

func vulnerabilitiesAsStringSlice(vulnerabilities []bookingv1.Vulnerability) []string {
	vulns := []string{}

	for _, vulnerability := range vulnerabilities {
		vulns = append(vulns, vulnerability.String())
	}

	return vulns
}
