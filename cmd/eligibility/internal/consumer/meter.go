package consumer

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	energy_contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	energy_domain "github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-pkg/substratemessage"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type MeterStore interface {
	Add(ctx context.Context, meter *domain.Meter) error
	InstallMeter(ctx context.Context, meterID string, at time.Time) error
	AddMeterType(ctx context.Context, meterID string, meterType string) error
	UninstallMeter(ctx context.Context, meterID string, at time.Time) error
	ReInstallMeter(ctx context.Context, meterID string) error
	AddMeterCapacity(ctx context.Context, meterID string, capacity float32) error
	GetMpxnByID(ctx context.Context, meterID string) (string, error)
}

type OccupancyMeterStore interface {
	GetIDsByMPXN(ctx context.Context, mpxn string) ([]string, error)
}

type Evaluator interface {
	RunFull(ctx context.Context, occupancyID string) error
}

func HandleMeter(s MeterStore, occupancyStore OccupancyMeterStore, evaluator Evaluator, stateRebuild bool) substratemessage.BatchHandlerFunc {
	return func(ctx context.Context, messages []substrate.Message) error {
		for _, msg := range messages {
			var env energy_contracts.Envelope
			if err := proto.Unmarshal(msg.Data(), &env); err != nil {
				return err
			}

			if env.Message == nil {
				log.Info("skipping empty meter message")
				metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
				continue
			}

			inner, err := env.Message.UnmarshalNew()
			if err != nil {
				return fmt.Errorf("error unmarshaling meter event [%s] %s: %w", env.GetUuid(), env.GetMessage().GetTypeUrl(), err)
			}

			switch x := inner.(type) {
			default:
				return nil
			case *platform.ElectricityMeterDiscoveredEvent:
				err = s.Add(ctx, &domain.Meter{
					ID:         x.GetMeterId(),
					Mpxn:       x.GetMpan(),
					MSN:        x.GetMeterSerialNumber(),
					SupplyType: energy_domain.SupplyTypeElectricity,
					MeterType:  x.GetMeterType().String(),
				})
			case *platform.ElectricityMeterTypeCorrectedEvent:
				err = s.AddMeterType(ctx, x.GetMeterId(), x.GetMeterType().String())

			case *platform.GasMeterDiscoveredEvent:
				err = s.Add(ctx, &domain.Meter{
					ID:         x.GetMeterId(),
					Mpxn:       x.GetMprn(),
					MSN:        x.GetMeterSerialNumber(),
					SupplyType: energy_domain.SupplyTypeGas,
					MeterType:  x.GetMeterType().String(),
				})
				if err != nil {
					return fmt.Errorf("failed to process meter event %s: %w", env.Uuid, err)
				}
				if x.Capacity != nil {
					err = s.AddMeterCapacity(ctx, x.GetMeterId(), x.GetCapacity())
				}
			case *platform.GasMeterTypeCorrectedEvent:
				err = s.AddMeterType(ctx, x.GetMeterId(), x.GetMeterType().String())
			case *platform.GasMeterCapacityChangedEvent:
				err = s.AddMeterCapacity(ctx, x.GetMeterId(), x.GetCapacity())

			case *platform.ElectricityMeterInstalledEvent, *platform.GasMeterInstalledEvent:
				//nolint return value not check on interface assertion
				err = s.InstallMeter(ctx, x.(meterIdentifier).GetMeterId(), env.OccurredAt.AsTime())
			case *platform.ElectricityMeterUninstalledEvent, *platform.GasMeterUninstalledEvent:
				//nolint return value not check on interface assertion
				err = s.UninstallMeter(ctx, x.(meterIdentifier).GetMeterId(), env.OccurredAt.AsTime())
			case *platform.ElectricityMeterErroneouslyUninstalledEvent, *platform.GasMeterErroneouslyUninstalledEvent:
				//nolint return value not check on interface assertion
				err = s.ReInstallMeter(ctx, x.(meterIdentifier).GetMeterId())

			}
			if err != nil {
				return fmt.Errorf("failed to process meter event %s: %w", env.Uuid, err)
			}

			if !stateRebuild {
				//nolint return value not check on interface assertion
				meterID := inner.(meterIdentifier).GetMeterId()
				mpxn, err := s.GetMpxnByID(ctx, meterID)
				if err != nil {
					return fmt.Errorf("failed to get meter mpxn for msg %s, meter ID %s: %w", env.GetUuid(), meterID, err)
				}
				occupanciesIDs, err := occupancyStore.GetIDsByMPXN(ctx, mpxn)
				if err != nil {
					return fmt.Errorf("failed to get occupancies for msg %s, mpxn %s: %w", env.GetUuid(), mpxn, err)
				}
				for _, occupancyID := range occupanciesIDs {
					err = evaluator.RunFull(ctx, occupancyID)
					if err != nil {
						return fmt.Errorf("failed to run evaluation for meter msg %s, occupancyID %s: %w", env.GetUuid(), occupancyID, err)
					}
				}
			}
		}
		return nil
	}
}

type meterIdentifier interface {
	GetMeterId() string
}
