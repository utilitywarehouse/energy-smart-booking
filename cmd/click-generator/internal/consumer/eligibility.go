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

type EligibilityEvaluationStore interface {
	UpsertEligibility(ctx context.Context, accountID, occupancyID string, eligible bool) error
}

type LinkStore interface {
	Remove(ctx context.Context, accountID, occupancyID string) error
}

type EligibilityHandler struct {
	evaluationStore EligibilityEvaluationStore
	linkStore       LinkStore
}

func NewEligibility(evaluationStore EligibilityEvaluationStore, linkStore LinkStore) *EligibilityHandler {
	return &EligibilityHandler{
		evaluationStore: evaluationStore,
		linkStore:       linkStore,
	}
}

func (h *EligibilityHandler) Handle(ctx context.Context, message substrate.Message) error {
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
		return fmt.Errorf("failed to unmarshall eligibility event [%s|%s]: %w", eventUuid, env.Message.TypeUrl, err)
	}

	switch ev := payload.(type) {
	case *smart.EligibleOccupancyAddedEvent:
		err = h.evaluationStore.UpsertEligibility(ctx, ev.AccountId, ev.OccupancyId, true)
	case *smart.EligibleOccupancyRemovedEvent:
		err = h.evaluationStore.UpsertEligibility(ctx, ev.AccountId, ev.OccupancyId, false)
		if err != nil {
			return fmt.Errorf("failed to upsert eligibility for event %s: %w", env.Uuid, err)
		}
		err = h.linkStore.Remove(ctx, ev.AccountId, ev.OccupancyId)
	}
	if err != nil {
		return fmt.Errorf("failed to process eligibility event %s: %w", env.Uuid, err)
	}

	return err
}
