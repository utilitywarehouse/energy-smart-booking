package consumer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	energy_entities "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/service/v1"
	"github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store/migrations"
	"github.com/utilitywarehouse/energy-smart-booking/internal/test_common"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestServiceConsumer(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)
	container, err := postgres.SetupTestContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(ctx)

	postgresURL, err := postgres.GetTestContainerDSN(container)
	if err != nil {
		t.Fatal(err)
	}

	pool, err := postgres.Setup(ctx, postgresURL, migrations.Source)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = postgres.Teardown(pool, migrations.Source); err != nil {
			t.Fatal(err)
		}
	}()
	s := store.NewService(pool)
	occupancyStore := store.NewOccupancy(pool)

	handler := HandleService(s, occupancyStore, nil, true)

	serviceEv1, err := test_common.MakeMessage(&energy_entities.EnergyServiceEvent{
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
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{serviceEv1})
	assert.NoError(err, "failed to handle service state event")

	service, err := s.Get(ctx, "serviceID1")
	assert.NoError(err, "failed to get service")
	expected := store.Service{
		ID:          "serviceID1",
		Mpxn:        "mpxn1",
		OccupancyID: "occupancyID1",
		SupplyType:  domain.SupplyTypeGas,
		IsLive:      false,
		StartDate:   nil,
		EndDate:     nil,
	}
	assert.Equal(expected, service, "service mismatch")

	startedDate := time.Date(2022, time.Month(12), 7, 0, 0, 0, 0, time.UTC)
	serviceEv2, err := test_common.MakeMessage(&energy_entities.EnergyServiceEvent{
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
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{serviceEv2})
	assert.NoError(err, "failed to handle service state event")
	service, err = s.Get(ctx, "serviceID1")
	assert.NoError(err, "failed to get service")

	expected.StartDate = &startedDate
	expected.IsLive = true
	assert.Equal(expected, service, "service mismatch")
}
