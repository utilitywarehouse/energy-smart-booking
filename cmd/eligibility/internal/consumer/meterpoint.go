package consumer

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	energy_contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-pkg/substratemessage"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type MeterpointStore interface {
	AddProfileClass(ctx context.Context, mpxn string, supplyType domain.SupplyType, profileClass platform.ProfileClass) error
	AddSsc(ctx context.Context, mpxn string, supplyType domain.SupplyType, ssc string) error
}

func HandleMeterpoint(s MeterpointStore) substratemessage.BatchHandlerFunc {
	return func(ctx context.Context, messages []substrate.Message) error {
		for _, msg := range messages {
			var env energy_contracts.Envelope
			if err := proto.Unmarshal(msg.Data(), &env); err != nil {
				return err
			}

			if env.Message == nil {
				log.Info("skipping empty meterpoint message")
				metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
				continue
			}

			inner, err := env.Message.UnmarshalNew()
			if err != nil {
				return fmt.Errorf("error unmarshaling meterpoint event [%s] %s: %w", env.GetUuid(), env.GetMessage().GetTypeUrl(), err)
			}
			switch x := inner.(type) {
			case *platform.ElectricityMeterpointProfileClassChangedEvent:
				err = s.AddProfileClass(ctx, x.GetMpan(), domain.SupplyTypeElectricity, x.GetPc())
			case *platform.ElectricityMeterPointSSCChangedEvent:
				err = s.AddSsc(ctx, x.GetMpan(), domain.SupplyTypeElectricity, x.GetSsc())
			}
			if err != nil {
				return fmt.Errorf("failed to process meterpoint event %s: %w", env.Uuid, err)
			}

		}
		return nil
	}
}
