//go:generate mockgen -source=ecoes.go -destination ./mocks/ecoes_mocks.go

package gateway_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	ecoesv2 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/ecoes/v2"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
	mock_gateways "github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway/mocks"
	"google.golang.org/genproto/googleapis/type/date"
)

func Test_GetMPANTechnicalDetails(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	mEcoes := mock_gateways.NewMockEcoesClient(ctrl)
	myGw := gateway.NewEcoesGateway(mEcoes)

	ssc := "ssc-1"
	profileClass := platform.ProfileClass_PROFILE_CLASS_01
	mEcoes.EXPECT().GetTechnicalDetailsByMPAN(ctx, &ecoesv2.GetTechnicalDetailsByMPANRequest{
		Mpan: "mpan-1",
	}).Return(&ecoesv2.GetTechnicalDetailsByMPANResponse{TechnicalDetails: &ecoesv2.TechnicalDetails{
		StandardSettlementConfiguration: &ssc,
		ProfileClass:                    &profileClass,
		Meters: []*ecoesv2.MeterDetails{
			{
				MeterType: platform.MeterTypeElec_METER_TYPE_ELEC_CHECK,
				MeterInstalledDate: &date.Date{
					Year:  2020,
					Month: 1,
					Day:   1,
				},
			},
			{
				MeterType: platform.MeterTypeElec_METER_TYPE_ELEC_CHECK,
				MeterInstalledDate: &date.Date{
					Year:  2020,
					Month: 1,
					Day:   2,
				},
			},
			{
				MeterType: platform.MeterTypeElec_METER_TYPE_ELEC_CHECK,
				MeterInstalledDate: &date.Date{
					Year:  2020,
					Month: 1,
					Day:   3,
				},
			},
			{
				MeterType:          platform.MeterTypeElec_METER_TYPE_ELEC_2ADEF,
				MeterInstalledDate: nil,
			},
		},
	}}, nil)

	actual := &models.ElectricityMeterTechnicalDetails{
		Meters: []models.ElectricityMeter{
			{
				MeterType:   platform.MeterTypeElec_METER_TYPE_ELEC_CHECK,
				InstalledAt: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
			},
			{
				MeterType:   platform.MeterTypeElec_METER_TYPE_ELEC_CHECK,
				InstalledAt: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			{
				MeterType:   platform.MeterTypeElec_METER_TYPE_ELEC_CHECK,
				InstalledAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		ProfileClass:                    platform.ProfileClass_PROFILE_CLASS_01,
		SettlementStandardConfiguration: "ssc-1",
	}

	expected, err := myGw.GetMPANTechnicalDetails(ctx, "mpan-1")
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expected, actual, cmpopts.IgnoreUnexported()); diff != "" {
		t.Fatal(diff)
	}
}

func Test_GetRelatedMPANs(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	mEcoes := mock_gateways.NewMockEcoesClient(ctrl)
	myGw := gateway.NewEcoesGateway(mEcoes)

	mEcoes.EXPECT().GetRelatedMPANs(ctx, &ecoesv2.GetRelatedMPANsRequest{
		Mpan: "mpan-1",
	}).Return(&ecoesv2.GetRelatedMPANsResponse{
		Relationships: []*ecoesv2.GetRelatedMPANsResponse_Relationship{
			{
				PrimaryMpan: &ecoesv2.RelatedMpan{
					Mpan: "mpan-1",
				},
				SecondaryMpan: &ecoesv2.RelatedMpan{
					Mpan: "mpan-2",
				},
			},
		},
	}, nil)

	expected := true
	actual, err := myGw.HasRelatedMPAN(ctx, "mpan-1")
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, actual, cmpopts.IgnoreUnexported()) {
		t.Fatalf("expected: %+v, actual: %+v", expected, actual)
	}
}
