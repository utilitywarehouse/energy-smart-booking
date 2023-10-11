package gateway

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	ecoesv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/ecoes/v1"
	"google.golang.org/grpc"
)

type EcoesClient interface {
	GetTechnicalDetailsByMPAN(context.Context, *ecoesv1.SearchByMPANRequest, ...grpc.CallOption) (*ecoesv1.TechnicalDetailsResponse, error)
	GetRelatedMPANs(context.Context, *ecoesv1.SearchByMPANRequest, ...grpc.CallOption) (*ecoesv1.GetRelatedMPANsResponse, error)
}

type EcoesGateway struct {
	client EcoesClient
}

func NewEcoesGateway(client EcoesClient) *EcoesGateway {
	return &EcoesGateway{client}
}

type ElectricityMeter struct {
	MeterType   platform.MeterTypeElec
	InstalledAt time.Time
}

type ElectricityMeterTechnicalDetails struct {
	ProfileClass                    platform.ProfileClass
	SettlementStandardConfiguration string
	Meters                          []ElectricityMeter
}

type MPANRelation struct {
	Primary   string
	Secondary string
}

type ElectricityMeterRelatedMPAN struct {
	Relations []MPANRelation
}

func (gw *EcoesGateway) GetMPANTechnicalDetails(ctx context.Context, mpan string) (*ElectricityMeterTechnicalDetails, error) {

	technicalDetails, err := gw.client.GetTechnicalDetailsByMPAN(ctx, &ecoesv1.SearchByMPANRequest{
		Mpan: mpan,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get technical details by mpan: %s, %w", mpan, err)
	}

	meters := []ElectricityMeter{}

	for _, elem := range technicalDetails.GetMeters() {
		meters = append(meters, ElectricityMeter{
			MeterType:   elem.MeterType,
			InstalledAt: time.Date(int(elem.GetMeterInstalledDate().GetYear()), time.Month(elem.GetMeterInstalledDate().Month), int(elem.GetMeterInstalledDate().GetDay()), 0, 0, 0, 0, time.UTC),
		})
	}

	sort.Slice(meters, func(i, j int) bool {
		return meters[i].InstalledAt.After(meters[j].InstalledAt)
	})

	return &ElectricityMeterTechnicalDetails{
		ProfileClass:                    technicalDetails.GetProfileClass(),
		SettlementStandardConfiguration: technicalDetails.GetStandardSettlementConfiguration(),
		Meters:                          meters,
	}, nil
}

func (gw *EcoesGateway) GetRelatedMPAN(ctx context.Context, mpan string) (*ElectricityMeterRelatedMPAN, error) {

	relatedMPAN, err := gw.client.GetRelatedMPANs(ctx, &ecoesv1.SearchByMPANRequest{
		Mpan: mpan,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get technical details by mpan: %s, %w", mpan, err)
	}

	relations := []MPANRelation{}

	for _, elem := range relatedMPAN.GetRelationships() {
		relations = append(relations, MPANRelation{
			Primary:   elem.PrimaryMpan.Mpan,
			Secondary: elem.SecondaryMpan.Mpan,
		})
	}

	return &ElectricityMeterRelatedMPAN{
		Relations: relations,
	}, nil
}
