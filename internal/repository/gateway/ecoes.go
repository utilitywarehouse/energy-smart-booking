package gateway

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	ecoesv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/ecoes/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type EcoesClient interface {
	GetTechnicalDetailsByMPAN(context.Context, *ecoesv1.SearchByMPANRequest, ...grpc.CallOption) (*ecoesv1.TechnicalDetailsResponse, error)
	GetRelatedMPANs(context.Context, *ecoesv1.SearchByMPANRequest, ...grpc.CallOption) (*ecoesv1.GetRelatedMPANsResponse, error)
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

	technicalDetails, err := gw.client.GetTechnicalDetailsByMPAN(ctx, &ecoesv1.SearchByMPANRequest{
		Mpan: mpan,
	})
	if err != nil {
		code := status.Convert(err).Code()
		ecoesAPIResponses.WithLabelValues(code.String()).Inc()
		return nil, fmt.Errorf("failed to get technical details by mpan: %s, %w", mpan, err)
	}

	meters := []models.ElectricityMeter{}

	for _, elem := range technicalDetails.GetMeters() {
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

func (gw *EcoesGateway) GetRelatedMPAN(ctx context.Context, mpan string) (*models.ElectricityMeterRelatedMPAN, error) {

	relatedMPAN, err := gw.client.GetRelatedMPANs(ctx, &ecoesv1.SearchByMPANRequest{
		Mpan: mpan,
	})
	if err != nil {
		code := status.Convert(err).Code()
		ecoesAPIResponses.WithLabelValues(code.String()).Inc()
		return nil, fmt.Errorf("failed to get technical details by mpan: %s, %w", mpan, err)
	}

	relations := []models.MPANRelation{}

	for _, elem := range relatedMPAN.GetRelationships() {
		relations = append(relations, models.MPANRelation{
			Primary:   elem.PrimaryMpan.Mpan,
			Secondary: elem.SecondaryMpan.Mpan,
		})
	}

	return &models.ElectricityMeterRelatedMPAN{
		Relations: relations,
	}, nil
}
