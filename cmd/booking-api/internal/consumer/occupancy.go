package consumer

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type OccupancyStore interface {
	Add(ctx context.Context, occupancyID, siteID, accountID string) error
	UpdateSite(ctx context.Context, occupancyID, siteID string) error
}

type OccupancyHandler struct {
	store OccupancyStore
}

func HandleOccupancy(store OccupancyStore) *OccupancyHandler {
	return &OccupancyHandler{
		store: store,
	}
}

func (h *OccupancyHandler) Handle(ctx context.Context, message substrate.Message) error {
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
		return fmt.Errorf("failed to unmarshall event in occupancy topic [%s|%s]: %w", eventUuid, env.Message.TypeUrl, err)
	}

	switch ev := payload.(type) {
	case *platform.OccupancyStartedEvent:
		err = h.store.Add(ctx, ev.GetOccupancyId(), ev.GetSiteId(), ev.GetCustomerAccountId())
	case *platform.OccupancySiteCorrectedEvent:
		err = h.store.UpdateSite(ctx, ev.GetOccupancyId(), ev.GetSiteId())
	}
	if err != nil {
		return fmt.Errorf("failed to process occupancy event %s: %w", env.Uuid, err)
	}

	return nil
}
