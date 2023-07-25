package consumer

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	energy_contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	"github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-pkg/substratemessage"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type AltHanStore interface {
	AddAltHan(ctx context.Context, mpxn string, supplyType domain.SupplyType, altHan bool) error
}

type OccupancyAltHanStore interface {
	GetIDsByAccount(ctx context.Context, accountID string) ([]string, error)
}

type SuppliableEvaluator interface {
	RunSuppliability(ctx context.Context, occupancyID string) error
}

func HandleAltHan(store AltHanStore, occupancyStore OccupancyAltHanStore, evaluator SuppliableEvaluator, stateRebuild bool) substratemessage.BatchHandlerFunc {
	return func(ctx context.Context, messages []substrate.Message) error {
		for _, msg := range messages {
			var env energy_contracts.Envelope
			if err := proto.Unmarshal(msg.Data(), &env); err != nil {
				return err
			}

			if env.Message == nil {
				log.Info("skipping empty alt han message")
				metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
				continue
			}

			inner, err := env.Message.UnmarshalNew()
			if err != nil {
				return fmt.Errorf("error unmarshaling alt han event [%s] %s: %w", env.GetUuid(), env.GetMessage().GetTypeUrl(), err)
			}
			switch x := inner.(type) {
			case *smart.ElectricityAltHanMeterpointDiscoveredEvent:
				err = store.AddAltHan(ctx, x.GetMpan(), domain.SupplyTypeElectricity, true)
			case *smart.ElectricityAltHanMeterpointRemovedEvent:
				err = store.AddAltHan(ctx, x.GetMpan(), domain.SupplyTypeElectricity, false)
			case *smart.GasAltHanMeterpointDiscoveredEvent:
				err = store.AddAltHan(ctx, x.GetMprn(), domain.SupplyTypeGas, true)
			case *smart.GasAltHanMeterpointRemovedEvent:
				err = store.AddAltHan(ctx, x.GetMprn(), domain.SupplyTypeGas, false)
			}
			if err != nil {
				return fmt.Errorf("failed to process alt han event %s: %w", env.Uuid, err)
			}

			if !stateRebuild {
				accountID := inner.(altHanIdentifier).GetAccountId()
				occupanciesIDs, err := occupancyStore.GetIDsByAccount(ctx, accountID)
				if err != nil {
					return fmt.Errorf("failed to get occupancies for msg %s, account ID %s: %w", env.GetUuid(), accountID, err)
				}
				for _, occupancyID := range occupanciesIDs {
					err = evaluator.RunSuppliability(ctx, occupancyID)
					if err != nil {
						return fmt.Errorf("failed to run suppliability for alt han msg %s, occupancyID %s: %w", env.GetUuid(), occupancyID, err)
					}
				}
			}
		}
		return nil
	}
}

type altHanIdentifier interface {
	GetAccountId() string
}
