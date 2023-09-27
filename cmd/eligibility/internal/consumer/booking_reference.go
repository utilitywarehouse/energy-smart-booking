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

type BookingRefStore interface {
	Add(ctx context.Context, mpxn, reference string) error
	Remove(ctx context.Context, mpxn string) error
	GetReference(ctx context.Context, mpxn string) (string, error)
}

type OccupancyBookingRefStore interface {
	GetIDsByMPXN(ctx context.Context, mpxn string) ([]string, error)
}

func HandleBookingRef(store BookingRefStore, occupancyStore OccupancyBookingRefStore, evaluator Evaluator, stateRebuild bool) substratemessage.BatchHandlerFunc {
	return func(ctx context.Context, messages []substrate.Message) error {
		for _, msg := range messages {
			var env energy_contracts.Envelope
			if err := proto.Unmarshal(msg.Data(), &env); err != nil {
				return err
			}

			if env.Message == nil {
				log.Info("skipping empty booking ref message")
				metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
				continue
			}

			inner, err := env.Message.UnmarshalNew()
			if err != nil {
				return fmt.Errorf("error unmarshaling booking ref event [%s] %s: %w", env.GetUuid(), env.GetMessage().GetTypeUrl(), err)
			}
			switch x := inner.(type) {
			case *smart.BookingMpxnReferenceCreatedEvent:
				err = store.Add(ctx, x.GetMpxn(), x.GetReference())
			case *smart.BookingMpxnReferenceRemovedEvent:
				err = store.Remove(ctx, x.GetMpxn())
			}
			if err != nil {
				return fmt.Errorf("failed to process booking ref event %s: %w", env.GetUuid(), err)
			}

			if !stateRebuild {
				mpxn := inner.(refIdentifier).GetMpxn()
				occupanciesIDs, err := occupancyStore.GetIDsByMPXN(ctx, mpxn)
				if err != nil {
					return fmt.Errorf("failed to get occupancies for msg %s, mpxn %s: %w", env.GetUuid(), mpxn, err)
				}
				for _, occupancyID := range occupanciesIDs {
					err = evaluator.RunFull(ctx, occupancyID)
					if err != nil {
						return fmt.Errorf("failed to run evaluation for meter msg %s, occupancyID %s: %w", env.GetUuid(), occupancyID, err)
					}
				}
			}
		}

		return nil
	}
}

type refIdentifier interface {
	GetMpxn() string
}
