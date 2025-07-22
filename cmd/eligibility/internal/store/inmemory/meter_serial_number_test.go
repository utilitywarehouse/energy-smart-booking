package inmemory_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store/inmemory"
)

func Test_FindMeterSerialNumber(t *testing.T) {

	msnStore, err := inmemory.NewMeterSerialNumber("./testdata/msn.csv")
	require.NoError(t, err)

	require.True(t, msnStore.FindMeterSerialNumber("GSM1031"))
	require.True(t, msnStore.FindMeterSerialNumber("ESM302020"))
	require.True(t, msnStore.FindMeterSerialNumber("L001"))
	require.False(t, msnStore.FindMeterSerialNumber("CANTSEEME"))

}
