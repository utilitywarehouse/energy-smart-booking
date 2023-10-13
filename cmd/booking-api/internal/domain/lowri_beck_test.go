//go:generate mockgen -source=domain.go -destination ./mocks/domain_mocks.go

package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	addressv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/address/v1"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain"
	mocks "github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain/mocks"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
	"google.golang.org/genproto/googleapis/type/date"
)

func Test_GetAvailableSlots(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	accGw := mocks.NewMockAccountGateway(ctrl)
	lbGw := mocks.NewMockLowriBeckGateway(ctrl)
	occSt := mocks.NewMockOccupancyStore(ctrl)
	siteSt := mocks.NewMockSiteStore(ctrl)
	bookingSt := mocks.NewMockBookingStore(ctrl)
	partialBookingSt := mocks.NewMockPartialBookingStore(ctrl)

	myDomain := domain.NewBookingDomain(accGw, lbGw, occSt, siteSt, bookingSt, partialBookingSt, false)

	type inputParams struct {
		params domain.GetAvailableSlotsParams
	}

	type outputParams struct {
		output domain.GetAvailableSlotsResponse
		err    error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, oSt *mocks.MockOccupancyStore, lbGw *mocks.MockLowriBeckGateway)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the available slots between From and To",
			input: inputParams{
				params: domain.GetAvailableSlotsParams{
					AccountID: "account-id-1",
					From: &date.Date{
						Year:  2023,
						Month: 12,
						Day:   1,
					},
					To: &date.Date{
						Year:  2023,
						Month: 12,
						Day:   30,
					},
				},
			},
			setup: func(ctx context.Context, oSt *mocks.MockOccupancyStore, lbGw *mocks.MockLowriBeckGateway) {

				oSt.EXPECT().GetSiteExternalReferenceByAccountID(ctx, "account-id-1").Return(
					&models.Site{
						Postcode: "E2 1ZZ",
					},
					&models.OccupancyEligibility{
						OccupancyID: "occupancy-id-1",
						Reference:   "booking-reference-1",
					}, nil)

				lbGw.EXPECT().GetAvailableSlots(ctx, "E2 1ZZ", "booking-reference-1").Return(gateway.AvailableSlotsResponse{
					BookingSlots: []models.BookingSlot{
						{
							Date:      mustDate(t, "2023-12-05"),
							StartTime: 9,
							EndTime:   12,
						},
						{
							Date:      mustDate(t, "2023-11-05"),
							StartTime: 17,
							EndTime:   19,
						},
						{
							Date:      mustDate(t, "2023-12-10"),
							StartTime: 11,
							EndTime:   15,
						},
					},
				}, nil)

			},
			output: outputParams{
				output: domain.GetAvailableSlotsResponse{
					Slots: []models.BookingSlot{
						{
							Date:      mustDate(t, "2023-12-05"),
							StartTime: 9,
							EndTime:   12,
						},
						{
							Date:      mustDate(t, "2023-12-10"),
							StartTime: 11,
							EndTime:   15,
						},
					},
				},
				err: nil,
			},
		},
		{
			description: "should return an error because no dates available for provided date",
			input: inputParams{
				params: domain.GetAvailableSlotsParams{
					AccountID: "account-id-1",
					From: &date.Date{
						Year:  2020,
						Month: 12,
						Day:   1,
					},
					To: &date.Date{
						Year:  2020,
						Month: 12,
						Day:   30,
					},
				},
			},
			setup: func(ctx context.Context, oSt *mocks.MockOccupancyStore, lbGw *mocks.MockLowriBeckGateway) {

				oSt.EXPECT().GetSiteExternalReferenceByAccountID(ctx, "account-id-1").Return(
					&models.Site{
						Postcode: "E2 1ZZ",
					},
					&models.OccupancyEligibility{
						OccupancyID: "occupancy-id-1",
						Reference:   "booking-reference-1",
					}, nil)

				lbGw.EXPECT().GetAvailableSlots(ctx, "E2 1ZZ", "booking-reference-1").Return(gateway.AvailableSlotsResponse{
					BookingSlots: []models.BookingSlot{
						{
							Date:      mustDate(t, "2023-12-05"),
							StartTime: 9,
							EndTime:   12,
						},
						{
							Date:      mustDate(t, "2023-11-05"),
							StartTime: 17,
							EndTime:   19,
						},
						{
							Date:      mustDate(t, "2023-12-10"),
							StartTime: 11,
							EndTime:   15,
						},
					},
				}, nil)

			},
			output: outputParams{
				output: domain.GetAvailableSlotsResponse{
					Slots: []models.BookingSlot{},
				},
				err: domain.ErrNoAvailableSlotsForProvidedDates,
			},
		},
		{
			description: "should return err ErrNoOccupanciesFound",
			input: inputParams{
				params: domain.GetAvailableSlotsParams{
					AccountID: "account-id-1",
					From: &date.Date{
						Year:  2023,
						Month: 12,
						Day:   1,
					},
					To: &date.Date{
						Year:  2023,
						Month: 12,
						Day:   30,
					},
				},
			},
			setup: func(ctx context.Context, oSt *mocks.MockOccupancyStore, lbGw *mocks.MockLowriBeckGateway) {

				oSt.EXPECT().GetSiteExternalReferenceByAccountID(ctx, "account-id-1").Return(
					&models.Site{
						Postcode: "E2 1ZZ",
					},
					&models.OccupancyEligibility{
						OccupancyID: "occupancy-id-1",
						Reference:   "booking-reference-1",
					}, store.ErrNoEligibleOccupancyFound)
			},
			output: outputParams{
				output: domain.GetAvailableSlotsResponse{
					Slots: nil,
				},
				err: domain.ErrNoEligibleOccupanciesFound,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, occSt, lbGw)

			actual, err := myDomain.GetAvailableSlots(ctx, tc.input.params)

			if tc.output.err != nil {
				if !errors.Is(err, tc.output.err) {
					t.Fatalf("expected: %s, actual: %s", err, tc.output.err)
				}
			}

			if diff := cmp.Diff(actual, tc.output.output, cmpopts.IgnoreUnexported(date.Date{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_CreateBooking(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	accGw := mocks.NewMockAccountGateway(ctrl)
	lbGw := mocks.NewMockLowriBeckGateway(ctrl)
	occSt := mocks.NewMockOccupancyStore(ctrl)
	siteSt := mocks.NewMockSiteStore(ctrl)
	bookingSt := mocks.NewMockBookingStore(ctrl)
	partialBookingSt := mocks.NewMockPartialBookingStore(ctrl)

	myDomain := domain.NewBookingDomain(accGw, lbGw, occSt, siteSt, bookingSt, partialBookingSt, false)

	var emptyMsg *bookingv1.BookingCreatedEvent

	type inputParams struct {
		params domain.CreateBookingParams
	}

	type outputParams struct {
		event domain.CreateBookingResponse
		err   error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, oSt *mocks.MockOccupancyStore, lbGw *mocks.MockLowriBeckGateway)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should create booking",
			input: inputParams{
				params: domain.CreateBookingParams{
					AccountID: "account-id-1",
					ContactDetails: models.AccountDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Dough",
						Email:     "jdough@example.com",
						Mobile:    "555-0145",
					},
					Slot: models.BookingSlot{
						Date:      mustDate(t, "2023-08-27"),
						StartTime: 9,
						EndTime:   15,
					},
					VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
						Vulnerabilities: []bookingv1.Vulnerability{
							bookingv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
						},
						Other: "",
					},
					Source: bookingv1.BookingSource_BOOKING_SOURCE_PLATFORM_APP,
				},
			},
			setup: func(ctx context.Context, oSt *mocks.MockOccupancyStore, lbGw *mocks.MockLowriBeckGateway) {

				oSt.EXPECT().GetSiteExternalReferenceByAccountID(ctx, "account-id-1").Return(
					&models.Site{
						SiteID:                  "site-id-1",
						Postcode:                "E2 1ZZ",
						UPRN:                    "u",
						BuildingNameNumber:      "bn",
						SubBuildingNameNumber:   "sb",
						DependentThoroughfare:   "dt",
						Thoroughfare:            "t",
						DoubleDependentLocality: "ddl",
						DependentLocality:       "dl",
						Locality:                "l",
						County:                  "c",
						Town:                    "pt",
						Department:              "d",
						Organisation:            "o",
						PoBox:                   "po",
						DeliveryPointSuffix:     "dps",
					},
					&models.OccupancyEligibility{
						OccupancyID: "occupancy-id-1",
						Reference:   "booking-reference-1",
					}, nil)

				lbGw.EXPECT().CreateBooking(ctx, "E2 1ZZ", "booking-reference-1", models.BookingSlot{
					Date:      mustDate(t, "2023-08-27"),
					StartTime: 9,
					EndTime:   15,
				}, models.AccountDetails{
					Title:     "Mr",
					FirstName: "John",
					LastName:  "Dough",
					Email:     "jdough@example.com",
					Mobile:    "555-0145",
				}, []lowribeckv1.Vulnerability{
					lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
				}, "").Return(gateway.CreateBookingResponse{
					Success: true,
				}, nil)

			},
			output: outputParams{
				event: domain.CreateBookingResponse{
					Event: &bookingv1.BookingCreatedEvent{
						BookingId:     "my-uuid",
						OccupancyId:   "occupancy-id-1",
						BookingSource: bookingv1.BookingSource_BOOKING_SOURCE_PLATFORM_APP,
						Details: &bookingv1.Booking{
							Id:        "my-uuid",
							AccountId: "account-id-1",
							SiteAddress: &addressv1.Address{
								Uprn: "u",
								Paf: &addressv1.Address_PAF{
									Organisation:            "o",
									Department:              "d",
									SubBuilding:             "sb",
									BuildingName:            "bn",
									BuildingNumber:          "bn",
									DependentThoroughfare:   "dt",
									Thoroughfare:            "t",
									DoubleDependentLocality: "ddl",
									DependentLocality:       "dl",
									PostTown:                "pt",
									Postcode:                "E2 1ZZ",
								},
							},
							ContactDetails: &bookingv1.ContactDetails{
								Title:     "Mr",
								FirstName: "John",
								LastName:  "Dough",
								Phone:     "555-0145",
								Email:     "jdough@example.com",
							},
							Slot: &bookingv1.BookingSlot{
								Date: &date.Date{
									Year:  2023,
									Month: 8,
									Day:   27,
								},
								StartTime: 9,
								EndTime:   15,
							},
							VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
								Vulnerabilities: []bookingv1.Vulnerability{
									bookingv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
								},
								Other: "",
							},
							Status:            bookingv1.BookingStatus_BOOKING_STATUS_COMPLETED,
							ExternalReference: "booking-reference-1",
						},
					},
				},
				err: nil,
			},
		},
		{
			description: "should return nil event because booking call was unsuccessful",
			input: inputParams{
				params: domain.CreateBookingParams{
					AccountID: "account-id-1",
					ContactDetails: models.AccountDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Dough",
						Email:     "jdough@example.com",
						Mobile:    "555-0145",
					},
					Slot: models.BookingSlot{
						Date:      mustDate(t, "2023-08-27"),
						StartTime: 9,
						EndTime:   15,
					},
					VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
						Vulnerabilities: []bookingv1.Vulnerability{
							bookingv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
						},
						Other: "",
					},
				},
			},
			setup: func(ctx context.Context, oSt *mocks.MockOccupancyStore, lbGw *mocks.MockLowriBeckGateway) {

				oSt.EXPECT().GetSiteExternalReferenceByAccountID(ctx, "account-id-1").Return(
					&models.Site{
						SiteID:                  "site-id-1",
						Postcode:                "E2 1ZZ",
						UPRN:                    "u",
						BuildingNameNumber:      "bn",
						SubBuildingNameNumber:   "sb",
						DependentThoroughfare:   "dt",
						Thoroughfare:            "t",
						DoubleDependentLocality: "ddl",
						DependentLocality:       "dl",
						Locality:                "l",
						County:                  "c",
						Town:                    "pt",
						Department:              "d",
						Organisation:            "o",
						PoBox:                   "po",
						DeliveryPointSuffix:     "dps",
					},
					&models.OccupancyEligibility{
						OccupancyID: "occupancy-id-1",
						Reference:   "booking-reference-1",
					}, nil)

				lbGw.EXPECT().CreateBooking(ctx, "E2 1ZZ", "booking-reference-1", models.BookingSlot{
					Date:      mustDate(t, "2023-08-27"),
					StartTime: 9,
					EndTime:   15,
				}, models.AccountDetails{
					Title:     "Mr",
					FirstName: "John",
					LastName:  "Dough",
					Email:     "jdough@example.com",
					Mobile:    "555-0145",
				}, []lowribeckv1.Vulnerability{
					lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
				}, "").Return(gateway.CreateBookingResponse{
					Success: false,
				}, nil)

			},
			output: outputParams{
				event: domain.CreateBookingResponse{
					Event: emptyMsg,
				},
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, occSt, lbGw)

			actual, err := myDomain.CreateBooking(ctx, tc.input.params)

			if tc.output.err != nil {
				if !errors.Is(err, tc.output.err) {
					t.Fatalf("expected: %s, actual: %s", err, tc.output.err)
				}
			}

			if diff := cmp.Diff(actual, tc.output.event, cmpopts.IgnoreUnexported(date.Date{}, bookingv1.BookingCreatedEvent{}, bookingv1.Booking{}, addressv1.Address{}, addressv1.Address_PAF{},
				bookingv1.ContactDetails{}, bookingv1.BookingSlot{}, bookingv1.VulnerabilityDetails{}), cmpopts.IgnoreFields(bookingv1.BookingCreatedEvent{}, "BookingId"), cmpopts.IgnoreFields(bookingv1.Booking{}, "Id")); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_RescheduleBooking(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	accGw := mocks.NewMockAccountGateway(ctrl)
	lbGw := mocks.NewMockLowriBeckGateway(ctrl)
	occSt := mocks.NewMockOccupancyStore(ctrl)
	siteSt := mocks.NewMockSiteStore(ctrl)
	bookingSt := mocks.NewMockBookingStore(ctrl)
	partialBookingSt := mocks.NewMockPartialBookingStore(ctrl)

	myDomain := domain.NewBookingDomain(accGw, lbGw, occSt, siteSt, bookingSt, partialBookingSt, false)

	type inputParams struct {
		params domain.RescheduleBookingParams
	}

	type outputParams struct {
		event domain.RescheduleBookingResponse
		err   error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, oSt *mocks.MockOccupancyStore, lbGw *mocks.MockLowriBeckGateway)
		input       inputParams
		output      outputParams
	}

	var emptyMsg *bookingv1.BookingRescheduledEvent

	testCases := []testSetup{
		{
			description: "should reschedule booking",
			input: inputParams{
				params: domain.RescheduleBookingParams{
					AccountID: "account-id-1",
					BookingID: "booking-id-1",
					Slot: models.BookingSlot{
						Date:      mustDate(t, "2023-08-27"),
						StartTime: 9,
						EndTime:   15,
					},
				},
			},
			setup: func(ctx context.Context, oSt *mocks.MockOccupancyStore, lbGw *mocks.MockLowriBeckGateway) {

				oSt.EXPECT().GetSiteExternalReferenceByAccountID(ctx, "account-id-1").Return(
					&models.Site{
						SiteID:                  "site-id-1",
						Postcode:                "E2 1ZZ",
						UPRN:                    "u",
						BuildingNameNumber:      "bn",
						SubBuildingNameNumber:   "sb",
						DependentThoroughfare:   "dt",
						Thoroughfare:            "t",
						DoubleDependentLocality: "ddl",
						DependentLocality:       "dl",
						Locality:                "l",
						County:                  "c",
						Town:                    "pt",
						Department:              "d",
						Organisation:            "o",
						PoBox:                   "po",
						DeliveryPointSuffix:     "dps",
					},
					&models.OccupancyEligibility{
						OccupancyID: "occupancy-id-1",
						Reference:   "booking-reference-1",
					}, nil)

				bookingSt.EXPECT().GetBookingByBookingID(ctx, "booking-id-1").Return(models.Booking{
					BookingID:   "booking-id-1",
					AccountID:   "account-id-1",
					Status:      bookingv1.BookingStatus_BOOKING_STATUS_COMPLETED,
					OccupancyID: "occupancy-id-1",
					Contact: models.AccountDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Dough",
						Email:     "jdough@example.com",
						Mobile:    "555-0145",
					},
					Slot: models.BookingSlot{
						Date:      time.Time{},
						StartTime: 1,
						EndTime:   1,
					},
					VulnerabilityDetails: models.VulnerabilityDetails{
						Vulnerabilities: models.Vulnerabilities{
							bookingv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
						},
						Other: "Bad Knee",
					},
					BookingReference: "booking-reference-1",
				}, nil)

				lbGw.EXPECT().CreateBooking(ctx, "E2 1ZZ", "booking-reference-1", models.BookingSlot{
					Date:      mustDate(t, "2023-08-27"),
					StartTime: 9,
					EndTime:   15,
				}, models.AccountDetails{
					Title:     "Mr",
					FirstName: "John",
					LastName:  "Dough",
					Email:     "jdough@example.com",
					Mobile:    "555-0145",
				}, []lowribeckv1.Vulnerability{
					lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
				}, "Bad Knee").Return(gateway.CreateBookingResponse{
					Success: true,
				}, nil)

			},
			output: outputParams{
				event: domain.RescheduleBookingResponse{
					Event: &bookingv1.BookingRescheduledEvent{
						BookingId: "my-uuid",
						Slot: &bookingv1.BookingSlot{
							Date: &date.Date{
								Year:  2023,
								Month: 8,
								Day:   27,
							},
							StartTime: 9,
							EndTime:   15,
						},
					},
				},
				err: nil,
			},
		},
		{
			description: "should return error from rescheduling",
			input: inputParams{
				params: domain.RescheduleBookingParams{
					AccountID: "account-id-1",
					BookingID: "booking-id-1",
					Slot: models.BookingSlot{
						Date:      mustDate(t, "2023-08-27"),
						StartTime: 9,
						EndTime:   15,
					},
				},
			},
			setup: func(ctx context.Context, oSt *mocks.MockOccupancyStore, lbGw *mocks.MockLowriBeckGateway) {

				oSt.EXPECT().GetSiteExternalReferenceByAccountID(ctx, "account-id-1").Return(
					&models.Site{
						SiteID:                  "site-id-1",
						Postcode:                "E2 1ZZ",
						UPRN:                    "u",
						BuildingNameNumber:      "bn",
						SubBuildingNameNumber:   "sb",
						DependentThoroughfare:   "dt",
						Thoroughfare:            "t",
						DoubleDependentLocality: "ddl",
						DependentLocality:       "dl",
						Locality:                "l",
						County:                  "c",
						Town:                    "pt",
						Department:              "d",
						Organisation:            "o",
						PoBox:                   "po",
						DeliveryPointSuffix:     "dps",
					},
					&models.OccupancyEligibility{
						OccupancyID: "occupancy-id-1",
						Reference:   "booking-reference-1",
					}, nil)

				bookingSt.EXPECT().GetBookingByBookingID(ctx, "booking-id-1").Return(models.Booking{
					BookingID:   "booking-id-1",
					AccountID:   "account-id-1",
					Status:      bookingv1.BookingStatus_BOOKING_STATUS_COMPLETED,
					OccupancyID: "occupancy-id-1",
					Contact: models.AccountDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Dough",
						Email:     "jdough@example.com",
						Mobile:    "555-0145",
					},
					Slot: models.BookingSlot{
						Date:      time.Time{},
						StartTime: 1,
						EndTime:   1,
					},
					VulnerabilityDetails: models.VulnerabilityDetails{
						Vulnerabilities: models.Vulnerabilities{
							bookingv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
						},
						Other: "Bad Knee",
					},
				}, nil)

				lbGw.EXPECT().CreateBooking(ctx, "E2 1ZZ", "booking-reference-1", models.BookingSlot{
					Date:      mustDate(t, "2023-08-27"),
					StartTime: 9,
					EndTime:   15,
				}, models.AccountDetails{
					Title:     "Mr",
					FirstName: "John",
					LastName:  "Dough",
					Email:     "jdough@example.com",
					Mobile:    "555-0145",
				}, []lowribeckv1.Vulnerability{
					lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
				}, "Bad Knee").Return(gateway.CreateBookingResponse{
					Success: false,
				}, nil)

			},
			output: outputParams{
				event: domain.RescheduleBookingResponse{
					Event: emptyMsg,
				},
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, occSt, lbGw)

			actual, err := myDomain.RescheduleBooking(ctx, tc.input.params)

			if tc.output.err != nil {
				if !errors.Is(err, tc.output.err) {
					t.Fatalf("expected: %s, actual: %s", err, tc.output.err)
				}
			}

			if diff := cmp.Diff(actual, tc.output.event, cmpopts.IgnoreUnexported(date.Date{}, bookingv1.BookingRescheduledEvent{}, bookingv1.Booking{}, addressv1.Address{}, addressv1.Address_PAF{},
				bookingv1.ContactDetails{}, bookingv1.BookingSlot{}, bookingv1.VulnerabilityDetails{}), cmpopts.IgnoreFields(bookingv1.BookingRescheduledEvent{}, "BookingId")); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

