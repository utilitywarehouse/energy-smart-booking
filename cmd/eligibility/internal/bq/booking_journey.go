package bq

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/sirupsen/logrus"
	energy_contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type BookingJourneyEligibilityRef struct {
	OccupancyID string    `bigquery:"occupancy_id"`
	BookingRef  string    `bigquery:"reference"`
	Eligible    bool      `bigquery:"eligible"`
	UpdatedAt   time.Time `bigquery:"updated_at"`
}

type BookingJourneyEligibilityIndexer struct {
	client  *bigquery.Client
	dataset string
	table   string
	buffer  []BookingJourneyEligibilityRef
}

func (i *BookingJourneyEligibilityIndexer) PreHandle(_ context.Context) error {
	i.buffer = []BookingJourneyEligibilityRef{}
	return nil
}

func (i *BookingJourneyEligibilityIndexer) PostHandle(ctx context.Context) error {
	err := i.client.Dataset(i.dataset).Table(i.table).Inserter().Put(ctx, i.buffer)
	if err != nil {
		return fmt.Errorf("failed to put %d rows in booking_journey_reference table, %w", len(i.buffer), err)
	}

	return nil
}

func NewBookingJourneyEligibilityIndexer(client *bigquery.Client, dataset, table string) *BookingJourneyEligibilityIndexer {
	return &BookingJourneyEligibilityIndexer{
		client:  client,
		dataset: dataset,
		table:   table,
	}
}

func (i *BookingJourneyEligibilityIndexer) Handle(_ context.Context, message substrate.Message) error {
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
	case *smart.SmartBookingJourneyOccupancyAddedEvent:
		update := BookingJourneyEligibilityRef{
			OccupancyID: x.GetOccupancyId(),
			BookingRef:  x.GetReference(),
			Eligible:    true,
			UpdatedAt:   env.OccurredAt.AsTime(),
		}
		i.buffer = append(i.buffer, update)

	case *smart.SmartBookingJourneyOccupancyRemovedEvent:
		update := BookingJourneyEligibilityRef{
			OccupancyID: x.GetOccupancyId(),
			Eligible:    false,
			UpdatedAt:   env.OccurredAt.AsTime(),
		}
		i.buffer = append(i.buffer, update)
	}

	return nil
}
