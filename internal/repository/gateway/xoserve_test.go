//go:generate mockgen -source=xoserve.go -destination ./mocks/xoserve_mocks.go

package gateway_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	xoservev1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/xoserve/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
	mock_gateways "github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway/mocks"
	"go.uber.org/mock/gomock"
)

func Test_GetMPRNTechnicalDetails(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	mXOServe := mock_gateways.NewMockXOServeClient(ctrl)
	mai := fakeMachineAuthInjector{}
	mai.ctx = ctx

	myGw := gateway.NewXOServeGateway(mai, mXOServe)

	mXOServe.EXPECT().GetSwitchDataByMPRN(ctx, &xoservev1.SearchByMPRNRequest{
		Mprn: "mprn-1",
	}).Return(&xoservev1.TechnicalDetailsResponse{
		Meter: &xoservev1.MeterDetails{
			MeterType:     platform.MeterTypeGas_METER_TYPE_GAS_COIN,
			MeterCapacity: float32(6),
		},
	}, nil)

	actual := &models.GasMeterTechnicalDetails{
		MeterType: platform.MeterTypeGas_METER_TYPE_GAS_COIN,
		Capacity:  float32(6),
	}

	expected, err := myGw.GetMPRNTechnicalDetails(ctx, "mprn-1")
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, actual, cmpopts.IgnoreUnexported()) {
		t.Fatalf("expected: %+v, actual: %+v", expected, actual)
	}
}
