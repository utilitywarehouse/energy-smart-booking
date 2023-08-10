package consumer

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	energy_contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-pkg/substratemessage/v2"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type SuppliabilityStore interface {
	Add(ctx context.Context, occupancyID, accountID string, reasons domain.IneligibleReasons) error
}

func HandleSuppliability(store SuppliabilityStore) substratemessage.BatchHandlerFunc {
	return func(ctx context.Context, messages []substrate.Message) error {
		for _, msg := range messages {
			var env energy_contracts.Envelope
			if err := proto.Unmarshal(msg.Data(), &env); err != nil {
				return err
			}

			if env.Message == nil {
				log.Info("skipping empty suppliability message")
				metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
				continue
			}

			inner, err := env.Message.UnmarshalNew()
			if err != nil {
				return fmt.Errorf("error unmarshaling suppliability event [%s] %s: %w", env.GetUuid(), env.GetMessage().GetTypeUrl(), err)
			}
			switch x := inner.(type) {
			case *smart.SuppliableOccupancyAddedEvent:
				err = store.Add(ctx, x.GetOccupancyId(), x.GetAccountId(), nil)
			case *smart.SuppliableOccupancyRemovedEvent:
				var reasons domain.IneligibleReasons
				protoReasons := x.GetReasons()
				for _, r := range protoReasons {
					ir, err := domain.MapIneligibleProtoToDomainReason(r)
					if err != nil {
						return fmt.Errorf("failed to process suppliability event %s: %w", env.GetUuid(), err)
					}
					reasons = append(reasons, ir)
				}
				err = store.Add(ctx, x.GetOccupancyId(), x.GetAccountId(), reasons)
			}
			if err != nil {
				return fmt.Errorf("failed to process suppliability event %s: %w", env.GetUuid(), err)
			}
		}
		return nil
	}
}
