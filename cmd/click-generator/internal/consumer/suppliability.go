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

type SuppliabilityEvaluationStore interface {
	UpsertSuppliability(ctx context.Context, accountID, occupancyID string, suppliable bool) error
}

type SuppliabilityHandler struct {
	evaluationStore SuppliabilityEvaluationStore
	linkStore       LinkStore
}

func NewSuppliability(evaluationStore SuppliabilityEvaluationStore, linkStore LinkStore) *SuppliabilityHandler {
	return &SuppliabilityHandler{
		evaluationStore: evaluationStore,
		linkStore:       linkStore,
	}
}

func (h *SuppliabilityHandler) Handle(ctx context.Context, message substrate.Message) error {
	var env generated.Envelope
	if err := proto.Unmarshal(message.Data(), &env); err != nil {
		return err
	}

	eventUuid := env.Uuid
	if env.Message == nil {
		log.Infof("skipping empty message [%s]", eventUuid)
		metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
		return nil
	}

	payload, err := env.Message.UnmarshalNew()
	if err != nil {
		return fmt.Errorf("failed to unmarshall suppliability event [%s|%s]: %w", eventUuid, env.Message.TypeUrl, err)
	}

	switch ev := payload.(type) {
	case *smart.SuppliableOccupancyAddedEvent:
		err = h.evaluationStore.UpsertSuppliability(ctx, ev.AccountId, ev.OccupancyId, true)
	case *smart.SuppliableOccupancyRemovedEvent:
		err = h.evaluationStore.UpsertSuppliability(ctx, ev.AccountId, ev.OccupancyId, false)
		if err != nil {
			return fmt.Errorf("failed to upsert suppliability for event %s: %w", env.Uuid, err)
		}
		err = h.linkStore.Remove(ctx, ev.AccountId, ev.OccupancyId)
	}
	if err != nil {
		return fmt.Errorf("failed to process suppliability event %s: %w", env.Uuid, err)
	}

	return err
}