// Point Of Sale Journey
func Test_GetPOSAvailableSlots(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	accGw := mocks.NewMockAccountGateway(ctrl)
	lbGw := mocks.NewMockLowriBeckGateway(ctrl)
	bookingSt := mocks.NewMockBookingStore(ctrl)

	myDomain := domain.NewBookingDomain(accGw, lbGw, nil, nil, bookingSt, nil, false)

	type inputParams struct {
		params domain.GetPOSAvailableSlotsParams
	}

	type outputParams struct {
		output domain.GetAvailableSlotsResponse
		err    error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, lbGw *mocks.MockLowriBeckGateway)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the available slots between From and To",
			input: inputParams{
				params: domain.GetPOSAvailableSlotsParams{
					Postcode:          "E2 1ZZ",
					Mpan:              "mpan-1",
					TariffElectricity: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
					From: &date.Date{
						Year:  2023,
						Month: 12,
						Day:   1,
					},
					To: &date.Date{
						Year:  2023,
						Month: 12,
						Day:   30,
					},
				},
			},
			setup: func(ctx context.Context, lbGw *mocks.MockLowriBeckGateway) {

				lbGw.EXPECT().GetAvailableSlotsPointOfSale(ctx,
					"E2 1ZZ",
					"mpan-1",
					"",
					lowribeckv1.TariffType_TARIFF_TYPE_CREDIT,
					lowribeckv1.TariffType_TARIFF_TYPE_UNKNOWN,
				).Return(gateway.AvailableSlotsResponse{
					BookingSlots: []models.BookingSlot{
						{
							Date:      mustDate(t, "2023-12-05"),
							StartTime: 9,
							EndTime:   12,
						},
						{
							Date:      mustDate(t, "2023-11-05"),
							StartTime: 17,
							EndTime:   19,
						},
						{
							Date:      mustDate(t, "2023-12-10"),
							StartTime: 11,
							EndTime:   15,
						},
					},
				}, nil)

			},
			output: outputParams{
				output: domain.GetAvailableSlotsResponse{
					Slots: []models.BookingSlot{
						{
							Date:      mustDate(t, "2023-12-05"),
							StartTime: 9,
							EndTime:   12,
						},
						{
							Date:      mustDate(t, "2023-12-10"),
							StartTime: 11,
							EndTime:   15,
						},
					},
				},
				err: nil,
			},
		},
		{
			description: "should return an error because no dates available for provided date",
			input: inputParams{
				params: domain.GetPOSAvailableSlotsParams{
					Postcode:          "E2 1ZZ",
					Mpan:              "mpan-1",
					TariffElectricity: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
					From: &date.Date{
						Year:  2020,
						Month: 12,
						Day:   1,
					},
					To: &date.Date{
						Year:  2020,
						Month: 12,
						Day:   30,
					},
				},
			},
			setup: func(ctx context.Context, lbGw *mocks.MockLowriBeckGateway) {

				lbGw.EXPECT().GetAvailableSlotsPointOfSale(ctx,
					"E2 1ZZ",
					"mpan-1",
					"",
					lowribeckv1.TariffType_TARIFF_TYPE_CREDIT,
					lowribeckv1.TariffType_TARIFF_TYPE_UNKNOWN,
				).Return(gateway.AvailableSlotsResponse{
					BookingSlots: []models.BookingSlot{
						{
							Date:      mustDate(t, "2023-12-05"),
							StartTime: 9,
							EndTime:   12,
						},
						{
							Date:      mustDate(t, "2023-11-05"),
							StartTime: 17,
							EndTime:   19,
						},
						{
							Date:      mustDate(t, "2023-12-10"),
							StartTime: 11,
							EndTime:   15,
						},
					},
				}, nil)

			},
			output: outputParams{
				output: domain.GetAvailableSlotsResponse{
					Slots: []models.BookingSlot{},
				},
				err: domain.ErrNoAvailableSlotsForProvidedDates,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, lbGw)

			actual, err := myDomain.GetAvailableSlotsPointOfSale(ctx, tc.input.params)

			if tc.output.err != nil {
				if !errors.Is(err, tc.output.err) {
					t.Fatalf("expected: %s, actual: %s", err, tc.output.err)
				}
			}

			if diff := cmp.Diff(actual, tc.output.output, cmpopts.IgnoreUnexported(date.Date{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_CreatePOSBooking(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	accGw := mocks.NewMockAccountGateway(ctrl)
	lbGw := mocks.NewMockLowriBeckGateway(ctrl)
	bookingSt := mocks.NewMockBookingStore(ctrl)
	occupancySt := mocks.NewMockOccupancyStore(ctrl)
	partialBookingSt := mocks.NewMockPartialBookingStore(ctrl)

	myDomain := domain.NewBookingDomain(accGw, lbGw, occupancySt, nil, bookingSt, partialBookingSt, false)

	type inputParams struct {
		params domain.CreatePOSBookingParams
	}

	type outputParams struct {
		event domain.CreateBookingResponse
		err   error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, lbGw *mocks.MockLowriBeckGateway, occupancySt *mocks.MockOccupancyStore, partialBookingSt *mocks.MockPartialBookingStore)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should create booking",
			input: inputParams{
				params: domain.CreatePOSBookingParams{
					AccountID:         "account-id-1",
					Postcode:          "E2 1ZZ",
					Mpan:              "mpan-1",
					TariffElectricity: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
					ContactDetails: models.AccountDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Dough",
						Email:     "jdough@example.com",
						Mobile:    "555-0145",
					},
					Slot: models.BookingSlot{
						Date:      mustDate(t, "2023-08-27"),
						StartTime: 9,
						EndTime:   15,
					},
					VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
						Vulnerabilities: []bookingv1.Vulnerability{
							bookingv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
						},
						Other: "",
					},
					Source: bookingv1.BookingSource_BOOKING_SOURCE_PLATFORM_APP,
				},
			},
			setup: func(ctx context.Context, lbGw *mocks.MockLowriBeckGateway, occupancySt *mocks.MockOccupancyStore, partialBookingSt *mocks.MockPartialBookingStore) {

				lbGw.EXPECT().CreateBookingPointOfSale(ctx,
					"E2 1ZZ",
					"mpan-1",
					"",
					lowribeckv1.TariffType_TARIFF_TYPE_CREDIT,
					lowribeckv1.TariffType_TARIFF_TYPE_UNKNOWN,
					models.BookingSlot{
						Date:      mustDate(t, "2023-08-27"),
						StartTime: 9,
						EndTime:   15,
					}, models.AccountDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Dough",
						Email:     "jdough@example.com",
						Mobile:    "555-0145",
					}, []lowribeckv1.Vulnerability{
						lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
					}, "").Return(gateway.CreateBookingPointOfSaleResponse{
					Success:     true,
					ReferenceID: "test-ref",
				}, nil)

				occupancySt.EXPECT().GetOccupancyByAccountID(ctx, "account-id-1").Return(&models.Occupancy{
					OccupancyID: "occ-id-1",
					AccountID:   "account-id-1",
				}, nil)
			},
			output: outputParams{
				event: domain.CreateBookingResponse{
					Event: &bookingv1.BookingCreatedEvent{
						BookingId:     "my-uuid",
						BookingSource: bookingv1.BookingSource_BOOKING_SOURCE_PLATFORM_APP,
						Details: &bookingv1.Booking{
							Id:        "my-uuid",
							AccountId: "account-id-1",
							ContactDetails: &bookingv1.ContactDetails{
								Title:     "Mr",
								FirstName: "John",
								LastName:  "Dough",
								Phone:     "555-0145",
								Email:     "jdough@example.com",
							},
							Slot: &bookingv1.BookingSlot{
								Date: &date.Date{
									Year:  2023,
									Month: 8,
									Day:   27,
								},
								StartTime: 9,
								EndTime:   15,
							},
							VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
								Vulnerabilities: []bookingv1.Vulnerability{
									bookingv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
								},
								Other: "",
							},
							Status:            bookingv1.BookingStatus_BOOKING_STATUS_COMPLETED,
							ExternalReference: "test-ref",
						},
						OccupancyId: "occ-id-1",
					},
				},
				err: nil,
			},
		},
		{
			description: "should create booking but not return an event to be published because occupancy id does not exist yet",
			input: inputParams{
				params: domain.CreatePOSBookingParams{
					AccountID:         "account-id-1",
					Postcode:          "E2 1ZZ",
					Mpan:              "mpan-1",
					TariffElectricity: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
					ContactDetails: models.AccountDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Dough",
						Email:     "jdough@example.com",
						Mobile:    "555-0145",
					},
					Slot: models.BookingSlot{
						Date:      mustDate(t, "2023-08-27"),
						StartTime: 9,
						EndTime:   15,
					},
					VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
						Vulnerabilities: []bookingv1.Vulnerability{
							bookingv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
						},
						Other: "",
					},
					Source: bookingv1.BookingSource_BOOKING_SOURCE_PLATFORM_APP,
				},
			},
			setup: func(ctx context.Context, lbGw *mocks.MockLowriBeckGateway, occupancySt *mocks.MockOccupancyStore, partialBookingSt *mocks.MockPartialBookingStore) {

				lbGw.EXPECT().CreateBookingPointOfSale(ctx,
					"E2 1ZZ",
					"mpan-1",
					"",
					lowribeckv1.TariffType_TARIFF_TYPE_CREDIT,
					lowribeckv1.TariffType_TARIFF_TYPE_UNKNOWN,
					models.BookingSlot{
						Date:      mustDate(t, "2023-08-27"),
						StartTime: 9,
						EndTime:   15,
					}, models.AccountDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Dough",
						Email:     "jdough@example.com",
						Mobile:    "555-0145",
					}, []lowribeckv1.Vulnerability{
						lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
					}, "").Return(gateway.CreateBookingPointOfSaleResponse{
					Success:     true,
					ReferenceID: "test-ref",
				}, nil)

				occupancySt.EXPECT().GetOccupancyByAccountID(ctx, "account-id-1").Return(nil, store.ErrOccupancyNotFound)

				partialBookingSt.EXPECT().Upsert(ctx, gomock.Any(), gomock.Any()).Return(nil)
			},
			output: outputParams{
				event: domain.CreateBookingResponse{
					Event: &bookingv1.BookingCreatedEvent{
						BookingId:     "my-uuid",
						BookingSource: bookingv1.BookingSource_BOOKING_SOURCE_PLATFORM_APP,
						Details: &bookingv1.Booking{
							Id:        "my-uuid",
							AccountId: "account-id-1",
							ContactDetails: &bookingv1.ContactDetails{
								Title:     "Mr",
								FirstName: "John",
								LastName:  "Dough",
								Phone:     "555-0145",
								Email:     "jdough@example.com",
							},
							Slot: &bookingv1.BookingSlot{
								Date: &date.Date{
									Year:  2023,
									Month: 8,
									Day:   27,
								},
								StartTime: 9,
								EndTime:   15,
							},
							VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
								Vulnerabilities: []bookingv1.Vulnerability{
									bookingv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
								},
								Other: "",
							},
							Status:            bookingv1.BookingStatus_BOOKING_STATUS_COMPLETED,
							ExternalReference: "test-ref",
						},
						OccupancyId: "",
					},
				},
				err: nil,
			},
		},
		{
			description: "should return nil event because booking call was unsuccessful",
			input: inputParams{
				params: domain.CreatePOSBookingParams{
					AccountID:         "account-id-1",
					Postcode:          "E2 1ZZ",
					Mpan:              "mpan-1",
					TariffElectricity: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
					ContactDetails: models.AccountDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Dough",
						Email:     "jdough@example.com",
						Mobile:    "555-0145",
					},
					Slot: models.BookingSlot{
						Date:      mustDate(t, "2023-08-27"),
						StartTime: 9,
						EndTime:   15,
					},
					VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
						Vulnerabilities: []bookingv1.Vulnerability{
							bookingv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
						},
						Other: "",
					},
				},
			},
			setup: func(ctx context.Context, lbGw *mocks.MockLowriBeckGateway, occupancySt *mocks.MockOccupancyStore, partialBookingSt *mocks.MockPartialBookingStore) {

				lbGw.EXPECT().CreateBookingPointOfSale(ctx,
					"E2 1ZZ",
					"mpan-1",
					"",
					lowribeckv1.TariffType_TARIFF_TYPE_CREDIT,
					lowribeckv1.TariffType_TARIFF_TYPE_UNKNOWN,
					models.BookingSlot{
						Date:      mustDate(t, "2023-08-27"),
						StartTime: 9,
						EndTime:   15,
					}, models.AccountDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Dough",
						Email:     "jdough@example.com",
						Mobile:    "555-0145",
					}, []lowribeckv1.Vulnerability{
						lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
					}, "").Return(gateway.CreateBookingPointOfSaleResponse{
					Success:     false,
					ReferenceID: "test-ref",
				}, nil)

			},
			output: outputParams{
				event: domain.CreateBookingResponse{
					Event: nil,
				},
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, lbGw, occupancySt, partialBookingSt)

			actual, err := myDomain.CreateBookingPointOfSale(ctx, tc.input.params)

			if tc.output.err != nil {
				if !errors.Is(err, tc.output.err) {
					t.Fatalf("expected: %s, actual: %s", err, tc.output.err)
				}
			}

			if diff := cmp.Diff(actual, tc.output.event, cmpopts.IgnoreUnexported(date.Date{}, bookingv1.BookingCreatedEvent{}, bookingv1.Booking{}, addressv1.Address{}, addressv1.Address_PAF{},
				bookingv1.ContactDetails{}, bookingv1.BookingSlot{}, bookingv1.VulnerabilityDetails{}), cmpopts.IgnoreFields(bookingv1.BookingCreatedEvent{}, "BookingId"), cmpopts.IgnoreFields(bookingv1.Booking{}, "Id")); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
