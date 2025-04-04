package consumer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	smart "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	"github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/consumer"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/uw-labs/substrate"
)

func TestAltHanConsumerElectricity(t *testing.T) {
	s := store.NewMeterpoint(pool)
	defer truncateDB(t)

	handler := consumer.HandleAltHan(s, nil, nil, true)

	altHanEv1, err := makeMessage(&smart.ElectricityAltHanMeterpointDiscoveredEvent{
		Mpan: "mpan1",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{altHanEv1})
	assert.NoError(t, err, "failed to handle alt han event")

	meterpoint, err := s.Get(t.Context(), "mpan1")
	assert.NoError(t, err, "failed to get meterpoint")
	expected := store.Meterpoint{
		Mpxn:         "mpan1",
		SupplyType:   domain.SupplyTypeElectricity,
		AltHan:       true,
		ProfileClass: platform.ProfileClass_PROFILE_CLASS_NONE,
		SSC:          "",
	}
	assert.Equal(t, expected, meterpoint, "meterpoint mismatch")

	altHanEv2, err := makeMessage(&smart.ElectricityAltHanMeterpointRemovedEvent{
		Mpan: "mpan1",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{altHanEv2})
	assert.NoError(t, err, "failed to handle alt han removed event")

	meterpoint, err = s.Get(t.Context(), "mpan1")
	assert.NoError(t, err, "failed to get meterpoint")
	expected.AltHan = false
	assert.Equal(t, expected, meterpoint, "meterpoint mismatch")
}

func TestAltHanConsumerGas(t *testing.T) {
	s := store.NewMeterpoint(pool)
	defer truncateDB(t)

	handler := consumer.HandleAltHan(s, nil, nil, true)

	altHanEv1, err := makeMessage(&smart.GasAltHanMeterpointRemovedEvent{
		Mprn: "mprn1",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{altHanEv1})
	assert.NoError(t, err, "failed to handle alt han event")

	meterpoint, err := s.Get(t.Context(), "mprn1")
	assert.NoError(t, err, "failed to get meterpoint")
	expected := store.Meterpoint{
		Mpxn:       "mprn1",
		SupplyType: domain.SupplyTypeGas,
		AltHan:     false,
	}
	assert.Equal(t, expected, meterpoint, "meterpoint mismatch")

	altHanEv2, err := makeMessage(&smart.GasAltHanMeterpointDiscoveredEvent{
		Mprn: "mprn1",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{altHanEv2})
	assert.NoError(t, err, "failed to handle alt han removed event")

	meterpoint, err = s.Get(t.Context(), "mprn1")
	assert.NoError(t, err, "failed to get meterpoint")
	expected.AltHan = true
	assert.Equal(t, expected, meterpoint, "meterpoint mismatch")
}
