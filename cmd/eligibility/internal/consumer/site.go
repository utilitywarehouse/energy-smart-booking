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

type SiteStore interface {
	Add(ctx context.Context, id, postCode string, at time.Time) error
}

type OccupancySiteStore interface {
	GetIDsBySite(ctx context.Context, siteID string) ([]string, error)
}

func HandleSite(store SiteStore, occupancyStore OccupancySiteStore, evaluator Evaluator, stateRebuild bool) substratemessage.BatchHandlerFunc {
	return func(ctx context.Context, messages []substrate.Message) error {
		for _, msg := range messages {
			var env energy_contracts.Envelope
			if err := proto.Unmarshal(msg.Data(), &env); err != nil {
				return err
			}

			if env.Message == nil {
				log.Info("skipping empty site message")
				metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
				continue
			}

			inner, err := env.Message.UnmarshalNew()
			if err != nil {
				return fmt.Errorf("error unmarshaling site event [%s] %s: %w", env.GetUuid(), env.GetMessage().GetTypeUrl(), err)
			}
			switch x := inner.(type) {
			case *platform.SiteDiscoveredEvent:
				if x.GetAddress() == nil {
					log.Infof("skip site event %s: empty address", x.GetSiteId())
					continue
				}
				if x.GetAddress().GetPostcode() == "" {
					log.Infof("skip site event %s: empty post code", x.GetSiteId())
					continue
				}
				err = store.Add(ctx, x.GetSiteId(), x.GetAddress().GetPostcode(), env.OccurredAt.AsTime())
				if err != nil {
					return fmt.Errorf("failed to process site event %s: %w", env.Uuid, err)
				}

				if !stateRebuild {
					occupanciesIDs, err := occupancyStore.GetIDsBySite(ctx, x.GetSiteId())
					if err != nil {
						return fmt.Errorf("failed to get occupancies for msg %s, site ID %s: %w", env.GetUuid(), x.GetSiteId(), err)
					}
					for _, occupancyID := range occupanciesIDs {
						err = evaluator.RunFull(ctx, occupancyID)
						if err != nil {
							return fmt.Errorf("failed to run evaluation for msg %s, occupancyID %s: %w", env.GetUuid(), occupancyID, err)
						}
					}
				}
			}
		}
		return nil
	}
}
