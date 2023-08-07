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
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

/*booking_id             TEXT PRIMARY KEY,
  account_id             TEXT NOT NULL,
  status                 INT NOT NULL,

  -- address (normalized)
  occupancy_id           TEXT NOT NULL,

  -- contact details
  contact_title          TEXT NOT NULL,
  contact_first_name     TEXT NOT NULL,
  contact_last_name      TEXT NOT NULL,
  contact_phone          TEXT NOT NULL,
  contact_email          TEXT NOT NULL,

  -- booking slot
  booking_date           DATE NOT NULL,
  booking_start_time     INT NOT NULL,
  booking_end_time       INT NOT NULL,

  -- vulnerability details
  vulnerabilities_list   INT[] NOT NULL,
  vulnerabilities_other  TEXT NOT NULL,*/

type Booking struct {
	BookingID            string    `bigquery:"booking_id"`
	AccountID            string    `bigquery:"account_id"`
	Status               string    `bigquery:"status"`
	Source               string    `bigquery:"source"`
	OccupancyID          string    `bigquery:"occupancy_id"`
	BookingDate          time.Time `bigquery:"booking_date"`
	StartTime            int32     `bigquery:"start_time"`
	EndTime              int32     `bigquery:"end_time"`
	VulnerabilitiesList  []string  `bigquery:"vulnerabilities_list"`
	VulnerabilitiesOther string    `bigquery:vulnerabilities_other"`

	OccurredAt time.Time `bigquery:"occurred_at"`
}

type BookingIndexer struct {
	client  *bigquery.Client
	dataset string
	table   string
}

func NewBookingIndexer(client *bigquery.Client, dataset, table string) *BookingIndexer {
	return &BookingIndexer{
		client:  client,
		dataset: dataset,
		table:   table,
	}
}

func (i *BookingIndexer) Handle(ctx context.Context, message substrate.Message) error {
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
	default:
		return nil
	case *bookingv1.BookingCreatedEvent:
		update := Booking{
			BookingID:   x.GetBookingId(),
			AccountID:   x.GetDetails().AccountId,
			Status:      x.GetDetails().Status.String(),
			Source:      x.BookingSource.String(),
			OccupancyID: x.GetOccupancyId(),
			BookingDate: utilities.DateIntoTime(x.GetDetails().GetSlot().GetDate()),
			StartTime:   x.Details.Slot.GetStartTime(),
			EndTime:     x.Details.Slot.GetEndTime(),
			OccurredAt:  env.OccurredAt.AsTime(),
		}
		err = i.client.Dataset(i.dataset).Table(i.table).Inserter().Put(ctx, update)
	case *bookingv1.BookingRescheduledEvent:

		update := Booking{
			BookingID: x.GetBookingId(),
			Source:    x.BookingSource.String(),

			BookingDate: utilities.DateIntoTime(x.GetDetails().GetSlot().GetDate()),
			StartTime:   x.Details.Slot.GetStartTime(),
			EndTime:     x.Details.Slot.GetEndTime(),
			OccurredAt:  env.OccurredAt.AsTime(),
		}
		err = i.client.Dataset(i.dataset).Table(i.table).Inserter().Put(ctx, update)
	}
	if err != nil {
		return fmt.Errorf("failed to index campaignability event %s: %w", env.Uuid, err)
	}

	return nil
}
