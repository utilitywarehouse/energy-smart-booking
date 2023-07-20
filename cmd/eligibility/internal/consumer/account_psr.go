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

type AccountPSRStore interface {
	Add(ctx context.Context, accountID string, codes []string) error
	Remove(ctx context.Context, accountID string) error
	GetPSRCodes(ctx context.Context, accountID string) ([]string, error)
}

func HandleAccountPSR(store AccountPSRStore) substratemessage.BatchHandlerFunc {
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
				if len(x.GetCodes()) == 0 {
					err = store.Remove(ctx, x.GetAccountId())
				} else {
					err = store.Add(ctx, x.AccountId, x.GetCodes())
				}
				if err != nil {
					return fmt.Errorf("failed to update account psr codes for account %s, event %s: %w", x.GetAccountId(), env.GetUuid(), err)
				}
			}
		}
		return nil
	}
}
