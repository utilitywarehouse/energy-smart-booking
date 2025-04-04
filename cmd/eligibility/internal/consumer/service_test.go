package consumer_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	energy_entities "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/service/v1"
	"github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/consumer"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestServiceConsumer(t *testing.T) {
	s := store.NewService(pool)
	occupancyStore := store.NewOccupancy(pool)
	defer truncateDB(t)

	handler := consumer.HandleService(s, occupancyStore, nil, true)

	serviceEv1, err := makeMessage(&energy_entities.EnergyServiceEvent{
		Service: &energy_entities.Service{
			Service: &energy_entities.Service_Gas{
				Gas: &energy_entities.GasService{
					ServiceId:         "serviceID1",
					ServiceState:      energy_entities.ServiceState_SERVICE_STATE_REQUESTED,
					Mpxn:              "mpxn1",
					OccupancyId:       "occupancyID1",
					CustomerAccountId: "accountID1",
					IsLive:            false,
				},
			},
		},
		ServiceId: "serviceID1",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{serviceEv1})
	assert.NoError(t, err, "failed to handle service state event")

	service, err := s.Get(t.Context(), "serviceID1")
	assert.NoError(t, err, "failed to get service")
	expected := store.Service{
		ID:          "serviceID1",
		Mpxn:        "mpxn1",
		OccupancyID: "occupancyID1",
		SupplyType:  domain.SupplyTypeGas,
		IsLive:      false,
		StartDate:   nil,
		EndDate:     nil,
	}
	assert.Equal(t, expected, service, "service mismatch")

	startedDate := time.Date(2022, time.Month(12), 7, 0, 0, 0, 0, time.UTC)
	serviceEv2, err := makeMessage(&energy_entities.EnergyServiceEvent{
		Service: &energy_entities.Service{
			Service: &energy_entities.Service_Gas{
				Gas: &energy_entities.GasService{
					ServiceId:         "serviceID1",
					ServiceState:      energy_entities.ServiceState_SERVICE_STATE_STARTED,
					Mpxn:              "mpxn1",
					OccupancyId:       "occupancyID1",
					CustomerAccountId: "accountID1",
					IsLive:            true,
					StartDate:         timestamppb.New(startedDate),
				},
			},
		},
		ServiceId: "serviceID1",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{serviceEv2})
	assert.NoError(t, err, "failed to handle service state event")
	service, err = s.Get(t.Context(), "serviceID1")
	assert.NoError(t, err, "failed to get service")

	expected.StartDate = &startedDate
	expected.IsLive = true
	assert.Equal(t, expected, service, "service mismatch")
}
