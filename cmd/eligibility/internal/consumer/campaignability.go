package consumer

import (
	"context"
	"fmt"
	"log/slog"

	energy_contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	smart "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-pkg/substratemessage"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type CampaignabilityStore interface {
	Add(ctx context.Context, occupancyID, accountID string, reasons domain.IneligibleReasons) error
}

func HandleCampaignability(store SuppliabilityStore) substratemessage.BatchHandlerFunc {
	return func(ctx context.Context, messages []substrate.Message) error {
		for _, msg := range messages {
			var env energy_contracts.Envelope
			if err := proto.Unmarshal(msg.Data(), &env); err != nil {
				return err
			}

			if env.Message == nil {
				slog.Info("skipping empty campaignability message")
				metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
				continue
			}

			inner, err := env.Message.UnmarshalNew()
			if err != nil {
				return fmt.Errorf("error unmarshaling campaignability event [%s] %s: %w", env.GetUuid(), env.GetMessage().GetTypeUrl(), err)
			}
			switch x := inner.(type) {
			case *smart.CampaignableOccupancyAddedEvent:
				err = store.Add(ctx, x.GetOccupancyId(), x.GetAccountId(), nil)
			case *smart.CampaignableOccupancyRemovedEvent:
				var reasons domain.IneligibleReasons
				protoReasons := x.GetReasons()
				for _, r := range protoReasons {
					ir, err := domain.MapIneligibleProtoToDomainReason(r)
					if err != nil {
						return fmt.Errorf("failed to process campaignability event %s: %w", env.GetUuid(), err)
					}
					reasons = append(reasons, ir)
				}
				err = store.Add(ctx, x.GetOccupancyId(), x.GetAccountId(), reasons)
			}
			if err != nil {
				return fmt.Errorf("failed to process campaignability event %s: %w", env.GetUuid(), err)
			}
		}
		return nil
	}
}
