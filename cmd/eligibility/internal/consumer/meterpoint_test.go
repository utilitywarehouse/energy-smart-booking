package consumer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/consumer"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/uw-labs/substrate"
)

func TestMeterpointConsumer(t *testing.T) {
	s := store.NewMeterpoint(pool)
	defer truncateDB(t)

	handler := consumer.HandleMeterpoint(s, nil, nil, true)

	meterpointEv1, err := makeMessage(&platform.ElectricityMeterpointProfileClassChangedEvent{
		Mpan: "mpan1",
		Pc:   platform.ProfileClass_PROFILE_CLASS_01,
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{meterpointEv1})
	assert.NoError(t, err, "failed to handle meterpoint profile class changed event")

	meterpoint, err := s.Get(t.Context(), "mpan1")
	assert.NoError(t, err, "failed to get meterpoint")

	expected := store.Meterpoint{
		Mpxn:         "mpan1",
		SupplyType:   domain.SupplyTypeElectricity,
		AltHan:       false,
		ProfileClass: platform.ProfileClass_PROFILE_CLASS_01,
		SSC:          "",
	}
	assert.Equal(t, expected, meterpoint, "meterpoint mismatch")

	meterpointEv2, err := makeMessage(&platform.ElectricityMeterPointSSCChangedEvent{
		Mpan: "mpan1",
		Ssc:  "ssc",
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{meterpointEv2})
	assert.NoError(t, err, "failed to handle meterpoint profile class changed event")

	meterpoint, err = s.Get(t.Context(), "mpan1")
	assert.NoError(t, err, "failed to get meterpoint")

	expected.SSC = "ssc"
	assert.Equal(t, expected, meterpoint, "meterpoint mismatch")
}
