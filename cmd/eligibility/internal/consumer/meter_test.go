package consumer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	energy_domain "github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/consumer"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/uw-labs/substrate"
)

func TestElectricityMeterConsumer(t *testing.T) {
	s := store.NewMeter(pool)
	defer truncateDB(t)

	handler := consumer.HandleMeter(s, nil, nil, true)

	meterEv1, err := makeMessage(&platform.ElectricityMeterDiscoveredEvent{
		MeterId:           "meterID1",
		MeterSerialNumber: "msn1",
		MeterType:         platform.MeterTypeElec_METER_TYPE_ELEC_S2AD,
		Mpan:              "mpan1",
	})
	assert.NoError(t, err)
	meterEv2, err := makeMessage(&platform.ElectricityMeterInstalledEvent{
		MeterId: "meterID1",
	})
	meterEv3, err := makeMessage(&platform.ElectricityMeterTypeCorrectedEvent{
		MeterId:   "meterID1",
		MeterType: platform.MeterTypeElec_METER_TYPE_ELEC_S2ADE,
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{meterEv1, meterEv2, meterEv3})
	assert.NoError(t, err, "failed to handle meter events")

	meter, err := s.Get(t.Context(), "mpan1")
	assert.NoError(t, err, "failed to get meter")

	expected := domain.Meter{
		ID:         "meterID1",
		Mpxn:       "mpan1",
		MSN:        "msn1",
		SupplyType: energy_domain.SupplyTypeElectricity,
		Capacity:   nil,
		MeterType:  platform.MeterTypeElec_METER_TYPE_ELEC_S2ADE.String(),
	}
	assert.Equal(t, expected, meter, "meter mismatch")

	meterEv4, err := makeMessage(&platform.ElectricityMeterUninstalledEvent{
		MeterId: "meterID1",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{meterEv4})
	assert.NoError(t, err, "failed to handle electricity meter uninstalled event")

	_, err = s.Get(t.Context(), "mpan1")
	assert.ErrorIs(t, err, store.ErrMeterNotFound)

	meterEv5, err := makeMessage(&platform.ElectricityMeterErroneouslyUninstalledEvent{
		MeterId: "meterID1",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{meterEv5})
	assert.NoError(t, err, "failed to handle meter erroneously uninstalled event")

	meter, err = s.Get(t.Context(), "mpan1")
	assert.NoError(t, err, "failed to get meter")
	assert.Equal(t, expected, meter, "meter mismatch")
}

func TestGasMeterConsumer(t *testing.T) {
	s := store.NewMeter(pool)
	defer truncateDB(t)

	handler := consumer.HandleMeter(s, nil, nil, true)

	var capacity float32 = 11.55
	meterEv1, err := makeMessage(&platform.GasMeterDiscoveredEvent{
		MeterId:           "meterID1",
		MeterSerialNumber: "msn1",
		MeterType:         platform.MeterTypeGas_METER_TYPE_GAS_PREPAYMENT,
		Mprn:              "mprn1",
		Capacity:          &capacity,
	})
	assert.NoError(t, err)
	meterEv2, err := makeMessage(&platform.GasMeterInstalledEvent{
		MeterId: "meterID1",
	})
	meterEv3, err := makeMessage(&platform.GasMeterTypeCorrectedEvent{
		MeterId:   "meterID1",
		MeterType: platform.MeterTypeGas_METER_TYPE_GAS_SMETS2,
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{meterEv1, meterEv2, meterEv3})
	assert.NoError(t, err, "failed to handle meter events")

	meter, err := s.Get(t.Context(), "mprn1")
	assert.NoError(t, err, "failed to get meter")

	expected := domain.Meter{
		ID:         "meterID1",
		Mpxn:       "mprn1",
		MSN:        "msn1",
		SupplyType: energy_domain.SupplyTypeGas,
		Capacity:   &capacity,
		MeterType:  platform.MeterTypeGas_METER_TYPE_GAS_SMETS2.String(),
	}
	assert.Equal(t, expected, meter, "meter mismatch")

	meterEv4, err := makeMessage(&platform.GasMeterUninstalledEvent{
		MeterId: "meterID1",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{meterEv4})
	assert.NoError(t, err, "failed to handle electricity meter uninstalled event")

	_, err = s.Get(t.Context(), "mprn1")
	assert.ErrorIs(t, err, store.ErrMeterNotFound)

	meterEv5, err := makeMessage(&platform.GasMeterErroneouslyUninstalledEvent{
		MeterId: "meterID1",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{meterEv5})
	assert.NoError(t, err, "failed to handle meter erroneously uninstalled event")

	meter, err = s.Get(t.Context(), "mprn1")
	assert.NoError(t, err, "failed to get meter")
	assert.Equal(t, expected, meter, "meter mismatch")

	meterEv6, err := makeMessage(&platform.GasMeterCapacityChangedEvent{
		MeterId:  "meterID1",
		Capacity: 45.6,
	})

	var newCap float32 = 45.6
	expected.Capacity = &newCap
	err = handler(t.Context(), []substrate.Message{meterEv6})
	assert.NoError(t, err, "failed to handle meter capacity changed event")

	meter, err = s.Get(t.Context(), "mprn1")
	assert.NoError(t, err, "failed to get meter")
	assert.Equal(t, expected, meter, "meter mismatch")
}
