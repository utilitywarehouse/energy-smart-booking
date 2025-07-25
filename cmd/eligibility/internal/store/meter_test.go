package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	energy_domain "github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-pkg/fabrication"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
)

func date(year, month, day int) time.Time {
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

	err := s.Add(ctx, &domain.Meter{
		ID:         meterID,
		Mpxn:       mpxn,
		MSN:        serialNumber,
		SupplyType: energy_domain.SupplyTypeElectricity,
		MeterType:  platform.MeterTypeElec_METER_TYPE_ELEC_S2AD.String(),
	})
	assert.NoError(err, "failed to add meter")

	err = s.InstallMeter(ctx, meterID, installationDate)
	assert.NoError(err, "failed to install meter")

	err = s.AddMeterType(ctx, meterID, platform.MeterTypeElec_METER_TYPE_ELEC_SMETS1.String())
	assert.NoError(err, "failed to update meter type")

	expected := domain.Meter{
		ID:         meterID,
		Mpxn:       mpxn,
		MSN:        serialNumber,
		SupplyType: energy_domain.SupplyTypeElectricity,
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
	assert.NoError(err)

	meter, err = s.Get(ctx, mpxn)
	assert.NoError(err, "failed to get reinstalled meter")
	assert.Equal(expected, meter, "mismatch")

	meterID1 := fabrication.NewMeterID(mpxn, serialNumber1)
	installationDate1 := date(2022, 2, 1)

	err = s.Add(ctx, &domain.Meter{
		ID:         meterID1,
		Mpxn:       mpxn,
		MSN:        serialNumber1,
		SupplyType: energy_domain.SupplyTypeElectricity,
	})
	assert.NoError(err, "failed to add meter")

	err = s.InstallMeter(ctx, meterID1, installationDate1)
	assert.NoError(err, "failed to install meter")

	// most recently installed meter should be returned
	meter, err = s.Get(ctx, mpxn)
	assert.NoError(err)

	expected = domain.Meter{
		ID:         meterID1,
		Mpxn:       mpxn,
		MSN:        serialNumber1,
		SupplyType: energy_domain.SupplyTypeElectricity,
	}
	assert.Equal(expected, meter, "mismatch")
}
