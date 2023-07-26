package consumer

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated"
	energy_entities "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/service/v1"
	"github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/store"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ServiceStore interface {
	Upsert(ctx context.Context, service *store.Service) error
}

type ServiceStateHandler struct {
	store ServiceStore
}

func HandleServiceState(store ServiceStore) *ServiceStateHandler {
	return &ServiceStateHandler{store: store}
}

func (h *ServiceStateHandler) Handle(ctx context.Context, message substrate.Message) error {
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
		return fmt.Errorf("failed to unmarshall event in service state topic [%s|%s]: %w", eventUuid, env.Message.TypeUrl, err)
	}

	switch ev := payload.(type) {
	case *energy_entities.EnergyServiceEvent:
		svc, err := extractService(ev)
		err = h.store.Upsert(ctx, svc)
		if err != nil {
			return fmt.Errorf("failed to persist service state update for event [%s]: %w", eventUuid, err)
		}
	}
	return nil
}

func nullTimeForNullTimestamp(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := new(time.Time)
	*t = ts.AsTime()
	return t
}

func extractService(generic *energy_entities.EnergyServiceEvent) (*store.Service, error) {
	if elec := generic.GetService().GetElectricity(); elec != nil {
		return &store.Service{
			ServiceID:   elec.GetServiceId(),
			Mpxn:        elec.GetMpxn(),
			OccupancyID: elec.GetOccupancyId(),
			SupplyType:  domain.SupplyTypeElectricity,
			AccountID:   elec.GetCustomerAccountId(),
			StartDate:   nullTimeForNullTimestamp(elec.GetStartDate()),
			EndDate:     nullTimeForNullTimestamp(elec.GetEndDate()),
			IsLive:      elec.GetIsLive(),
		}, nil
	}
	if gas := generic.GetService().GetGas(); gas != nil {
		return &store.Service{
			ServiceID:   gas.GetServiceId(),
			Mpxn:        gas.GetMpxn(),
			OccupancyID: gas.GetOccupancyId(),
			SupplyType:  domain.SupplyTypeGas,
			AccountID:   gas.GetCustomerAccountId(),
			StartDate:   nullTimeForNullTimestamp(gas.GetStartDate()),
			EndDate:     nullTimeForNullTimestamp(gas.GetEndDate()),
			IsLive:      gas.GetIsLive(),
		}, nil
	}
	return nil, fmt.Errorf("could not extract service information")
}
