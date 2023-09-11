package consumer

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type OccupancyStore interface {
	Insert(occupancy models.Occupancy)
	UpdateSiteID(occupancyID, siteID string)

	Begin()
	Commit(ctx context.Context) error
}

type OccupancyHandler struct {
	store OccupancyStore
}

func HandleOccupancy(store OccupancyStore) *OccupancyHandler {
	return &OccupancyHandler{
		store: store,
	}
}

func (h *OccupancyHandler) PreHandle(_ context.Context) error {
	h.store.Begin()
	return nil
}

func (h *OccupancyHandler) PostHandle(ctx context.Context) error {
	return h.store.Commit(ctx)
}

func (h *OccupancyHandler) Handle(_ context.Context, message substrate.Message) error {
	var env generated.Envelope
	if err := proto.Unmarshal(message.Data(), &env); err != nil {
		return err
	}

	eventUUID := env.Uuid
	if env.Message == nil {
		log.Infof("skipping empty message [%s]", eventUUID)
		metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
		return nil
	}

	payload, err := env.Message.UnmarshalNew()
	if err != nil {
		return fmt.Errorf("failed to unmarshall event in occupancy topic [%s|%s]: %w", eventUUID, env.Message.TypeUrl, err)
	}

	switch ev := payload.(type) {
	case *platform.OccupancyStartedEvent:

		occupancy := models.Occupancy{
			OccupancyID: ev.GetOccupancyId(),
			SiteID:      ev.GetSiteId(),
			AccountID:   ev.GetCustomerAccountId(),
			CreatedAt:   env.GetCreatedAt().AsTime(),
		}

		h.store.Insert(occupancy)

	case *platform.OccupancySiteCorrectedEvent:
		h.store.UpdateSiteID(ev.GetOccupancyId(), ev.GetSiteId())
	}

	if err != nil {
		return fmt.Errorf("failed to process occupancy event %s: %w", env.Uuid, err)
	}

	return nil
}
