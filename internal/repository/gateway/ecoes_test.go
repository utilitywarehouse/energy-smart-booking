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
	ecoesv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/ecoes/v1"
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
	mai := fakeMachineAuthInjector{}
	mai.ctx = ctx

	myGw := gateway.NewEcoesGateway(mEcoes)

	mEcoes.EXPECT().GetTechnicalDetailsByMPAN(ctx, &ecoesv1.SearchByMPANRequest{
		Mpan: "mpan-1",
	}).Return(&ecoesv1.TechnicalDetailsResponse{
		StandardSettlementConfiguration: "ssc-1",
		ProfileClass:                    platform.ProfileClass_PROFILE_CLASS_01,
		Meters: []*ecoesv1.MeterDetails{
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
		},
	}, nil)

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

	if !cmp.Equal(expected, actual, cmpopts.IgnoreUnexported()) {
		t.Fatalf("expected: %+v, actual: %+v", expected, actual)
	}
}

func Test_GetRelatedMPANs(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	mEcoes := mock_gateways.NewMockEcoesClient(ctrl)
	mai := fakeMachineAuthInjector{}
	mai.ctx = ctx

	myGw := gateway.NewEcoesGateway(mEcoes)

	mEcoes.EXPECT().GetRelatedMPANs(ctx, &ecoesv1.SearchByMPANRequest{
		Mpan: "mpan-1",
	}).Return(&ecoesv1.GetRelatedMPANsResponse{
		Relationships: []*ecoesv1.GetRelatedMPANsResponse_Relationship{
			{
				PrimaryMpan: &ecoesv1.RelatedMpan{
					Mpan: "mpan-1",
				},
				SecondaryMpan: &ecoesv1.RelatedMpan{
					Mpan: "mpan-2",
				},
			},
		},
	}, nil)

	actual := &models.ElectricityMeterRelatedMPAN{
		Relations: []models.MPANRelation{
			{
				Primary:   "mpan-1",
				Secondary: "mpan-2",
			},
		},
	}

	expected, err := myGw.GetRelatedMPAN(ctx, "mpan-1")
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, actual, cmpopts.IgnoreUnexported()) {
		t.Fatalf("expected: %+v, actual: %+v", expected, actual)
	}
}
