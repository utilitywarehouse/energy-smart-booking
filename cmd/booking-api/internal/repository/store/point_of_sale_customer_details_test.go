package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store/serialisers"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"google.golang.org/protobuf/testing/protocmp"
)

// This test case tests for both upsert(only insert) and get actions.
func Test_PointOfSaleCustomerDetails_Insert_Get(t *testing.T) {
	ctx := context.Background()

	testContainer, err := setupTestContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	dsn, err := postgres.GetTestContainerDSN(testContainer)
	if err != nil {
		t.Fatal(err)
	}

	db, err := store.Setup(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	pointOfSaleCustomerDetailsStore := store.NewPointOfSaleCustomerDetails(db, serialisers.PointOfSaleCustomerDetails{})

	type inputParams struct {
		accountNumber string
		details       models.PointOfSaleCustomerDetails
	}

	type outputParams struct {
		detail *models.PointOfSaleCustomerDetails
		err    error
	}

	type testSetup struct {
		description string
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should upsert a booking created event and retrieve it",
			input: inputParams{
				accountNumber: "1",
				details: models.PointOfSaleCustomerDetails{
					AccountNumber: "1",
					Details: models.AccountDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Doe",
						Email:     "jdoe@example.com",
						Mobile:    "555-1001",
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
					Meterpoints: []models.Meterpoint{
						{
							MPXN:       "mpxn-1",
							TariffType: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
						},
						{
							MPXN:       "mpxn-2",
							TariffType: bookingv1.TariffType_TARIFF_TYPE_PREPAYMENT,
						},
					},
				},
			},
			output: outputParams{
				err: nil,
				detail: &models.PointOfSaleCustomerDetails{
					AccountNumber: "1",
					Details: models.AccountDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Doe",
						Email:     "jdoe@example.com",
						Mobile:    "555-1001",
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
					Meterpoints: []models.Meterpoint{
						{
							MPXN:       "mpxn-1",
							TariffType: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
						},
						{
							MPXN:       "mpxn-2",
							TariffType: bookingv1.TariffType_TARIFF_TYPE_PREPAYMENT,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			err := pointOfSaleCustomerDetailsStore.Upsert(ctx, tc.input.accountNumber, tc.input.details)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}

			actual, err := pointOfSaleCustomerDetailsStore.GetByAccountNumber(ctx, "1")

			if diff := cmp.Diff(tc.output.detail, actual, cmpopts.IgnoreUnexported(), protocmp.Transform(), cmpopts.EquateApproxTime(time.Second)); diff != "" {
				t.Fatal(diff)
			}

			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}
		})

	}
}

func Test_PointOfSaleCustomerDetails_Upsert_Get(t *testing.T) {
	ctx := context.Background()

	testContainer, err := setupTestContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	dsn, err := postgres.GetTestContainerDSN(testContainer)
	if err != nil {
		t.Fatal(err)
	}

	db, err := store.Setup(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	pointOfSaleCustomerDetailsStore := store.NewPointOfSaleCustomerDetails(db, serialisers.PointOfSaleCustomerDetails{})

	type inputParams struct {
		accountNumber string
		detail        models.PointOfSaleCustomerDetails
		detail2       models.PointOfSaleCustomerDetails
	}

	type outputParams struct {
		detail *models.PointOfSaleCustomerDetails
		err    error
	}

	type testSetup struct {
		description string
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should upsert a booking created event for booking-id-1 and get the upserted value",
			input: inputParams{
				accountNumber: "1",
				detail: models.PointOfSaleCustomerDetails{
					AccountNumber: "1",
					Details: models.AccountDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Doe",
						Email:     "jdoe@example.com",
						Mobile:    "555-1001",
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
					Meterpoints: []models.Meterpoint{
						{
							MPXN:       "mpxn-1",
							TariffType: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
						},
						{
							MPXN:       "mpxn-2",
							TariffType: bookingv1.TariffType_TARIFF_TYPE_PREPAYMENT,
						},
					},
				},
				detail2: models.PointOfSaleCustomerDetails{
					AccountNumber: "1",
					Details: models.AccountDetails{
						Title:     "Mr",
						FirstName: "Jane",
						LastName:  "Dough",
						Email:     "jane_dough@example.com",
						Mobile:    "555-1002",
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
					Meterpoints: []models.Meterpoint{
						{
							MPXN:       "mpxn-2-updated",
							TariffType: bookingv1.TariffType_TARIFF_TYPE_PREPAYMENT,
						},
						{
							MPXN:       "mpxn-2-updated",
							TariffType: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
						},
					},
				},
			},
			output: outputParams{
				err: nil,
				detail: &models.PointOfSaleCustomerDetails{
					AccountNumber: "1",
					Details: models.AccountDetails{
						Title:     "Mr",
						FirstName: "Jane",
						LastName:  "Dough",
						Email:     "jane_dough@example.com",
						Mobile:    "555-1002",
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
					Meterpoints: []models.Meterpoint{
						{
							MPXN:       "mpxn-2-updated",
							TariffType: bookingv1.TariffType_TARIFF_TYPE_PREPAYMENT,
						},
						{
							MPXN:       "mpxn-2-updated",
							TariffType: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			err := pointOfSaleCustomerDetailsStore.Upsert(ctx, tc.input.accountNumber, tc.input.detail)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}

			err = pointOfSaleCustomerDetailsStore.Upsert(ctx, tc.input.accountNumber, tc.input.detail2)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}

			actual, err := pointOfSaleCustomerDetailsStore.GetByAccountNumber(ctx, "1")

			if diff := cmp.Diff(tc.output.detail, actual, cmpopts.IgnoreUnexported(), protocmp.Transform(), cmpopts.EquateApproxTime(time.Second)); diff != "" {
				t.Fatal(diff)
			}

			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}
		})

	}
}

func Test_PointOfSaleCustomerDetails_Insert_Delete_Get(t *testing.T) {
	ctx := context.Background()

	testContainer, err := setupTestContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	dsn, err := postgres.GetTestContainerDSN(testContainer)
	if err != nil {
		t.Fatal(err)
	}

	db, err := store.Setup(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	pointOfSaleCustomerDetailsStore := store.NewPointOfSaleCustomerDetails(db, serialisers.PointOfSaleCustomerDetails{})

	type inputParams struct {
		accountNumber string
		detail        models.PointOfSaleCustomerDetails
	}

	type outputParams struct {
		err            error
		partialBooking *models.PointOfSaleCustomerDetails
	}

	type testSetup struct {
		description string
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should insert customer details, mark it as deleted and fail to retrieve it, because it was deleted",
			input: inputParams{
				accountNumber: "1",
				detail: models.PointOfSaleCustomerDetails{
					AccountNumber: "1",
					Details: models.AccountDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Doe",
						Email:     "jdoe@example.com",
						Mobile:    "555-1001",
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
					Meterpoints: []models.Meterpoint{
						{
							MPXN:       "mpxn-1",
							TariffType: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
						},
						{
							MPXN:       "mpxn-2",
							TariffType: bookingv1.TariffType_TARIFF_TYPE_PREPAYMENT,
						},
					},
				},
			},
			output: outputParams{
				err:            store.ErrPointOfSaleCustomerDetailsNotFound,
				partialBooking: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			err := pointOfSaleCustomerDetailsStore.Upsert(ctx, tc.input.accountNumber, tc.input.detail)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}

			err = pointOfSaleCustomerDetailsStore.Delete(ctx, tc.input.accountNumber)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}

			actual, err := pointOfSaleCustomerDetailsStore.GetByAccountNumber(ctx, tc.input.accountNumber)
			if err != nil {
				if !errors.Is(err, tc.output.err) {
					t.Fatalf("should not have errored, %s", err)
				}
			}

			if diff := cmp.Diff(tc.output.partialBooking, actual, cmpopts.IgnoreUnexported(), protocmp.Transform(), cmpopts.EquateApproxTime(time.Second)); diff != "" {
				t.Fatal(diff)
			}

		})

	}
}
