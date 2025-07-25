package consumer

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/utilitywarehouse/energy-contracts/pkg/generated"
	smart_contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type OccupancyEligibleStore interface {
	Upsert(occupancy models.OccupancyEligibility)
	Delete(occupancy models.OccupancyEligibility)

	Begin()
	Commit(context.Context) error
}

type OccupancyEligibleHandler struct {
	store OccupancyEligibleStore
}

func NewOccupancyEligibleHandler(store OccupancyEligibleStore) *OccupancyEligibleHandler {
	return &OccupancyEligibleHandler{store: store}
}

func (h *OccupancyEligibleHandler) PreHandle(_ context.Context) error {
	h.store.Begin()
	return nil
}

func (h *OccupancyEligibleHandler) PostHandle(ctx context.Context) error {
	return h.store.Commit(ctx)
}

func (h *OccupancyEligibleHandler) Handle(_ context.Context, message substrate.Message) error {
	var env generated.Envelope
	if err := proto.Unmarshal(message.Data(), &env); err != nil {
		return err
	}

	eventUUID := env.Uuid
	if env.Message == nil {
		slog.Info("skipping empty message", "event_uuid", eventUUID)
		metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
		return nil
	}

	payload, err := env.Message.UnmarshalNew()
	if err != nil {
		return fmt.Errorf("failed to unmarshall event in booking topic [%s|%s]: %w", eventUUID, env.Message.TypeUrl, err)
	}
	switch ev := payload.(type) {
	case *smart_contracts.SmartBookingJourneyOccupancyAddedEvent:

		occupancyEligibility := models.OccupancyEligibility{
			OccupancyID: ev.GetOccupancyId(),
			Reference:   ev.GetReference(),
		}

		h.store.Upsert(occupancyEligibility)

	case *smart_contracts.SmartBookingJourneyOccupancyRemovedEvent:

		occupancyEligibility := models.OccupancyEligibility{
			OccupancyID: ev.GetOccupancyId(),
		}

		h.store.Delete(occupancyEligibility)
	}

	return nil
}
