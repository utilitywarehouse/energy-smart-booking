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

type SiteStore interface {
	Add(ctx context.Context, siteID, postcode string) error
}

type SiteHandler struct {
	store SiteStore
}

func HandleSite(store SiteStore) *SiteHandler {
	return &SiteHandler{store: store}
}

func (h *SiteHandler) Handle(ctx context.Context, message substrate.Message) error {
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
		return fmt.Errorf("failed to unmarshall event in site topic [%s|%s]: %w", eventUuid, env.Message.TypeUrl, err)
	}

	switch ev := payload.(type) {
	case *platform.SiteDiscoveredEvent:
		{
			address := ev.GetAddress()
			if address == nil {
				log.Infof("skip event [%s] for site [%s]: empty address", eventUuid, ev.GetSiteId())
				return nil
			}

			postcode := address.GetPostcode()
			if postcode == "" {
				log.Infof("skip event [%s] for site [%s]: empty post code", eventUuid, ev.GetSiteId())
				return nil
			}
			err = h.store.Add(ctx, ev.GetSiteId(), postcode)
			if err != nil {
				return fmt.Errorf("failed to process site event %s: %w", eventUuid, err)
			}
		}
	}
	return nil
}
