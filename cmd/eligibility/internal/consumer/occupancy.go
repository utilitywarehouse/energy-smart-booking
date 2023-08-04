package consumer

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	energy_contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-pkg/substratemessage"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type OccupancyStore interface {
	Add(ctx context.Context, id, siteID, accountID string, at time.Time) error
	AddSite(ctx context.Context, occupancyID, siteID string) error
}

func HandleOccupancy(store OccupancyStore, evaluator Evaluator, stateRebuild bool) substratemessage.BatchHandlerFunc {
	return func(ctx context.Context, messages []substrate.Message) error {
		for _, msg := range messages {
			var env energy_contracts.Envelope
			if err := proto.Unmarshal(msg.Data(), &env); err != nil {
				return err
			}

			if env.Message == nil {
				log.Info("skipping empty occupancy message")
				metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
				continue
			}

			inner, err := env.Message.UnmarshalNew()
			if err != nil {
				return fmt.Errorf("error unmarshaling occupancy event [%s] %s: %w", env.GetUuid(), env.GetMessage().GetTypeUrl(), err)
			}
			switch x := inner.(type) {
			default:
				return nil
			case *platform.OccupancyStartedEvent:
				err = store.Add(ctx, x.GetOccupancyId(), x.GetSiteId(), x.GetCustomerAccountId(), env.OccurredAt.AsTime())
			case *platform.OccupancySiteCorrectedEvent:
				err = store.AddSite(ctx, x.GetOccupancyId(), x.GetSiteId())
			}
			if err != nil {
				return fmt.Errorf("failed to process occupancy event %s: %w", env.Uuid, err)
			}

			if !stateRebuild {
				occupancyID := inner.(occupancyIdentifier).GetOccupancyId()
				err = evaluator.RunFull(ctx, occupancyID)
				if err != nil {
					return fmt.Errorf("failed to run evaluation for occupancy msg %s, occupancyID %s: %w", env.GetUuid(), occupancyID, err)
				}
			}
		}
		return nil
	}
}

type occupancyIdentifier interface {
	GetOccupancyId() string
}
