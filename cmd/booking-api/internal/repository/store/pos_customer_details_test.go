package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

func Test_AccountDetailsStore_CacheAccountDetails(t *testing.T) {
	ctx := context.Background()

	container, err := SetupRedisTestContainer(ctx)
	if err != nil {
		t.Fatalf("could not set up redis test container: %s", err.Error())
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	containerAddr, err := GetRedisTestContainerAddr(ctx, container)
	if err != nil {
		t.Fatalf("could not get redis test container address: %s", err.Error())
	}
	accountDetailsStore := store.NewAccountDetailsStore(redis.NewClient(&redis.Options{Addr: containerAddr}), 6*time.Hour)

	type inputParams struct {
		accountDetails models.PointOfSaleCustomerDetails
	}

	type testSetup struct {
		description string
		input       inputParams
		output      error
	}

	testCases := []testSetup{
		{
			description: "should cache a pos customer account details record",
			input: inputParams{
				accountDetails: models.PointOfSaleCustomerDetails{
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
							Postcode:                "pc",
							SubBuilding:             "sb",
							Thoroughfare:            "tf",
						},
					},
					ElecOrderSupplies: models.OrderSupply{
						MPXN:       "mpxn-1",
						TariffType: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
					},
					GasOrderSupplies: models.OrderSupply{
						MPXN:       "mpxn-2",
						TariffType: bookingv1.TariffType_TARIFF_TYPE_PREPAYMENT,
					},
				},
			},
			output: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err := accountDetailsStore.Upsert(ctx, "1", tc.input.accountDetails)
			if err != tc.output {
				t.Fatalf("error output does not match, expected: %s | actual: %s", tc.output, err)
			}
		})
	}
}
func Test_AccountDetailsStore_GetAccountDetails(t *testing.T) {
	ctx := context.Background()

	container, err := SetupRedisTestContainer(ctx)
	if err != nil {
		t.Fatalf("could not set up redis test container: %s", err.Error())
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	containerAddr, err := GetRedisTestContainerAddr(ctx, container)
	if err != nil {
		t.Fatalf("could not get redis test container address: %s", err.Error())
	}
	accountDetailsStore := store.NewAccountDetailsStore(redis.NewClient(&redis.Options{Addr: containerAddr}), 6*time.Hour)

	type inputParams struct {
		accountDetails models.PointOfSaleCustomerDetails
		skipInsertion  bool
	}

	type testOutput struct {
		res *models.PointOfSaleCustomerDetails
		err error
	}

	type testSetup struct {
		description string
		input       inputParams
		output      testOutput
	}

	testCases := []testSetup{
		{
			description: "should cache and retrieve an account details record",
			input: inputParams{
				accountDetails: models.PointOfSaleCustomerDetails{
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
							Postcode:                "pc",
							SubBuilding:             "sb",
							Thoroughfare:            "tf",
						},
					},
					ElecOrderSupplies: models.OrderSupply{
						MPXN:       "mpxn-1",
						TariffType: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
					},
					GasOrderSupplies: models.OrderSupply{
						MPXN:       "mpxn-2",
						TariffType: bookingv1.TariffType_TARIFF_TYPE_PREPAYMENT,
					},
				},
			},
			output: testOutput{
				res: &models.PointOfSaleCustomerDetails{
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
							Postcode:                "pc",
							SubBuilding:             "sb",
							Thoroughfare:            "tf",
						},
					},
					ElecOrderSupplies: models.OrderSupply{
						MPXN:       "mpxn-1",
						TariffType: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
					},
					GasOrderSupplies: models.OrderSupply{
						MPXN:       "mpxn-2",
						TariffType: bookingv1.TariffType_TARIFF_TYPE_PREPAYMENT,
					},
				},
				err: nil,
			},
		},
		{
			description: "should not cache and get a NotFound error when attempting retrieval",
			input: inputParams{
				accountDetails: models.PointOfSaleCustomerDetails{AccountNumber: "2"},
				skipInsertion:  true,
			},
			output: testOutput{
				res: nil,
				err: store.ErrPOSCustomerDetailsNotFound,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if !tc.input.skipInsertion {
				err := accountDetailsStore.Upsert(ctx, "1", tc.input.accountDetails)
				if err != nil {
					t.Fatal("error when caching eligibility")
				}
			}
			res, err := accountDetailsStore.GetByAccountNumber(ctx, tc.input.accountDetails.AccountNumber)
			if err != tc.output.err {
				t.Fatalf("error output does not match, expected: %s | actual: %s", tc.output.err, err)
			}
			if cmp.Diff(res, tc.output.res, cmpopts.IgnoreUnexported()) != "" {
				t.Fatalf("account details output does not match, expected: %v | actual: %v", *tc.output.res, *res)
			}
		})
	}
}
