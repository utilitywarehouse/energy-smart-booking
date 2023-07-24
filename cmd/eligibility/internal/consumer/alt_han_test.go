package consumer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	"github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store/migrations"
	"github.com/utilitywarehouse/energy-smart-booking/internal/test_common"
	"github.com/uw-labs/substrate"
)

func TestAltHanConsumerElectricity(t *testing.T) {
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

	handler := HandleAltHan(s, nil, nil, true)

	altHanEv1, err := test_common.MakeMessage(&smart.ElectricityAltHanMeterpointDiscoveredEvent{
		Mpan: "mpan1",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{altHanEv1})
	assert.NoError(err, "failed to handle alt han event")

	meterpoint, err := s.Get(ctx, "mpan1")
	assert.NoError(err, "failed to get meterpoint")
	expected := store.Meterpoint{
		Mpxn:         "mpan1",
		SupplyType:   domain.SupplyTypeElectricity,
		AltHan:       true,
		ProfileClass: platform.ProfileClass_PROFILE_CLASS_NONE,
		SSC:          "",
	}
	assert.Equal(expected, meterpoint, "meterpoint mismatch")

	altHanEv2, err := test_common.MakeMessage(&smart.ElectricityAltHanMeterpointRemovedEvent{
		Mpan: "mpan1",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{altHanEv2})
	assert.NoError(err, "failed to handle alt han removed event")

	meterpoint, err = s.Get(ctx, "mpan1")
	assert.NoError(err, "failed to get meterpoint")
	expected.AltHan = false
	assert.Equal(expected, meterpoint, "meterpoint mismatch")
}

func TestAltHanConsumerGas(t *testing.T) {
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

	handler := HandleAltHan(s, nil, nil, true)

	altHanEv1, err := test_common.MakeMessage(&smart.GasAltHanMeterpointRemovedEvent{
		Mprn: "mprn1",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{altHanEv1})
	assert.NoError(err, "failed to handle alt han event")

	meterpoint, err := s.Get(ctx, "mprn1")
	assert.NoError(err, "failed to get meterpoint")
	expected := store.Meterpoint{
		Mpxn:       "mprn1",
		SupplyType: domain.SupplyTypeGas,
		AltHan:     false,
	}
	assert.Equal(expected, meterpoint, "meterpoint mismatch")

	altHanEv2, err := test_common.MakeMessage(&smart.GasAltHanMeterpointDiscoveredEvent{
		Mprn: "mprn1",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{altHanEv2})
	assert.NoError(err, "failed to handle alt han removed event")

	meterpoint, err = s.Get(ctx, "mprn1")
	assert.NoError(err, "failed to get meterpoint")
	expected.AltHan = true
	assert.Equal(expected, meterpoint, "meterpoint mismatch")
}
