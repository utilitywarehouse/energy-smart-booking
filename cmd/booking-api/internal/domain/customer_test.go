//go:generate mockgen -source=customer.go -destination ./mocks/customer_mocks.go

package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain"
	mocks "github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain/mocks"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

func Test_GetCustomerContactDetails(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	accGw := mocks.NewMockAccountGateway(ctrl)
	eliGw := mocks.NewMockEligibilityGateway(ctrl)
	occSt := mocks.NewMockOccupancyStore(ctrl)
	siteSt := mocks.NewMockSiteStore(ctrl)

	myDomain := domain.NewCustomerDomain(accGw, eliGw, occSt, siteSt)

	type inputParams struct {
		accountID string
	}

	type outputParams struct {
		account models.Account
		err     error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, aGw *mocks.MockAccountGateway, eGw *mocks.MockEligibilityGateway, oSt *mocks.MockOccupancyStore, sSt *mocks.MockSiteStore)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the account details by account id",
			input: inputParams{
				accountID: "account-id-1",
			},
			setup: func(ctx context.Context, aGw *mocks.MockAccountGateway, eGw *mocks.MockEligibilityGateway, oSt *mocks.MockOccupancyStore, sSt *mocks.MockSiteStore) {
				aGw.EXPECT().GetAccountByAccountID(ctx, "account-id-1").Return(models.Account{
					AccountID: "account-id-1",
					Details: models.AccountDetails{
						Title:     "Mr.",
						FirstName: "John",
						LastName:  "Doe",
						Email:     "johndoe@example.com",
						Mobile:    "555-0000",
					},
				}, nil)
			},
			output: outputParams{
				account: models.Account{
					AccountID: "account-id-1",
					Details: models.AccountDetails{
						Title:     "Mr.",
						FirstName: "John",
						LastName:  "Doe",
						Email:     "johndoe@example.com",
						Mobile:    "555-0000",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, accGw, eliGw, occSt, siteSt)

			expected, err := myDomain.GetCustomerContactDetails(ctx, tc.input.accountID)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(expected, tc.output.account); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_GetAccountAddressByAccountID(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	accGw := mocks.NewMockAccountGateway(ctrl)
	eliGw := mocks.NewMockEligibilityGateway(ctrl)
	occSt := mocks.NewMockOccupancyStore(ctrl)
	siteSt := mocks.NewMockSiteStore(ctrl)

	myDomain := domain.NewCustomerDomain(accGw, eliGw, occSt, siteSt)

	type inputParams struct {
		accountID string
	}

	type outputParams struct {
		address models.AccountAddress
		err     error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, aGw *mocks.MockAccountGateway, eGw *mocks.MockEligibilityGateway, oSt *mocks.MockOccupancyStore, sSt *mocks.MockSiteStore)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the account addres based on the account id and the first eligible occupancy",
			input: inputParams{
				accountID: "account-id-1",
			},
			setup: func(ctx context.Context, aGw *mocks.MockAccountGateway, eGw *mocks.MockEligibilityGateway, oSt *mocks.MockOccupancyStore, sSt *mocks.MockSiteStore) {
				oSt.EXPECT().GetOccupanciesByAccountID(ctx, "account-id-1").Return(
					[]models.Occupancy{
						{
							OccupancyID: "occupancy-id-1",
							SiteID:      "site-id-1",
							AccountID:   "account-id-1",
							CreatedAt:   time.Time{},
						},
						{
							OccupancyID: "occupancy-id-2",
							SiteID:      "site-id-2",
							AccountID:   "account-id-1",
							CreatedAt:   time.Time{},
						},
						{
							OccupancyID: "occupancy-id-3",
							SiteID:      "site-id-3",
							AccountID:   "account-id-1",
							CreatedAt:   time.Time{},
						},
					}, nil)

				eGw.EXPECT().GetEligibility(ctx, "account-id-1", "occupancy-id-1").Return(false, nil)
				eGw.EXPECT().GetEligibility(ctx, "account-id-1", "occupancy-id-2").Return(true, nil)

				sSt.EXPECT().GetSiteBySiteID(ctx, "site-id-2").Return(
					&models.Site{
						SiteID:                  "site-id-1",
						Postcode:                "post-code",
						UPRN:                    "uprn",
						BuildingNameNumber:      "building-name-number",
						DependentThoroughfare:   "dependent-thoroughfare",
						ThoroughFare:            "thoroughfare",
						DoubleDependentLocality: "ddl-1",
						DependentLocality:       "dl-1",
						Locality:                "locality",
						County:                  "county",
						Town:                    "town",
						Department:              "department",
						Organisation:            "organisation",
						PoBox:                   "pobox",
						DeliveryPointSuffix:     "delivery-point-suffix",
					}, nil)
			},
			output: outputParams{
				address: models.AccountAddress{
					UPRN: "uprn",
					PAF: models.PAF{
						BuildingName:            "",
						BuildingNumber:          "building-name-number",
						Department:              "department",
						DependentLocality:       "dl-1",
						DependentThoroughfare:   "dependent-thoroughfare",
						DoubleDependentLocality: "ddl-1",
						Organisation:            "organisation",
						PostTown:                "town",
						Postcode:                "post-code",
						SubBuilding:             "",
						Thoroughfare:            "thoroughfare",
					},
				},
			},
		},
		{
			description: "account does not have occupancies",
			input: inputParams{
				accountID: "account-id-1",
			},
			setup: func(ctx context.Context, aGw *mocks.MockAccountGateway, eGw *mocks.MockEligibilityGateway, oSt *mocks.MockOccupancyStore, sSt *mocks.MockSiteStore) {
				oSt.EXPECT().GetOccupanciesByAccountID(ctx, "account-id-1").Return([]models.Occupancy{}, nil)
			},
			output: outputParams{
				address: models.AccountAddress{},
				err:     domain.ErrNoOccupanciesFound,
			},
		},
		{
			description: "account does not have any eligible occupancies",
			input: inputParams{
				accountID: "account-id-1",
			},
			setup: func(ctx context.Context, aGw *mocks.MockAccountGateway, eGw *mocks.MockEligibilityGateway, oSt *mocks.MockOccupancyStore, sSt *mocks.MockSiteStore) {
				oSt.EXPECT().GetOccupanciesByAccountID(ctx, "account-id-1").Return([]models.Occupancy{
					{
						OccupancyID: "occupancy-id-1",
						SiteID:      "site-id-1",
						AccountID:   "account-id-1",
						CreatedAt:   time.Time{},
					},
					{
						OccupancyID: "occupancy-id-2",
						SiteID:      "site-id-2",
						AccountID:   "account-id-1",
						CreatedAt:   time.Time{},
					},
				}, nil)

				eGw.EXPECT().GetEligibility(ctx, "account-id-1", "occupancy-id-1").Return(false, nil)
				eGw.EXPECT().GetEligibility(ctx, "account-id-1", "occupancy-id-2").Return(false, nil)
			},
			output: outputParams{
				address: models.AccountAddress{},
				err:     domain.ErrNoEligibleOccupanciesFound,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, accGw, eliGw, occSt, siteSt)

			expected, err := myDomain.GetAccountAddressByAccountID(ctx, tc.input.accountID)

			if diff := cmp.Diff(err, tc.output.err, cmpopts.EquateErrors()); diff != "" {
				t.Fatal(err)
			}

			if diff := cmp.Diff(expected, tc.output.address); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
