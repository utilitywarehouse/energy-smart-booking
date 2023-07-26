package consumer

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type BookingReferenceStore interface {
	Add(ctx context.Context, mpxn, reference string) error
}

type BookingReferenceHandler struct {
	store BookingReferenceStore
}

func HandleBookingReference(store BookingReferenceStore) *BookingReferenceHandler {
	return &BookingReferenceHandler{
		store: store,
	}
}

func (h *BookingReferenceHandler) Handle(ctx context.Context, message substrate.Message) error {
	var env generated.Envelope
	if err := proto.Unmarshal(message.Data(), &env); err != nil {
		return err
	}

	if env.Message == nil {
		log.Info("skipping empty message")
		metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
		return nil
	}

	payload, err := env.Message.UnmarshalNew()
	if err != nil {
		return fmt.Errorf("error unmarshalling booking reference event [%s]: %w", env.Message.TypeUrl, err)
	}

	switch ev := payload.(type) {
	case *smart.BookingMpxnReferenceCreatedEvent:
		err = h.store.Add(ctx, ev.GetMpxn(), ev.GetReference())
		if err != nil {
			return fmt.Errorf("failed to persist booking reference for event [%s]: %w", env.GetUuid(), err)
		}
	}

	return nil
}
