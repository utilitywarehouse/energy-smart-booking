package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-pkg/fabrication"
)

func date(year, month, day int) time.Time { //nolint:unparam
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func TestMeter(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	s := NewMeter(connect(ctx))
	defer s.pool.Close()

	const (
		mpxn          = "1234567890"
		serialNumber  = "meter-serial-number"
		serialNumber1 = "meter-serial-number-1"
	)

	meterID := fabrication.NewMeterID(mpxn, serialNumber)
	installationDate := date(2020, 2, 1)

	err := s.Add(ctx, &Meter{
		ID:         meterID,
		Mpxn:       mpxn,
		Msn:        serialNumber,
		SupplyType: domain.SupplyTypeElectricity,
		MeterType:  platform.MeterTypeElec_METER_TYPE_ELEC_S2AD.String(),
	})
	assert.NoError(err, "failed to add meter")

	err = s.InstallMeter(ctx, meterID, installationDate)
	assert.NoError(err, "failed to install meter")

	err = s.AddMeterType(ctx, meterID, platform.MeterTypeElec_METER_TYPE_ELEC_SMETS1.String())
	assert.NoError(err, "failed to update meter type")

	expected := Meter{
		ID:         meterID,
		Mpxn:       mpxn,
		Msn:        serialNumber,
		SupplyType: domain.SupplyTypeElectricity,
		MeterType:  platform.MeterTypeElec_METER_TYPE_ELEC_SMETS1.String(),
	}

	meter, err := s.Get(ctx, mpxn)
	assert.NoError(err, "failed to get meter")
	assert.Equal(expected, meter, "mismatch")

	// uninstall
	err = s.UninstallMeter(ctx, meterID, time.Now())
	assert.NoError(err, "failed to uninstall meter")

	// meter should not be gettable
	_, err = s.Get(ctx, mpxn)
	assert.ErrorIs(err, ErrMeterNotFound)

	// reinstall meter
	err = s.ReInstallMeter(ctx, meterID)
	meter, err = s.Get(ctx, mpxn)
	assert.NoError(err, "failed to get reinstalled meter")
	assert.Equal(expected, meter, "mismatch")

	meterID1 := fabrication.NewMeterID(mpxn, serialNumber1)
	installationDate1 := date(2022, 2, 1)

	err = s.Add(ctx, &Meter{
		ID:         meterID1,
		Mpxn:       mpxn,
		Msn:        serialNumber1,
		SupplyType: domain.SupplyTypeElectricity,
	})
	assert.NoError(err, "failed to add meter")

	err = s.InstallMeter(ctx, meterID1, installationDate1)
	assert.NoError(err, "failed to install meter")

	// most recently installed meter should be returned
	meter, err = s.Get(ctx, mpxn)
	expected = Meter{
		ID:         meterID1,
		Mpxn:       mpxn,
		Msn:        serialNumber1,
		SupplyType: domain.SupplyTypeElectricity,
	}
	assert.Equal(expected, meter, "mismatch")
}
