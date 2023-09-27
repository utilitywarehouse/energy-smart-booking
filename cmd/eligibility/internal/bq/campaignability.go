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

type Campaignable struct {
	AccountID    string    `bigquery:"account_id"`
	OccupancyID  string    `bigquery:"occupancy_id"`
	Campaignable bool      `bigquery:"campaignable"`
	Reason       []string  `bigquery:"reason"`
	UpdatedAt    time.Time `bigquery:"updated_at"`
}

type CampaignabilityIndexer struct {
	client  *bigquery.Client
	dataset string
	table   string
}

func NewCampaignabilityIndexer(client *bigquery.Client, dataset, table string) *CampaignabilityIndexer {
	return &CampaignabilityIndexer{
		client:  client,
		dataset: dataset,
		table:   table,
	}
}

func (i *CampaignabilityIndexer) Handle(ctx context.Context, message substrate.Message) error {
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
	case *smart.CampaignableOccupancyAddedEvent:
		update := Campaignable{
			OccupancyID:  x.GetOccupancyId(),
			AccountID:    x.GetAccountId(),
			Campaignable: true,
			Reason:       nil,
			UpdatedAt:    env.OccurredAt.AsTime(),
		}
		err = i.client.Dataset(i.dataset).Table(i.table).Inserter().Put(ctx, update)
	case *smart.CampaignableOccupancyRemovedEvent:
		update := Campaignable{
			OccupancyID:  x.GetOccupancyId(),
			AccountID:    x.GetAccountId(),
			Campaignable: false,
			Reason:       stringReasons(x.GetReasons()),
			UpdatedAt:    env.OccurredAt.AsTime(),
		}
		err = i.client.Dataset(i.dataset).Table(i.table).Inserter().Put(ctx, update)
	}
	if err != nil {
		return fmt.Errorf("failed to index campaignability event %s: %w", env.Uuid, err)
	}

	return nil
}
