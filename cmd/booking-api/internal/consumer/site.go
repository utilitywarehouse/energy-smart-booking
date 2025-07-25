package consumer

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/utilitywarehouse/energy-contracts/pkg/generated"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type SiteStore interface {
	Upsert(site models.Site)

	Begin()
	Commit(ctx context.Context) error
}

type SiteHandler struct {
	store SiteStore
}

func HandleSite(store SiteStore) *SiteHandler {
	return &SiteHandler{store: store}
}

func (h *SiteHandler) PreHandle(_ context.Context) error {
	h.store.Begin()
	return nil
}

func (h *SiteHandler) PostHandle(ctx context.Context) error {
	return h.store.Commit(ctx)
}

func (h *SiteHandler) Handle(_ context.Context, message substrate.Message) error {
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
		return fmt.Errorf("failed to unmarshall event in site topic [%s|%s]: %w", eventUUID, env.Message.TypeUrl, err)
	}

	switch ev := payload.(type) {
	case *platform.SiteDiscoveredEvent:
		{
			address := ev.GetAddress()
			if address == nil {
				slog.Info("skip event, empty address", "event_uuid", eventUUID, "site_id", ev.GetSiteId())
				return nil
			}

			site := models.Site{
				SiteID:                  ev.GetSiteId(),
				Postcode:                address.GetPostcode(),
				UPRN:                    address.GetUprn(),
				BuildingNameNumber:      address.GetBuildingNameNumber(),
				DependentThoroughfare:   address.GetDependentThoroughfare(),
				Thoroughfare:            address.GetThoroughfare(),
				DoubleDependentLocality: address.GetDoubleDependentLocality(),
				DependentLocality:       address.GetDependentLocality(),
				Locality:                address.GetLocality(),
				County:                  address.GetCounty(),
				Town:                    address.GetTown(),
				Department:              address.GetDepartment(),
				Organisation:            address.GetOrganisation(),
				PoBox:                   address.GetPoBox(),
				DeliveryPointSuffix:     address.GetDeliveryPointSuffix(),
				SubBuildingNameNumber:   address.GetSubBuildingNameNumber(),
			}

			h.store.Upsert(site)
		}
	}
	return nil
}
