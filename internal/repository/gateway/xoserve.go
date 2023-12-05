package gateway

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	xoservev1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/xoserve/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type XOServeClient interface {
	GetSwitchDataByMPRN(context.Context, *xoservev1.SearchByMPRNRequest, ...grpc.CallOption) (*xoservev1.TechnicalDetailsResponse, error)
}

var xoserveAPIResponses = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "smart_booking_xoserve_response_total",
	Help: "The number of xoserve API error responses made by status code",
}, []string{"status"})

type XOServeGateway struct {
	mai    MachineAuthInjector
	client XOServeClient
}

func NewXOServeGateway(mai MachineAuthInjector, client XOServeClient) *XOServeGateway {
	return &XOServeGateway{mai, client}
}

func (gw *XOServeGateway) GetMPRNTechnicalDetails(ctx context.Context, mprn string) (*models.GasMeterTechnicalDetails, error) {
	technicalDetails, err := gw.client.GetSwitchDataByMPRN(gw.mai.ToCtx(ctx), &xoservev1.SearchByMPRNRequest{
		Mprn: mprn,
	})
	if err != nil {
		code := status.Convert(err).Code()
		xoserveAPIResponses.WithLabelValues(code.String()).Inc()
		return nil, fmt.Errorf("failed to get switch data by mprn: %s, %w", mprn, err)
	}

	return &models.GasMeterTechnicalDetails{
		MeterType: technicalDetails.GetMeter().GetMeterType(),
		Capacity:  technicalDetails.GetMeter().GetMeterCapacity(),
	}, nil
}
