package gateway

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	ecoesv2 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/ecoes/v2"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EcoesClient interface {
	GetTechnicalDetailsByMPAN(context.Context, *ecoesv2.GetTechnicalDetailsByMPANRequest, ...grpc.CallOption) (*ecoesv2.GetTechnicalDetailsByMPANResponse, error)
	GetRelatedMPANs(context.Context, *ecoesv2.GetRelatedMPANsRequest, ...grpc.CallOption) (*ecoesv2.GetRelatedMPANsResponse, error)
}

var ecoesAPIResponses = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "smart_booking_ecoes_response_total",
	Help: "The number of ecoes API error responses made by status code",
}, []string{"status"})

type EcoesGateway struct {
	client EcoesClient
}

func NewEcoesGateway(client EcoesClient) *EcoesGateway {
	return &EcoesGateway{client}
}

func (gw *EcoesGateway) GetMPANTechnicalDetails(ctx context.Context, mpan string) (*models.ElectricityMeterTechnicalDetails, error) {

	res, err := gw.client.GetTechnicalDetailsByMPAN(ctx, &ecoesv2.GetTechnicalDetailsByMPANRequest{
		Mpan: mpan,
	})
	if err != nil {
		code := status.Convert(err).Code()
		ecoesAPIResponses.WithLabelValues(code.String()).Inc()
		return nil, fmt.Errorf("failed to get technical details by mpan: %s, %w", mpan, err)
	}
	technicalDetails := res.GetTechnicalDetails()

	meters := []models.ElectricityMeter{}

	for _, elem := range technicalDetails.GetMeters() {
		// skip meters that have not been installed
		if elem.GetMeterInstalledDate() == nil {
			continue
		}
		meters = append(meters, models.ElectricityMeter{
			MeterType:   elem.MeterType,
			InstalledAt: time.Date(int(elem.GetMeterInstalledDate().GetYear()), time.Month(elem.GetMeterInstalledDate().Month), int(elem.GetMeterInstalledDate().GetDay()), 0, 0, 0, 0, time.UTC),
		})
	}

	sort.Slice(meters, func(i, j int) bool {
		return meters[i].InstalledAt.After(meters[j].InstalledAt)
	})

	return &models.ElectricityMeterTechnicalDetails{
		ProfileClass:                    technicalDetails.GetProfileClass(),
		SettlementStandardConfiguration: technicalDetails.GetStandardSettlementConfiguration(),
		Meters:                          meters,
	}, nil
}

func (gw *EcoesGateway) HasRelatedMPAN(ctx context.Context, mpan string) (bool, error) {

	relatedMPAN, err := gw.client.GetRelatedMPANs(ctx, &ecoesv2.GetRelatedMPANsRequest{
		Mpan: mpan,
	})
	if err != nil {
		code := status.Convert(err).Code()
		if code == codes.NotFound {
			return false, nil
		}
		ecoesAPIResponses.WithLabelValues(code.String()).Inc()
		return false, fmt.Errorf("failed to get related mpan: %s, %w", mpan, err)
	}

	return len(relatedMPAN.GetRelationships()) > 0, nil
}
