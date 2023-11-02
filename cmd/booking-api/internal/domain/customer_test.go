//go:generate mockgen -source=domain.go -destination ./mocks/domain_mocks.go

package domain_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	addressv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/address/v1"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain"
	mocks "github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain/mocks"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"google.golang.org/genproto/googleapis/type/date"
)

func Test_GetCustomerContactDetails(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	accGw := mocks.NewMockAccountGateway(ctrl)

	myDomain := domain.NewBookingDomain(accGw, nil, nil, nil, nil, nil, nil, nil, nil, false)

	type inputParams struct {
		accountID string
	}

	type outputParams struct {
		account models.Account
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, aGw *mocks.MockAccountGateway)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the account details by account id",
			input: inputParams{
				accountID: "account-id-1",
			},
			setup: func(ctx context.Context, aGw *mocks.MockAccountGateway) {
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

			tc.setup(ctx, accGw)

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

	occSt := mocks.NewMockOccupancyStore(ctrl)

	myDomain := domain.NewBookingDomain(nil, nil, occSt, nil, nil, nil, nil, nil, nil, false)

	type inputParams struct {
		accountID string
	}

	type outputParams struct {
		address models.AccountAddress
		err     error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, oSt *mocks.MockOccupancyStore)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the account address based on the account id and the first eligible occupancy",
			input: inputParams{
				accountID: "account-id-1",
			},
			setup: func(ctx context.Context, oSt *mocks.MockOccupancyStore) {
				oSt.EXPECT().GetSiteExternalReferenceByAccountID(ctx, "account-id-1").Return(&models.Site{
					SiteID:                  "site-id-1",
					Postcode:                "post-code",
					UPRN:                    "uprn",
					BuildingNameNumber:      "building-name-number",
					SubBuildingNameNumber:   "sub-building-name-number",
					DependentThoroughfare:   "dependent-thoroughfare",
					Thoroughfare:            "thoroughfare",
					DoubleDependentLocality: "ddl-1",
					DependentLocality:       "dl-1",
					Locality:                "locality",
					County:                  "county",
					Town:                    "town",
					Department:              "department",
					Organisation:            "organisation",
					PoBox:                   "pobox",
					DeliveryPointSuffix:     "delivery-point-suffix",
				}, &models.OccupancyEligibility{
					OccupancyID: "occupancy-id-1",
					Reference:   "reference-1",
				}, nil)
			},
			output: outputParams{
				address: models.AccountAddress{
					UPRN: "uprn",
					PAF: models.PAF{
						BuildingName:            "building-name-number",
						BuildingNumber:          "building-name-number",
						Department:              "department",
						DependentLocality:       "dl-1",
						DependentThoroughfare:   "dependent-thoroughfare",
						DoubleDependentLocality: "ddl-1",
						Organisation:            "organisation",
						PostTown:                "town",
						Postcode:                "post-code",
						SubBuilding:             "sub-building-name-number",
						Thoroughfare:            "thoroughfare",
					},
				},
			},
		},
		{
			description: "account does not have eligible occupancies",
			input: inputParams{
				accountID: "account-id-1",
			},
			setup: func(ctx context.Context, oSt *mocks.MockOccupancyStore) {
				oSt.EXPECT().GetSiteExternalReferenceByAccountID(ctx, "account-id-1").Return(&models.Site{}, &models.OccupancyEligibility{}, store.ErrNoEligibleOccupancyFound)
			},
			output: outputParams{
				address: models.AccountAddress{},
				err:     domain.ErrNoEligibleOccupanciesFound,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, occSt)

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

func Test_GetCustomerBookings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	siteSt := mocks.NewMockSiteStore(ctrl)
	bookingSt := mocks.NewMockBookingStore(ctrl)

	myDomain := domain.NewBookingDomain(nil, nil, nil, siteSt, bookingSt, nil, nil, nil, nil, false)

	type inputParams struct {
		accountID string
	}

	type outputParams struct {
		bookings []*bookingv1.Booking
		err      error
	}

	type testSetup struct {
		description string
		setup       func(context.Context, *mocks.MockSiteStore, *mocks.MockBookingStore)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "basic bookings retrieval",
			setup: func(ctx context.Context, sSt *mocks.MockSiteStore, bSt *mocks.MockBookingStore) {
				bSt.EXPECT().GetBookingsByAccountID(ctx, "account-id-1").Return([]models.Booking{
					{
						BookingID:   "booking-id-1",
						AccountID:   "account-id-1",
						Status:      bookingv1.BookingStatus_BOOKING_STATUS_SCHEDULED,
						OccupancyID: "occupancy-id-1",
						Contact: models.AccountDetails{
							Title:     "Mr.",
							FirstName: "Foo",
							LastName:  "Bar",
							Email:     "foobar@example.com",
							Mobile:    "555-0000",
						},
						Slot: models.BookingSlot{
							Date:      mustDate(t, "2023-09-16"),
							StartTime: 13,
							EndTime:   16,
						},
						VulnerabilityDetails: models.VulnerabilityDetails{
							Vulnerabilities: []bookingv1.Vulnerability{},
							Other:           "",
						},
					},
				}, nil)
				sSt.EXPECT().GetSiteByOccupancyID(ctx, "occupancy-id-1").Return(&models.Site{
					SiteID:                  "site-id-1",
					Postcode:                "postcode",
					UPRN:                    "uprn",
					BuildingNameNumber:      "building-name-number",
					DependentThoroughfare:   "dt-1",
					Thoroughfare:            "thru-1",
					DoubleDependentLocality: "ddl-1",
					DependentLocality:       "dl-1",
					Locality:                "locality",
					County:                  "county",
					Town:                    "town",
					Department:              "department",
					Organisation:            "organisation",
					PoBox:                   "pobox",
					DeliveryPointSuffix:     "delivery-point-suffix",
					SubBuildingNameNumber:   "sub-building-name-number",
				}, nil)
			},
			input: inputParams{
				accountID: "account-id-1",
			},
			output: outputParams{
				bookings: []*bookingv1.Booking{
					{
						Id:        "booking-id-1",
						AccountId: "account-id-1",
						SiteAddress: &addressv1.Address{
							Uprn: "uprn",
							Paf: &addressv1.Address_PAF{
								Organisation:            "organisation",
								Department:              "department",
								SubBuilding:             "sub-building-name-number",
								BuildingName:            "building-name-number",
								BuildingNumber:          "building-name-number",
								DependentThoroughfare:   "dt-1",
								Thoroughfare:            "thru-1",
								DoubleDependentLocality: "ddl-1",
								DependentLocality:       "dl-1",
								PostTown:                "town",
								Postcode:                "postcode",
							},
						},
						ContactDetails: &bookingv1.ContactDetails{
							Title:     "Mr.",
							FirstName: "Foo",
							LastName:  "Bar",
							Phone:     "555-0000",
							Email:     "foobar@example.com",
						},
						Slot: &bookingv1.BookingSlot{
							Date: &date.Date{
								Year:  2023,
								Month: 9,
								Day:   16,
							},
							StartTime: 13,
							EndTime:   16,
						},
						VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
							Vulnerabilities: []bookingv1.Vulnerability{},
							Other:           "",
						},
						Status: bookingv1.BookingStatus_BOOKING_STATUS_SCHEDULED,
					},
				},
				err: nil,
			},
		},
		{
			description: "multiple bookings retrieval",
			setup: func(ctx context.Context, sSt *mocks.MockSiteStore, bSt *mocks.MockBookingStore) {
				bSt.EXPECT().GetBookingsByAccountID(ctx, "account-id-1").Return([]models.Booking{
					{
						BookingID:   "booking-id-1",
						AccountID:   "account-id-1",
						Status:      bookingv1.BookingStatus_BOOKING_STATUS_COMPLETED,
						OccupancyID: "occupancy-id-1",
						Contact: models.AccountDetails{
							Title:     "Mr.",
							FirstName: "Foo",
							LastName:  "Bar",
							Email:     "foobar@example.com",
							Mobile:    "555-0000",
						},
						Slot: models.BookingSlot{
							Date:      mustDate(t, "2023-07-30"),
							StartTime: 13,
							EndTime:   16,
						},
						VulnerabilityDetails: models.VulnerabilityDetails{
							Vulnerabilities: []bookingv1.Vulnerability{},
							Other:           "",
						},
					},
					{
						BookingID:   "booking-id-2",
						AccountID:   "account-id-1",
						Status:      bookingv1.BookingStatus_BOOKING_STATUS_SCHEDULED,
						OccupancyID: "occupancy-id-2",
						Contact: models.AccountDetails{
							Title:     "Ms.",
							FirstName: "Laurie",
							LastName:  "Bach",
							Email:     "laurie@bach.com",
							Mobile:    "123-4567",
						},
						Slot: models.BookingSlot{
							Date:      mustDate(t, "2023-11-11"),
							StartTime: 11,
							EndTime:   13,
						},
						VulnerabilityDetails: models.VulnerabilityDetails{
							Vulnerabilities: []bookingv1.Vulnerability{bookingv1.Vulnerability_VULNERABILITY_ILLNESS},
							Other:           "",
						},
					},
				}, nil)
				sSt.EXPECT().GetSiteByOccupancyID(ctx, "occupancy-id-1").Return(&models.Site{
					SiteID:                  "site-id-1",
					Postcode:                "postcode-1",
					UPRN:                    "uprn-1",
					BuildingNameNumber:      "building-name-number-1",
					DependentThoroughfare:   "dt-1",
					Thoroughfare:            "thru-1",
					DoubleDependentLocality: "ddl-1",
					DependentLocality:       "dl-1",
					Locality:                "locality-1",
					County:                  "county-1",
					Town:                    "town-1",
					Department:              "department-1",
					Organisation:            "organisation-1",
					PoBox:                   "pobox-1",
					DeliveryPointSuffix:     "delivery-point-suffix-1",
					SubBuildingNameNumber:   "sub-building-name-number-1",
				}, nil)
				sSt.EXPECT().GetSiteByOccupancyID(ctx, "occupancy-id-2").Return(&models.Site{
					SiteID:                  "site-id-2",
					Postcode:                "postcode-2",
					UPRN:                    "uprn-2",
					BuildingNameNumber:      "building-name-number-2",
					DependentThoroughfare:   "dt-2",
					Thoroughfare:            "thru-2",
					DoubleDependentLocality: "ddl-2",
					DependentLocality:       "dl-2",
					Locality:                "locality-2",
					County:                  "county-2",
					Town:                    "town-2",
					Department:              "department-2",
					Organisation:            "organisation-2",
					PoBox:                   "pobox-2",
					DeliveryPointSuffix:     "delivery-point-suffix-2",
					SubBuildingNameNumber:   "sub-building-name-number-2",
				}, nil)
			},
			input: inputParams{
				accountID: "account-id-1",
			},
			output: outputParams{
				bookings: []*bookingv1.Booking{
					{
						Id:        "booking-id-1",
						AccountId: "account-id-1",
						SiteAddress: &addressv1.Address{
							Uprn: "uprn-1",
							Paf: &addressv1.Address_PAF{
								Organisation:            "organisation-1",
								Department:              "department-1",
								SubBuilding:             "sub-building-name-number-1",
								BuildingName:            "building-name-number-1",
								BuildingNumber:          "building-name-number-1",
								DependentThoroughfare:   "dt-1",
								Thoroughfare:            "thru-1",
								DoubleDependentLocality: "ddl-1",
								DependentLocality:       "dl-1",
								PostTown:                "town-1",
								Postcode:                "postcode-1",
							},
						},
						ContactDetails: &bookingv1.ContactDetails{
							Title:     "Mr.",
							FirstName: "Foo",
							LastName:  "Bar",
							Phone:     "555-0000",
							Email:     "foobar@example.com",
						},
						Slot: &bookingv1.BookingSlot{
							Date: &date.Date{
								Year:  2023,
								Month: 7,
								Day:   30,
							},
							StartTime: 13,
							EndTime:   16,
						},
						VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
							Vulnerabilities: []bookingv1.Vulnerability{},
							Other:           "",
						},
						Status: bookingv1.BookingStatus_BOOKING_STATUS_COMPLETED,
					},
					{
						Id:        "booking-id-2",
						AccountId: "account-id-1",
						SiteAddress: &addressv1.Address{
							Uprn: "uprn-2",
							Paf: &addressv1.Address_PAF{
								Organisation:            "organisation-2",
								Department:              "department-2",
								SubBuilding:             "sub-building-name-number-2",
								BuildingName:            "building-name-number-2",
								BuildingNumber:          "building-name-number-2",
								DependentThoroughfare:   "dt-2",
								Thoroughfare:            "thru-2",
								DoubleDependentLocality: "ddl-2",
								DependentLocality:       "dl-2",
								PostTown:                "town-2",
								Postcode:                "postcode-2",
							},
						},
						ContactDetails: &bookingv1.ContactDetails{
							Title:     "Ms.",
							FirstName: "Laurie",
							LastName:  "Bach",
							Phone:     "123-4567",
							Email:     "laurie@bach.com",
						},
						Slot: &bookingv1.BookingSlot{
							Date: &date.Date{
								Year:  2023,
								Month: 11,
								Day:   11,
							},
							StartTime: 11,
							EndTime:   13,
						},
						VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
							Vulnerabilities: []bookingv1.Vulnerability{bookingv1.Vulnerability_VULNERABILITY_ILLNESS},
							Other:           "",
						},
						Status: bookingv1.BookingStatus_BOOKING_STATUS_SCHEDULED,
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			tc.setup(ctx, siteSt, bookingSt)

			actual, err := myDomain.GetCustomerBookings(ctx, tc.input.accountID)

			if diff := cmp.Diff(err, tc.output.err, cmpopts.EquateErrors()); diff != "" {
				t.Fatal(err)
			}

			if diff := cmp.Diff(
				actual,
				tc.output.bookings,
				cmpopts.IgnoreUnexported(bookingv1.Booking{}, addressv1.Address{}, addressv1.Address_PAF{}, bookingv1.ContactDetails{}, bookingv1.BookingSlot{}, bookingv1.VulnerabilityDetails{}, date.Date{}),
			); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_GetCustomerDetailsPointOfSale(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	pointOfSaleCustomerDetailsSt := mocks.NewMockPointOfSaleCustomerDetailsStore(ctrl)

	myDomain := domain.NewBookingDomain(nil, nil, nil, nil, nil, nil, pointOfSaleCustomerDetailsSt, nil, nil, false)

	type inputParams struct {
		accountNumber string
	}

	type outputParams struct {
		details *models.PointOfSaleCustomerDetails
		err     error
	}

	type testSetup struct {
		description string
		setup       func(context.Context, *mocks.MockPointOfSaleCustomerDetailsStore)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the customer details for the provided account number",
			input: inputParams{
				accountNumber: "1",
			},
			setup: func(ctx context.Context, p *mocks.MockPointOfSaleCustomerDetailsStore) {
				p.EXPECT().GetByAccountNumber(ctx, "1").Return(
					&models.PointOfSaleCustomerDetails{
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
					}, nil)
			},
			output: outputParams{
				details: &models.PointOfSaleCustomerDetails{
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
				err: nil,
			},
		},
		{
			description: "should fail to get the customer details because they don't exist",
			input: inputParams{
				accountNumber: "1",
			},
			setup: func(ctx context.Context, p *mocks.MockPointOfSaleCustomerDetailsStore) {
				p.EXPECT().GetByAccountNumber(ctx, "1").Return(nil, store.ErrPointOfSaleCustomerDetailsNotFound)
			},
			output: outputParams{
				details: nil,
				err:     domain.ErrPointOfSaleCustomerDetailsNotFound,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, pointOfSaleCustomerDetailsSt)

			expected, err := myDomain.GetCustomerDetailsPointOfSale(ctx, tc.input.accountNumber)

			if diff := cmp.Diff(err, tc.output.err, cmpopts.EquateErrors()); diff != "" {
				t.Fatal(err)
			}

			if diff := cmp.Diff(expected, tc.output.details); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
