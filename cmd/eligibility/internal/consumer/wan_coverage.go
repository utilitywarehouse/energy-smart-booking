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

type PostcodeStore interface {
	AddWanCoverage(ctx context.Context, postCode string, covered bool) error
}

type OccupancyPostcodeStore interface {
	GetIDsByPostcode(ctx context.Context, postCode string) ([]string, error)
}

func HandleWanCoverage(store PostcodeStore, occupancyStore OccupancyPostcodeStore, evaluator Evaluator, stateRebuild bool) substratemessage.BatchHandlerFunc {
	return func(ctx context.Context, messages []substrate.Message) error {
		for _, msg := range messages {
			var env energy_contracts.Envelope
			if err := proto.Unmarshal(msg.Data(), &env); err != nil {
				return err
			}

			if env.Message == nil {
				slog.Info("skipping empty wan coverage message")
				metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
				continue
			}

			inner, err := env.Message.UnmarshalNew()
			if err != nil {
				return fmt.Errorf("error unmarshaling wan coverage event [%s] %s: %w", env.GetUuid(), env.GetMessage().GetTypeUrl(), err)
			}
			switch x := inner.(type) {
			case *smart.WanCoverageAtPostcodeStartedEvent:
				err = store.AddWanCoverage(ctx, x.GetPostcode(), true)
			case *smart.WanCoverageAtPostcodeEndedEvent:
				err = store.AddWanCoverage(ctx, x.GetPostcode(), false)
			}
			if err != nil {
				return fmt.Errorf("failed to process wan coverage event %s: %w", env.Uuid, err)
			}

			if !stateRebuild {
				postCode := inner.(wanCoverageIdentifier).GetPostcode()
				occupanciesIDs, err := occupancyStore.GetIDsByPostcode(ctx, postCode)
				if err != nil {
					return fmt.Errorf("failed to get occupancies for msg %s, postCode %s: %w", env.GetUuid(), postCode, err)
				}
				for _, occupancyID := range occupanciesIDs {
					err = evaluator.RunFull(ctx, occupancyID)
					if err != nil {
						return fmt.Errorf("failed to run evaluation for wan converage msg %s, occupancyID %s: %w", env.GetUuid(), occupancyID, err)
					}
				}
			}
		}
		return nil
	}
}

type wanCoverageIdentifier interface {
	GetPostcode() string
}
