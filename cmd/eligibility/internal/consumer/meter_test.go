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
	"github.com/utilitywarehouse/energy-smart-booking/internal/test_common"
	"github.com/uw-labs/substrate"
)

func TestElectricityMeterConsumer(t *testing.T) {
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

	s := store.NewMeter(pool)

	handler := HandleMeter(s)

	meterEv1, err := test_common.MakeMessage(&platform.ElectricityMeterDiscoveredEvent{
		MeterId:           "meterID1",
		MeterSerialNumber: "msn1",
		MeterType:         platform.MeterTypeElec_METER_TYPE_ELEC_S2AD,
		Mpan:              "mpan1",
	})
	assert.NoError(err)
	meterEv2, err := test_common.MakeMessage(&platform.ElectricityMeterInstalledEvent{
		MeterId: "meterID1",
	})
	meterEv3, err := test_common.MakeMessage(&platform.ElectricityMeterTypeCorrectedEvent{
		MeterId:   "meterID1",
		MeterType: platform.MeterTypeElec_METER_TYPE_ELEC_S2ADE,
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{meterEv1, meterEv2, meterEv3})
	assert.NoError(err, "failed to handle meter events")

	meter, err := s.Get(ctx, "mpan1")
	assert.NoError(err, "failed to get meter")

	expected := store.Meter{
		ID:         "meterID1",
		Mpxn:       "mpan1",
		Msn:        "msn1",
		SupplyType: domain.SupplyTypeElectricity,
		Capacity:   nil,
		MeterType:  platform.MeterTypeElec_METER_TYPE_ELEC_S2ADE.String(),
	}
	assert.Equal(expected, meter, "meter mismatch")

	meterEv4, err := test_common.MakeMessage(&platform.ElectricityMeterUninstalledEvent{
		MeterId: "meterID1",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{meterEv4})
	assert.NoError(err, "failed to handle electricity meter uninstalled event")

	_, err = s.Get(ctx, "mpan1")
	assert.ErrorIs(err, store.ErrMeterNotFound)

	meterEv5, err := test_common.MakeMessage(&platform.ElectricityMeterErroneouslyUninstalledEvent{
		MeterId: "meterID1",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{meterEv5})
	assert.NoError(err, "failed to handle meter erroneously uninstalled event")

	meter, err = s.Get(ctx, "mpan1")
	assert.NoError(err, "failed to get meter")
	assert.Equal(expected, meter, "meter mismatch")
}

func TestGasMeterConsumer(t *testing.T) {
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

	s := store.NewMeter(pool)

	handler := HandleMeter(s)

	var capacity float32 = 11.55
	meterEv1, err := test_common.MakeMessage(&platform.GasMeterDiscoveredEvent{
		MeterId:           "meterID1",
		MeterSerialNumber: "msn1",
		MeterType:         platform.MeterTypeGas_METER_TYPE_GAS_PREPAYMENT,
		Mprn:              "mprn1",
		Capacity:          &capacity,
	})
	assert.NoError(err)
	meterEv2, err := test_common.MakeMessage(&platform.GasMeterInstalledEvent{
		MeterId: "meterID1",
	})
	meterEv3, err := test_common.MakeMessage(&platform.GasMeterTypeCorrectedEvent{
		MeterId:   "meterID1",
		MeterType: platform.MeterTypeGas_METER_TYPE_GAS_SMETS2,
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{meterEv1, meterEv2, meterEv3})
	assert.NoError(err, "failed to handle meter events")

	meter, err := s.Get(ctx, "mprn1")
	assert.NoError(err, "failed to get meter")

	expected := store.Meter{
		ID:         "meterID1",
		Mpxn:       "mprn1",
		Msn:        "msn1",
		SupplyType: domain.SupplyTypeGas,
		Capacity:   &capacity,
		MeterType:  platform.MeterTypeGas_METER_TYPE_GAS_SMETS2.String(),
	}
	assert.Equal(expected, meter, "meter mismatch")

	meterEv4, err := test_common.MakeMessage(&platform.GasMeterUninstalledEvent{
		MeterId: "meterID1",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{meterEv4})
	assert.NoError(err, "failed to handle electricity meter uninstalled event")

	_, err = s.Get(ctx, "mprn1")
	assert.ErrorIs(err, store.ErrMeterNotFound)

	meterEv5, err := test_common.MakeMessage(&platform.GasMeterErroneouslyUninstalledEvent{
		MeterId: "meterID1",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{meterEv5})
	assert.NoError(err, "failed to handle meter erroneously uninstalled event")

	meter, err = s.Get(ctx, "mprn1")
	assert.NoError(err, "failed to get meter")
	assert.Equal(expected, meter, "meter mismatch")
}
