package consumer

import (
	"context"
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	energy_contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	energy_entities "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/service/v1"
	"github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-pkg/substratemessage"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ServiceStore interface {
	Add(ctx context.Context, service *store.Service) error
	Get(ctx context.Context, serviceID string) (store.Service, error)
	AddStartDate(ctx context.Context, serviceID string, at time.Time) error
	AddEndDate(ctx context.Context, serviceID string, at time.Time) error
}

type ServiceOccupancyStore interface {
	Get(ctx context.Context, occupancyID string) (store.Occupancy, error)
}

func HandleService(s ServiceStore, occupancyStore ServiceOccupancyStore, evaluator Evaluator, stateRebuild bool) substratemessage.BatchHandlerFunc {
	return func(ctx context.Context, messages []substrate.Message) error {
		for _, msg := range messages {
			var env energy_contracts.Envelope
			if err := proto.Unmarshal(msg.Data(), &env); err != nil {
				return err
			}

			if env.Message == nil {
				log.Info("skipping empty service message")
				metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
				continue
			}

			inner, err := env.Message.UnmarshalNew()
			if err != nil {
				return fmt.Errorf("error unmarshaling service event [%s] %s: %w", env.GetUuid(), env.GetMessage().GetTypeUrl(), err)
			}
			switch x := inner.(type) {
			case *energy_entities.EnergyServiceEvent:
				var supplyType domain.SupplyType
				var service servicer

				if x.GetService().GetGas() != nil {
					service = x.GetService().GetGas()
					supplyType = domain.SupplyTypeGas
				} else {
					service = x.GetService().GetGas()
					supplyType = domain.SupplyTypeElectricity
				}

				err = persistService(ctx, s, supplyType, service)
				if err != nil {
					return fmt.Errorf("failed to persist service for service event %s: %w", env.GetUuid(), err)
				}

				if !stateRebuild {
					occupancyID := service.(servicer).GetOccupancyId()
					if occupancyID != "" {
						// make sure occupancy exists by the time we are informed of service
						_, err = occupancyStore.Get(ctx, occupancyID)
						if err != nil && !errors.Is(err, store.ErrOccupancyNotFound) {
							return fmt.Errorf("failed to query occupancy for service state msg %s, occupancy id %s: %w", env.GetUuid(), occupancyID, err)
						}
						if err == nil {
							err = evaluator.RunFull(ctx, occupancyID)
							if err != nil {
								return fmt.Errorf("failed to run evaluation for service state msg %s, occupancyID %s: %w", env.GetUuid(), occupancyID, err)
							}
						}
					}
				}
			}
		}
		return nil
	}
}

func persistService(ctx context.Context, s ServiceStore, supplyType domain.SupplyType, service servicer) error {
	err := s.Add(ctx, &store.Service{
		ID:          service.GetServiceId(),
		Mpxn:        service.GetMpxn(),
		OccupancyID: service.GetOccupancyId(),
		SupplyType:  supplyType,
		IsLive:      service.GetIsLive(),
	})
	if err != nil {
		return err
	}

	if service.GetStartDate() != nil {
		err = s.AddStartDate(ctx, service.GetServiceId(), service.GetStartDate().AsTime())
		if err != nil {
			return err
		}
	}
	if service.GetEndDate() != nil {
		err = s.AddEndDate(ctx, service.GetServiceId(), service.GetEndDate().AsTime())
		if err != nil {
			return err
		}
	}

	return nil
}

type servicer interface {
	GetServiceId() string
	GetIsLive() bool
	GetOccupancyId() string
	GetMpxn() string
	GetStartDate() *timestamppb.Timestamp
	GetEndDate() *timestamppb.Timestamp
}
