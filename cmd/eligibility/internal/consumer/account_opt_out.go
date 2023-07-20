package consumer

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	energy_contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-pkg/substratemessage"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type AccountOptOutStore interface {
	AddOptOut(ctx context.Context, accountID string, optOut bool) error
}

func HandleAccountOptOut(store AccountOptOutStore) substratemessage.BatchHandlerFunc {
	return func(ctx context.Context, messages []substrate.Message) error {
		for _, msg := range messages {
			var env energy_contracts.Envelope
			if err := proto.Unmarshal(msg.Data(), &env); err != nil {
				return err
			}

			if env.Message == nil {
				log.Info("skipping empty account opt out message")
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
		}
		return nil
	}
}
