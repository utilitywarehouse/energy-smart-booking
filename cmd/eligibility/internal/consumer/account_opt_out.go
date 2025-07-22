package consumer

import (
	"context"
	"fmt"
	"log/slog"

	energy_contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	smart "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-pkg/substratemessage"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type AccountOptOutStore interface {
	AddOptOut(ctx context.Context, accountID string, optOut bool) error
}

type OccupancyOptOutStore interface {
	GetIDsByAccount(ctx context.Context, accountID string) ([]string, error)
}

type CampaignableEvaluator interface {
	RunCampaignability(ctx context.Context, occupancyID string) error
}

func HandleAccountOptOut(store AccountOptOutStore, occupancyStore OccupancyOptOutStore, evaluator CampaignableEvaluator, stateRebuild bool) substratemessage.BatchHandlerFunc {
	return func(ctx context.Context, messages []substrate.Message) error {
		for _, msg := range messages {
			var env energy_contracts.Envelope
			if err := proto.Unmarshal(msg.Data(), &env); err != nil {
				return err
			}

			if env.Message == nil {
				slog.Info("skipping empty account opt out message")
				metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
				continue
			}

			inner, err := env.Message.UnmarshalNew()
			if err != nil {
				return fmt.Errorf("error unmarshaling account opt out event [%s] %s: %w", env.GetUuid(), env.GetMessage().GetTypeUrl(), err)
			}

			switch x := inner.(type) {
			case *smart.AccountBookingOptOutAddedEvent:
				err = store.AddOptOut(ctx, x.AccountId, true)
			case *smart.AccountBookingOptOutRemovedEvent:
				err = store.AddOptOut(ctx, x.GetAccountId(), false)
			}
			if err != nil {
				return fmt.Errorf("failed to handle account opt out event %s: %w", env.GetUuid(), err)
			}

			if !stateRebuild {
				accountID := inner.(optOutIdentifier).GetAccountId()
				occupanciesIDs, err := occupancyStore.GetIDsByAccount(ctx, accountID)
				if err != nil {
					return fmt.Errorf("failed to get occupancies for msg %s, account ID %s: %w", env.GetUuid(), accountID, err)
				}
				for _, occupancyID := range occupanciesIDs {
					err = evaluator.RunCampaignability(ctx, occupancyID)
					if err != nil {
						return fmt.Errorf("failed to run campaignability for account opt out msg %s, occupancyID %s: %w", env.GetUuid(), occupancyID, err)
					}
				}
			}
		}
		return nil
	}
}

type optOutIdentifier interface {
	GetAccountId() string
}
