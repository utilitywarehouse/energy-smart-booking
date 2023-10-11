package gateway

import (
	"context"
	"fmt"

	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	xoservev1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/xoserve/v1"
	"google.golang.org/grpc"
)

type XOServeClient interface {
	GetSwitchDataByMPRN(context.Context, *xoservev1.SearchByMPRNRequest, ...grpc.CallOption) (*xoservev1.TechnicalDetailsResponse, error)
}

type XOServeGateway struct {
	client XOServeClient
}

func NewXOServeGateway(client XOServeClient) *XOServeGateway {
	return &XOServeGateway{client}
}

type GasMeterTechnicalDetails struct {
	MeterType platform.MeterTypeGas
	Capacity  float32
}

func (gw *XOServeGateway) GetMPRNTechnicalDetails(ctx context.Context, mprn string) (*GasMeterTechnicalDetails, error) {
	technicalDetails, err := gw.client.GetSwitchDataByMPRN(ctx, &xoservev1.SearchByMPRNRequest{
		Mprn: mprn,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get switch data by mprn: %s, %w", mprn, err)
	}

	return &GasMeterTechnicalDetails{
		MeterType: technicalDetails.GetMeter().GetMeterType(),
		Capacity:  technicalDetails.GetMeter().GetMeterCapacity(),
	}, nil
}
