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
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ServiceStore interface {
	Upsert(service models.Service)

	Begin()
	Commit(ctx context.Context) error
}

type ServiceStateHandler struct {
	store ServiceStore
}

func HandleServiceState(store ServiceStore) *ServiceStateHandler {
	return &ServiceStateHandler{store: store}
}

func (h *ServiceStateHandler) PreHandle(_ context.Context) error {
	h.store.Begin()
	return nil
}

func (h *ServiceStateHandler) PostHandle(ctx context.Context) error {
	return h.store.Commit(ctx)
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
		if err != nil {
			log.Infof("skipping service event, missing gas and electricity. event uuid: %s, service id: %s", eventUuid, ev.GetServiceId())
			return nil
		}

		h.store.Upsert(svc)
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

func extractService(energyEvent *energy_entities.EnergyServiceEvent) (models.Service, error) {
	if elec := energyEvent.GetService().GetElectricity(); elec != nil {
		return models.Service{
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
	if gas := energyEvent.GetService().GetGas(); gas != nil {
		return models.Service{
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
	return models.Service{}, fmt.Errorf("could not extract service information")
}
