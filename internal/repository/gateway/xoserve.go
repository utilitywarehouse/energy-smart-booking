package gateway

import (
	"context"
	"fmt"

	xoservev1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/xoserve/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
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

func (gw *XOServeGateway) GetMPRNTechnicalDetails(ctx context.Context, mprn string) (*models.GasMeterTechnicalDetails, error) {
	technicalDetails, err := gw.client.GetSwitchDataByMPRN(ctx, &xoservev1.SearchByMPRNRequest{
		Mprn: mprn,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get switch data by mprn: %s, %w", mprn, err)
	}

	return &models.GasMeterTechnicalDetails{
		MeterType: technicalDetails.GetMeter().GetMeterType(),
		Capacity:  technicalDetails.GetMeter().GetMeterCapacity(),
	}, nil
}
