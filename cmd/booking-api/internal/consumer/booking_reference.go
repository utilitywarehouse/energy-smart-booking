package consumer

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type BookingReferenceStore interface {
	Upsert(bookingReference models.BookingReference)

	Begin()
	Commit(ctx context.Context) error
}

type BookingReferenceHandler struct {
	store BookingReferenceStore
}

func HandleBookingReference(store BookingReferenceStore) *BookingReferenceHandler {
	return &BookingReferenceHandler{
		store: store,
	}
}

func (h *BookingReferenceHandler) PreHandle(_ context.Context) error {
	h.store.Begin()
	return nil
}

func (h *BookingReferenceHandler) PostHandle(ctx context.Context) error {
	return h.store.Commit(ctx)
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

		bookingReference := models.BookingReference{
			Reference: ev.GetReference(),
			MPXN:      ev.GetMpxn(),
		}

		h.store.Upsert(bookingReference)
	}

	return nil
}
