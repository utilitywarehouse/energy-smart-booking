//go:generate mockgen -source=customer.go -destination ./mocks/customer_mocks.go

package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	addressv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/address/v1"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain"
	mocks "github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain/mocks"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"google.golang.org/genproto/googleapis/type/date"
)

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()
	d, err := time.ParseInLocation(time.DateOnly, value, time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func Test_GetCustomerContactDetails(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	accGw := mocks.NewMockAccountGateway(ctrl)
	eliGw := mocks.NewMockEligibilityGateway(ctrl)
	occSt := mocks.NewMockOccupancyStore(ctrl)
	siteSt := mocks.NewMockSiteStore(ctrl)
	bookingSt := mocks.NewMockBookingStore(ctrl)

	myDomain := domain.NewCustomerDomain(accGw, eliGw, occSt, siteSt, bookingSt)

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
	bookingSt := mocks.NewMockBookingStore(ctrl)

	myDomain := domain.NewCustomerDomain(accGw, eliGw, occSt, siteSt, bookingSt)

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
				oSt.EXPECT().GetLiveOccupanciesByAccountID(ctx, "account-id-1").Return(
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
						SubBuildingNameNumber:   "sub-building-name-number",
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
			description: "account does not have occupancies",
			input: inputParams{
				accountID: "account-id-1",
			},
			setup: func(ctx context.Context, aGw *mocks.MockAccountGateway, eGw *mocks.MockEligibilityGateway, oSt *mocks.MockOccupancyStore, sSt *mocks.MockSiteStore) {
				oSt.EXPECT().GetLiveOccupanciesByAccountID(ctx, "account-id-1").Return([]models.Occupancy{}, nil)
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
				oSt.EXPECT().GetLiveOccupanciesByAccountID(ctx, "account-id-1").Return([]models.Occupancy{
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

func Test_GetCustomerBookings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	siteSt := mocks.NewMockSiteStore(ctrl)
	bookingSt := mocks.NewMockBookingStore(ctrl)

	myDomain := domain.NewCustomerDomain(
		mocks.NewMockAccountGateway(ctrl),
		mocks.NewMockEligibilityGateway(ctrl),
		mocks.NewMockOccupancyStore(ctrl),
		siteSt,
		bookingSt)

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
						BookingID: "booking-id-1",
						AccountID: "account-id-1",
						Status:    bookingv1.BookingStatus_BOOKING_STATUS_SCHEDULED,
						SiteID:    "site-id-1",
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
				sSt.EXPECT().GetSiteBySiteID(ctx, "site-id-1").Return(&models.Site{
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
						BookingID: "booking-id-1",
						AccountID: "account-id-1",
						Status:    bookingv1.BookingStatus_BOOKING_STATUS_COMPLETED,
						SiteID:    "site-id-1",
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
						BookingID: "booking-id-2",
						AccountID: "account-id-1",
						Status:    bookingv1.BookingStatus_BOOKING_STATUS_SCHEDULED,
						SiteID:    "site-id-2",
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
				sSt.EXPECT().GetSiteBySiteID(ctx, "site-id-1").Return(&models.Site{
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
				sSt.EXPECT().GetSiteBySiteID(ctx, "site-id-2").Return(&models.Site{
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
