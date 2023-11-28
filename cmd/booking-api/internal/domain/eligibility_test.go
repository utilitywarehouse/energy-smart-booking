//go:generate mockgen -source=domain.go -destination ./mocks/domain_mocks.go

package domain_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain"
	mocks "github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain/mocks"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

func Test_GetClickLink(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()
	pointOfSaleCustomerDetailsSt := mocks.NewMockPointOfSaleCustomerDetailsStore(ctrl)
	eligbilityGw := mocks.NewMockEligibilityGateway(ctrl)
	clickGw := mocks.NewMockClickGateway(ctrl)

	myDomain := domain.NewBookingDomain(nil, nil, nil, nil, nil, nil, pointOfSaleCustomerDetailsSt, eligbilityGw, clickGw, false)

	type inputParams struct {
		accountNumber string
		details       models.PointOfSaleCustomerDetails
	}

	type outputParams struct {
		result domain.GetClickLinkResult
		err    error
	}

	type testSetup struct {
		description string
		setup       func(context.Context, *mocks.MockPointOfSaleCustomerDetailsStore, *mocks.MockEligibilityGateway, *mocks.MockClickGateway)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should process the eligiblity of a meterpoint and store the customer details",
			input: inputParams{
				accountNumber: "1",
				details: models.PointOfSaleCustomerDetails{
					AccountNumber: "1",
					Details: models.AccountDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Doe",
						Email:     "jdoe@example.com",
						Mobile:    "555-100",
					},
					Address: models.AccountAddress{
						UPRN: "u",
						PAF: models.PAF{
							BuildingName:            "bn",
							BuildingNumber:          "bn1",
							Department:              "dp",
							DependentLocality:       "dl",
							DependentThoroughfare:   "dtg",
							DoubleDependentLocality: "ddl",
							Organisation:            "o",
							PostTown:                "pt",
							Postcode:                "E2 1Z",
							SubBuilding:             "sb",
							Thoroughfare:            "tf",
						},
					},
					ElecOrderSupplies: models.OrderSupply{
						MPXN:       "2199996734008",
						TariffType: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
					},
					GasOrderSupplies: models.OrderSupply{
						MPXN:       "2724968810",
						TariffType: bookingv1.TariffType_TARIFF_TYPE_PREPAYMENT,
					},
				},
			},
			setup: func(ctx context.Context, p *mocks.MockPointOfSaleCustomerDetailsStore, e *mocks.MockEligibilityGateway, c *mocks.MockClickGateway) {
				e.EXPECT().GetMeterpointEligibility(ctx, "2199996734008", "2724968810", "E2 1Z").Return(true, nil)

				p.EXPECT().Upsert(ctx, "1", models.PointOfSaleCustomerDetails{
					AccountNumber: "1",
					Details: models.AccountDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Doe",
						Email:     "jdoe@example.com",
						Mobile:    "555-100",
					},
					Address: models.AccountAddress{
						UPRN: "u",
						PAF: models.PAF{
							BuildingName:            "bn",
							BuildingNumber:          "bn1",
							Department:              "dp",
							DependentLocality:       "dl",
							DependentThoroughfare:   "dtg",
							DoubleDependentLocality: "ddl",
							Organisation:            "o",
							PostTown:                "pt",
							Postcode:                "E2 1Z",
							SubBuilding:             "sb",
							Thoroughfare:            "tf",
						},
					},
					ElecOrderSupplies: models.OrderSupply{
						MPXN:       "2199996734008",
						TariffType: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
					},
					GasOrderSupplies: models.OrderSupply{
						MPXN:       "2724968810",
						TariffType: bookingv1.TariffType_TARIFF_TYPE_PREPAYMENT,
					},
				}).Return(nil)

				c.EXPECT().GenerateAuthenticated(ctx, "1", map[string]string{
					"account_number": "1",
					"journey_type":   "point_of_sale"}).Return("best_link_ever", nil)
			},
			output: outputParams{
				result: domain.GetClickLinkResult{
					Eligible: true,
					Link:     "best_link_ever",
				},
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, pointOfSaleCustomerDetailsSt, eligbilityGw, clickGw)

			expected, err := myDomain.GetClickLink(ctx, domain.GetClickLinkParams{
				AccountNumber: tc.input.accountNumber,
				Details:       tc.input.details,
			})

			if diff := cmp.Diff(err, tc.output.err, cmpopts.EquateErrors()); diff != "" {
				t.Fatal(err)
			}

			if diff := cmp.Diff(expected, tc.output.result); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
