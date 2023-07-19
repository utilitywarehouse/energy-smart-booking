package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-pkg/domain"
)

func TestMeterpoints(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	s := NewMeterpoint(connect(ctx))
	defer s.pool.Close()

	const mpxn = "mpxn1"

	err := s.AddProfileClass(ctx, mpxn, domain.SupplyTypeElectricity, platform.ProfileClass_PROFILE_CLASS_02)
	assert.NoError(err, "failed to add meterpoint profile class")

	err = s.AddAltHan(ctx, mpxn, domain.SupplyTypeElectricity, true)
	assert.NoError(err, mpxn, "failed to add alt han")

	expected := Meterpoint{
		Mpxn:         mpxn,
		SupplyType:   domain.SupplyTypeElectricity,
		ProfileClass: platform.ProfileClass_PROFILE_CLASS_02,
		SSC:          "",
		AltHan:       true,
	}
	meterpoint, err := s.Get(ctx, mpxn)
	assert.NoError(err, "failed to get meterpoint")
	assert.Equal(expected, meterpoint, "mismatch")

	err = s.AddSsc(ctx, mpxn, domain.SupplyTypeElectricity, "ssc")
	assert.NoError(err, "failed to update meterpoint ssc")
	expected.SSC = "ssc"

	meterpoint, err = s.Get(ctx, mpxn)
	assert.NoError(err, "failed to get meterpoint")
	assert.Equal(expected, meterpoint, "mismatch")
}
