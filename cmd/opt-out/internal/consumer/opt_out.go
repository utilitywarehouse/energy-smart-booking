package consumer

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	energy_contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-pkg/substratemessage"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/opt-out/internal/store"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type OptOutAccountStore interface {
	Add(ctx context.Context, id, number, addedBy string) error
	Get(ctx context.Context, id string) (*store.Account, error)
	Remove(ctx context.Context, id string) error
}

func Handle(accountStore OptOutAccountStore) substratemessage.BatchHandlerFunc {
	return func(ctx context.Context, messages []substrate.Message) error {
		for _, msg := range messages {
			var env energy_contracts.Envelope
			if err := proto.Unmarshal(msg.Data(), &env); err != nil {
				return err
			}

			if env.Message == nil {
				logrus.Info("skipping empty meterpoint message")
				metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
				continue
			}

			inner, err := env.Message.UnmarshalNew()
			if err != nil {
				return fmt.Errorf("error unmarshaling meterpoint event [%s] %s: %w", env.GetUuid(), env.GetMessage().GetTypeUrl(), err)
			}
			switch x := inner.(type) {
			case *smart.AccountBookingOptOutAddedEvent:
				_, err = accountStore.Get(ctx, x.GetAccountId())
				if err != nil && !errors.Is(err, store.ErrAccountNotFound) {
					return fmt.Errorf("failed to check account %s: %w", x.GetAccountId(), err)
				}
				if err == nil {
					continue
				}
				err = accountStore.Add(ctx, x.GetAccountId(), x.GetAccountNumber(), x.GetAddedBy())
				if err != nil {
					return fmt.Errorf("failed to opt out account %s: %w", x.GetAccountId(), err)
				}
			case *smart.AccountBookingOptOutRemovedEvent:
				err = accountStore.Remove(ctx, x.GetAccountId())
				if err != nil {
					return fmt.Errorf("failed to remove booking opt out for account %s: %w", x.GetAccountId(), err)
				}
			default:
				// no nothing
			}
		}
		return nil
	}
}
