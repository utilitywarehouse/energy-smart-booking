package consumer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store/migrations"
	"github.com/utilitywarehouse/energy-smart-booking/internal/testcommon"
	"github.com/uw-labs/substrate"
)

func TestMeterpointConsumer(t *testing.T) {
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
	s := store.NewMeterpoint(pool)

	handler := HandleMeterpoint(s, nil, nil, true)

	meterpointEv1, err := testcommon.MakeMessage(&platform.ElectricityMeterpointProfileClassChangedEvent{
		Mpan: "mpan1",
		Pc:   platform.ProfileClass_PROFILE_CLASS_01,
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{meterpointEv1})
	assert.NoError(err, "failed to handle meterpoint profile class changed event")

	meterpoint, err := s.Get(ctx, "mpan1")
	assert.NoError(err, "failed to get meterpoint")

	expected := store.Meterpoint{
		Mpxn:         "mpan1",
		SupplyType:   domain.SupplyTypeElectricity,
		AltHan:       false,
		ProfileClass: platform.ProfileClass_PROFILE_CLASS_01,
		SSC:          "",
	}
	assert.Equal(expected, meterpoint, "meterpoint mismatch")

	meterpointEv2, err := testcommon.MakeMessage(&platform.ElectricityMeterPointSSCChangedEvent{
		Mpan: "mpan1",
		Ssc:  "ssc",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{meterpointEv2})
	assert.NoError(err, "failed to handle meterpoint profile class changed event")

	meterpoint, err = s.Get(ctx, "mpan1")
	assert.NoError(err, "failed to get meterpoint")

	expected.SSC = "ssc"
	assert.Equal(expected, meterpoint, "meterpoint mismatch")
}
