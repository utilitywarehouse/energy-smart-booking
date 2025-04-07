package consumer

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	energy_contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	smart "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-pkg/substratemessage"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type AccountPSRStore interface {
	AddPSRCodes(ctx context.Context, accountID string, codes []string) error
}

type OccupancyPSRStore interface {
	GetIDsByAccount(ctx context.Context, accountID string) ([]string, error)
}

type EligibileEvaluator interface {
	RunEligibility(ctx context.Context, occupancyID string) error
}

func HandleAccountPSR(store AccountPSRStore, occupancyStore OccupancyPSRStore, evaluator EligibileEvaluator, stateRebuild bool) substratemessage.BatchHandlerFunc {
	return func(ctx context.Context, messages []substrate.Message) error {
		for _, msg := range messages {
			var env energy_contracts.Envelope
			if err := proto.Unmarshal(msg.Data(), &env); err != nil {
				return err
			}

			if env.Message == nil {
				log.Info("skipping empty account psr message")
				metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
				continue
			}

			inner, err := env.Message.UnmarshalNew()
			if err != nil {
				return fmt.Errorf("error unmarshaling account psr event [%s] %s: %w", env.GetUuid(), env.GetMessage().GetTypeUrl(), err)
			}
			switch x := inner.(type) {
			case *smart.AccountPSRCodesChangedEvent:
				err = store.AddPSRCodes(ctx, x.AccountId, x.GetCodes())
			case *smart.AccountPSRCodesRemovedEvent:
				err = store.AddPSRCodes(ctx, x.GetAccountId(), nil)
			}
			if err != nil {
				return fmt.Errorf("failed to handle account psr codes event %s: %w", env.GetUuid(), err)
			}

			if !stateRebuild {
				//nolint return value not check on interface assertion
				accountID := inner.(psrIdentifier).GetAccountId()
				occupanciesIDs, err := occupancyStore.GetIDsByAccount(ctx, accountID)
				if err != nil {
					return fmt.Errorf("failed to get occupancies for msg %s, account ID %s: %w", env.GetUuid(), accountID, err)
				}
				for _, occupancyID := range occupanciesIDs {
					err = evaluator.RunEligibility(ctx, occupancyID)
					if err != nil {
						return fmt.Errorf("failed to run eligibility for account psr msg %s, occupancyID %s: %w", env.GetUuid(), occupancyID, err)
					}
				}
			}
		}
		return nil
	}
}

type psrIdentifier interface {
	GetAccountId() string
}
