package bq

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/sirupsen/logrus"
	energy_contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	smart "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type Eligible struct {
	OccupancyID string    `bigquery:"occupancy_id"`
	AccountID   string    `bigquery:"account_id"`
	Eligible    bool      `bigquery:"eligible"`
	Reason      []string  `bigquery:"reason"`
	UpdatedAt   time.Time `bigquery:"updated_at"`
}

type EligibilityIndexer struct {
	client  *bigquery.Client
	dataset string
	table   string
}

func NewEligibilityIndexer(client *bigquery.Client, dataset, table string) *EligibilityIndexer {
	return &EligibilityIndexer{
		client:  client,
		dataset: dataset,
		table:   table,
	}
}

func (i *EligibilityIndexer) Handle(ctx context.Context, message substrate.Message) error {
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
	case *smart.EligibleOccupancyAddedEvent:
		update := Eligible{
			OccupancyID: x.GetOccupancyId(),
			AccountID:   x.GetAccountId(),
			Eligible:    true,
			Reason:      nil,
			UpdatedAt:   env.OccurredAt.AsTime(),
		}
		err = i.client.Dataset(i.dataset).Table(i.table).Inserter().Put(ctx, update)
	case *smart.EligibleOccupancyRemovedEvent:
		update := Eligible{
			OccupancyID: x.GetOccupancyId(),
			AccountID:   x.GetAccountId(),
			Eligible:    false,
			Reason:      stringReasons(x.GetReasons()),
			UpdatedAt:   env.OccurredAt.AsTime(),
		}
		err = i.client.Dataset(i.dataset).Table(i.table).Inserter().Put(ctx, update)
	}
	if err != nil {
		return fmt.Errorf("failed to index eligibility event %s: %w", env.Uuid, err)
	}

	return nil
}

func stringReasons(reasons []smart.IneligibleReason) []string {
	r := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		r = append(r, reason.String())
	}

	return r
}
