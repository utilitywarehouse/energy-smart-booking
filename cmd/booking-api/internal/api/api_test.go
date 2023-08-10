//go:generate mockgen -source=api.go -destination ./mocks/api_mocks.go

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	addressv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/address/v1"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"google.golang.org/genproto/googleapis/type/date"

	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/api"
	mocks "github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/api/mocks"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

func Test_GetCustomerContactDetails(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	bookingDomain := mocks.NewMockBookingDomain(ctrl)
	bookingPublisher := mocks.NewMockBookingPublisher(ctrl)

	myAPIHandler := api.New(bookingDomain, bookingPublisher)

	type inputParams struct {
		req *bookingv1.GetCustomerContactDetailsRequest
	}

	type outputParams struct {
		res *bookingv1.GetCustomerContactDetailsResponse
		err error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, domain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the account details by account id",
			input: inputParams{
				req: &bookingv1.GetCustomerContactDetailsRequest{
					AccountId: "account-id-1",
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher) {

				bkDomain.EXPECT().GetCustomerContactDetails(ctx, "account-id-1").Return(models.Account{
					AccountID: "account-id-1",
					Details: models.AccountDetails{
						Title:     "Mr",
						FirstName: "Jon",
						LastName:  "Dough",
						Email:     "jdough@example.com",
						Mobile:    "555-0555",
					},
				}, nil)

			},
			output: outputParams{
				res: &bookingv1.GetCustomerContactDetailsResponse{
					Title:     "Mr",
					FirstName: "Jon",
					LastName:  "Dough",
					Phone:     "555-0555",
					Email:     "jdough@example.com",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, bookingDomain, bookingPublisher)

			expected, err := myAPIHandler.GetCustomerContactDetails(ctx, tc.input.req)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(expected, tc.output.res, cmpopts.IgnoreUnexported(date.Date{}, bookingv1.GetCustomerContactDetailsResponse{}, bookingv1.BookingSlot{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_GetCustomerSiteAddress(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	bookingDomain := mocks.NewMockBookingDomain(ctrl)
	bookingPublisher := mocks.NewMockBookingPublisher(ctrl)

	myAPIHandler := api.New(bookingDomain, bookingPublisher)

	type inputParams struct {
		req *bookingv1.GetCustomerSiteAddressRequest
	}

	type outputParams struct {
		res *bookingv1.GetCustomerSiteAddressResponse
		err error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, domain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the account details by account id",
			input: inputParams{
				req: &bookingv1.GetCustomerSiteAddressRequest{
					AccountId: "account-id-1",
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher) {

				bkDomain.EXPECT().GetAccountAddressByAccountID(ctx, "account-id-1").Return(models.AccountAddress{
					UPRN: "uprn",
					PAF: models.PAF{
						BuildingName:            "bn",
						BuildingNumber:          "bnu",
						Department:              "d",
						DependentLocality:       "dl",
						DependentThoroughfare:   "dt",
						DoubleDependentLocality: "ddl",
						Organisation:            "o",
						PostTown:                "pt",
						Postcode:                "pc",
						SubBuilding:             "sb",
						Thoroughfare:            "tf",
					},
				}, nil)

			},
			output: outputParams{
				res: &bookingv1.GetCustomerSiteAddressResponse{
					SiteAddress: &addressv1.Address{
						Uprn: "uprn",
						Paf: &addressv1.Address_PAF{
							Organisation:            "o",
							Department:              "d",
							SubBuilding:             "sb",
							BuildingName:            "bn",
							BuildingNumber:          "bnu",
							DependentThoroughfare:   "dt",
							Thoroughfare:            "tf",
							DoubleDependentLocality: "ddl",
							DependentLocality:       "dl",
							PostTown:                "pt",
							Postcode:                "pc",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, bookingDomain, bookingPublisher)

			expected, err := myAPIHandler.GetCustomerSiteAddress(ctx, tc.input.req)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(expected, tc.output.res, cmpopts.IgnoreUnexported(date.Date{}, addressv1.Address{}, bookingv1.BookingSlot{}, bookingv1.GetCustomerSiteAddressResponse{}, addressv1.Address_PAF{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_GetCustomerBookings(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	bookingDomain := mocks.NewMockBookingDomain(ctrl)
	bookingPublisher := mocks.NewMockBookingPublisher(ctrl)

	myAPIHandler := api.New(bookingDomain, bookingPublisher)

	type inputParams struct {
		req *bookingv1.GetCustomerBookingsRequest
	}

	type outputParams struct {
		res *bookingv1.GetCustomerBookingsResponse
		err error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, domain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the account details by account id",
			input: inputParams{
				req: &bookingv1.GetCustomerBookingsRequest{
					AccountId: "account-id-1",
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher) {

				bkDomain.EXPECT().GetCustomerBookings(ctx, "account-id-1").Return([]*bookingv1.Booking{
					{
						Id:        "booking-id-1",
						AccountId: "account-id-1",
						SiteAddress: &addressv1.Address{
							Uprn: "uprn",
							Paf: &addressv1.Address_PAF{
								Organisation:            "o",
								Department:              "d",
								SubBuilding:             "sb",
								BuildingName:            "bn",
								BuildingNumber:          "bnu",
								DependentThoroughfare:   "dt",
								Thoroughfare:            "tf",
								DoubleDependentLocality: "ddl",
								DependentLocality:       "dl",
								PostTown:                "pt",
								Postcode:                "pc",
							},
						},
						ContactDetails: &bookingv1.ContactDetails{
							Title:     "Mr",
							FirstName: "John",
							LastName:  "Doe",
							Phone:     "555-0555",
							Email:     "jdoe@example.com",
						},
						Slot: &bookingv1.BookingSlot{
							Date: &date.Date{
								Year:  2020,
								Month: 12,
								Day:   12,
							},
							StartTime: 15,
							EndTime:   20,
						},
						VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
							Vulnerabilities: []bookingv1.Vulnerability{
								bookingv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
							},
							Other: "Bad Knee",
						},
						Status: bookingv1.BookingStatus_BOOKING_STATUS_COMPLETED,
					},
				}, nil)

			},
			output: outputParams{
				res: &bookingv1.GetCustomerBookingsResponse{
					Bookings: []*bookingv1.Booking{
						{
							Id:        "booking-id-1",
							AccountId: "account-id-1",
							SiteAddress: &addressv1.Address{
								Uprn: "uprn",
								Paf: &addressv1.Address_PAF{
									Organisation:            "o",
									Department:              "d",
									SubBuilding:             "sb",
									BuildingName:            "bn",
									BuildingNumber:          "bnu",
									DependentThoroughfare:   "dt",
									Thoroughfare:            "tf",
									DoubleDependentLocality: "ddl",
									DependentLocality:       "dl",
									PostTown:                "pt",
									Postcode:                "pc",
								},
							},
							ContactDetails: &bookingv1.ContactDetails{
								Title:     "Mr",
								FirstName: "John",
								LastName:  "Doe",
								Phone:     "555-0555",
								Email:     "jdoe@example.com",
							},
							Slot: &bookingv1.BookingSlot{
								Date: &date.Date{
									Year:  2020,
									Month: 12,
									Day:   12,
								},
								StartTime: 15,
								EndTime:   20,
							},
							VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
								Vulnerabilities: []bookingv1.Vulnerability{
									bookingv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
								},
								Other: "Bad Knee",
							},
							Status: bookingv1.BookingStatus_BOOKING_STATUS_COMPLETED,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, bookingDomain, bookingPublisher)

			expected, err := myAPIHandler.GetCustomerBookings(ctx, tc.input.req)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(expected, tc.output.res, cmpopts.IgnoreUnexported(date.Date{}, addressv1.Address{}, bookingv1.BookingSlot{}, bookingv1.Booking{}, bookingv1.ContactDetails{},
				bookingv1.VulnerabilityDetails{}, bookingv1.GetCustomerBookingsResponse{}, addressv1.Address_PAF{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_GetAvailableSlot(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	bookingDomain := mocks.NewMockBookingDomain(ctrl)
	bookingPublisher := mocks.NewMockBookingPublisher(ctrl)

	myAPIHandler := api.New(bookingDomain, bookingPublisher)

	type inputParams struct {
		req *bookingv1.GetAvailableSlotsRequest
	}

	type outputParams struct {
		res *bookingv1.GetAvailableSlotsResponse
		err error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, domain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the account details by account id",
			input: inputParams{
				req: &bookingv1.GetAvailableSlotsRequest{
					AccountId: "account-id-1",
					From: &date.Date{
						Year:  2012,
						Month: 12,
						Day:   21,
					},
					To: &date.Date{
						Year:  2022,
						Month: 02,
						Day:   12,
					},
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher) {

				params := domain.GetAvailableSlotsParams{
					AccountID: "account-id-1",
					From: &date.Date{
						Year:  2012,
						Month: 12,
						Day:   21,
					},
					To: &date.Date{
						Year:  2022,
						Month: 02,
						Day:   12,
					},
				}

				bkDomain.EXPECT().GetAvailableSlots(ctx, params).Return(domain.GetAvailableSlotsResponse{
					Slots: []models.BookingSlot{
						{
							Date:      time.Date(2021, time.August, 1, 0, 0, 0, 0, time.UTC),
							StartTime: 12,
							EndTime:   18,
						},
						{
							Date:      time.Date(2021, time.August, 2, 0, 0, 0, 0, time.UTC),
							StartTime: 12,
							EndTime:   18,
						},
						{
							Date:      time.Date(2021, time.August, 3, 0, 0, 0, 0, time.UTC),
							StartTime: 12,
							EndTime:   18,
						},
						{
							Date:      time.Date(2021, time.August, 4, 0, 0, 0, 0, time.UTC),
							StartTime: 12,
							EndTime:   18,
						},
					},
				}, nil)

			},
			output: outputParams{
				res: &bookingv1.GetAvailableSlotsResponse{
					Slots: []*bookingv1.BookingSlot{
						{
							Date: &date.Date{
								Year:  2021,
								Month: 8,
								Day:   1,
							},
							StartTime: 12,
							EndTime:   18,
						},
						{
							Date: &date.Date{
								Year:  2021,
								Month: 8,
								Day:   2,
							},
							StartTime: 12,
							EndTime:   18,
						},
						{
							Date: &date.Date{
								Year:  2021,
								Month: 8,
								Day:   3,
							},
							StartTime: 12,
							EndTime:   18,
						},
						{
							Date: &date.Date{
								Year:  2021,
								Month: 8,
								Day:   4,
							},
							StartTime: 12,
							EndTime:   18,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, bookingDomain, bookingPublisher)

			expected, err := myAPIHandler.GetAvailableSlots(ctx, tc.input.req)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(expected, tc.output.res, cmpopts.IgnoreUnexported(date.Date{}, bookingv1.GetAvailableSlotsResponse{}, bookingv1.Booking{}, addressv1.Address{}, addressv1.Address_PAF{},
				bookingv1.ContactDetails{}, bookingv1.BookingSlot{}, bookingv1.VulnerabilityDetails{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_CreateBooking(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	bookingDomain := mocks.NewMockBookingDomain(ctrl)
	bookingPublisher := mocks.NewMockBookingPublisher(ctrl)

	myAPIHandler := api.New(bookingDomain, bookingPublisher)

	type inputParams struct {
		req *bookingv1.CreateBookingRequest
	}

	type outputParams struct {
		res *bookingv1.CreateBookingResponse
		err error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, domain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the account details by account id",
			input: inputParams{
				req: &bookingv1.CreateBookingRequest{
					AccountId: "account-id-1",
					Slot: &bookingv1.BookingSlot{
						Date: &date.Date{
							Year:  2020,
							Month: 10,
							Day:   10,
						},
						StartTime: 10,
						EndTime:   18,
					},
					VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
						Vulnerabilities: []bookingv1.Vulnerability{
							bookingv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
						},
						Other: "Bad Knee",
					},
					ContactDetails: &bookingv1.ContactDetails{
						Title:     "Mr",
						FirstName: "Joe",
						LastName:  "Dough",
						Phone:     "555-0555",
						Email:     "jd@example.com",
					},
					Platform: bookingv1.Platform_PLATFORM_APP,
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher) {

				params := domain.CreateBookingParams{
					AccountID: "account-id-1",
					ContactDetails: models.AccountDetails{
						Title:     "Mr",
						FirstName: "Joe",
						LastName:  "Dough",
						Mobile:    "555-0555",
						Email:     "jd@example.com",
					},
					Slot: models.BookingSlot{
						Date:      time.Date(2020, time.October, 10, 0, 0, 0, 0, time.UTC),
						StartTime: 10,
						EndTime:   18,
					},
					VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
						Vulnerabilities: []bookingv1.Vulnerability{
							bookingv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
						},
						Other: "Bad Knee",
					},
					Source: bookingv1.BookingSource_BOOKING_SOURCE_PLATFORM_APP,
				}

				bkDomain.EXPECT().CreateBooking(ctx, params).Return(domain.CreateBookingResponse{
					Event: &bookingv1.BookingCreatedEvent{
						BookingId: "booking-id-1",
						Details: &bookingv1.Booking{
							Id:        "booking-id-1",
							AccountId: params.AccountID,
							SiteAddress: &addressv1.Address{
								Uprn: "uprn",
								Paf: &addressv1.Address_PAF{
									Organisation:            "o",
									Department:              "d",
									SubBuilding:             "sb",
									BuildingName:            "bn",
									BuildingNumber:          "bnu",
									DependentThoroughfare:   "dt",
									Thoroughfare:            "tf",
									DoubleDependentLocality: "ddl",
									DependentLocality:       "dl",
									PostTown:                "pt",
									Postcode:                "pc",
								},
							},
							ContactDetails: &bookingv1.ContactDetails{
								Title:     "Mr",
								FirstName: "Joe",
								LastName:  "Dough",
								Phone:     "555-0555",
								Email:     "jd@example.com",
							},
							Slot: &bookingv1.BookingSlot{
								Date: &date.Date{
									Year:  int32(2020),
									Month: int32(10),
									Day:   int32(10),
								},
								StartTime: int32(10),
								EndTime:   int32(10),
							},
							VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
								Vulnerabilities: []bookingv1.Vulnerability{
									bookingv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
								},
								Other: "Bad Knee",
							},
							Status: bookingv1.BookingStatus_BOOKING_STATUS_COMPLETED,
						},
						OccupancyId:   "occupancy-id-1",
						BookingSource: bookingv1.BookingSource_BOOKING_SOURCE_PLATFORM_APP,
					},
				}, nil)

				publisher.EXPECT().Sink(ctx, &bookingv1.BookingCreatedEvent{
					BookingId:     "booking-id-1",
					OccupancyId:   "occupancy-id-1",
					BookingSource: bookingv1.BookingSource_BOOKING_SOURCE_PLATFORM_APP,
					Details: &bookingv1.Booking{
						Id:        "booking-id-1",
						AccountId: params.AccountID,
						SiteAddress: &addressv1.Address{
							Uprn: "uprn",
							Paf: &addressv1.Address_PAF{
								Organisation:            "o",
								Department:              "d",
								SubBuilding:             "sb",
								BuildingName:            "bn",
								BuildingNumber:          "bnu",
								DependentThoroughfare:   "dt",
								Thoroughfare:            "tf",
								DoubleDependentLocality: "ddl",
								DependentLocality:       "dl",
								PostTown:                "pt",
								Postcode:                "pc",
							},
						},
						ContactDetails: &bookingv1.ContactDetails{
							Title:     "Mr",
							FirstName: "Joe",
							LastName:  "Dough",
							Phone:     "555-0555",
							Email:     "jd@example.com",
						},
						Slot: &bookingv1.BookingSlot{
							Date: &date.Date{
								Year:  int32(2020),
								Month: int32(10),
								Day:   int32(10),
							},
							StartTime: int32(10),
							EndTime:   int32(10),
						},
						VulnerabilityDetails: &bookingv1.VulnerabilityDetails{
							Vulnerabilities: []bookingv1.Vulnerability{
								bookingv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
							},
							Other: "Bad Knee",
						},
						Status: bookingv1.BookingStatus_BOOKING_STATUS_COMPLETED,
					},
				}, gomock.Any()).Return(nil)

			},
			output: outputParams{
				res: &bookingv1.CreateBookingResponse{
					BookingId: "booking-id-1",
				},
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, bookingDomain, bookingPublisher)

			expected, err := myAPIHandler.CreateBooking(ctx, tc.input.req)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(expected, tc.output.res, cmpopts.IgnoreUnexported(date.Date{}, bookingv1.CreateBookingResponse{}, bookingv1.Booking{}, addressv1.Address{}, addressv1.Address_PAF{},
				bookingv1.ContactDetails{}, bookingv1.BookingSlot{}, bookingv1.VulnerabilityDetails{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_RescheduleBooking(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	bookingDomain := mocks.NewMockBookingDomain(ctrl)
	bookingPublisher := mocks.NewMockBookingPublisher(ctrl)

	myAPIHandler := api.New(bookingDomain, bookingPublisher)

	type inputParams struct {
		req *bookingv1.RescheduleBookingRequest
	}

	type outputParams struct {
		res *bookingv1.RescheduleBookingResponse
		err error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, domain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the account details by account id",
			input: inputParams{
				req: &bookingv1.RescheduleBookingRequest{
					AccountId: "account-id-1",
					BookingId: "booking-id-1",
					Slot: &bookingv1.BookingSlot{
						Date: &date.Date{
							Year:  2020,
							Month: 1,
							Day:   12,
						},
						StartTime: 10,
						EndTime:   20,
					},
					Platform: bookingv1.Platform_PLATFORM_APP,
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher) {

				params := domain.RescheduleBookingParams{
					AccountID: "account-id-1",
					BookingID: "booking-id-1",
					Slot: models.BookingSlot{
						Date:      time.Date(2020, time.January, 12, 0, 0, 0, 0, time.UTC),
						StartTime: 10,
						EndTime:   20,
					},
					Source: bookingv1.BookingSource_BOOKING_SOURCE_PLATFORM_APP,
				}

				bkDomain.EXPECT().RescheduleBooking(ctx, params).Return(domain.RescheduleBookingResponse{
					Event: &bookingv1.BookingRescheduledEvent{
						BookingId:     "booking-id-1",
						BookingSource: bookingv1.BookingSource_BOOKING_SOURCE_PLATFORM_APP,
						Slot: &bookingv1.BookingSlot{
							Date: &date.Date{
								Year:  2020,
								Month: 1,
								Day:   12,
							},
							StartTime: 10,
							EndTime:   20,
						},
					},
				}, nil)

				publisher.EXPECT().Sink(ctx, &bookingv1.BookingRescheduledEvent{
					BookingId: "booking-id-1",
					Slot: &bookingv1.BookingSlot{
						Date: &date.Date{
							Year:  2020,
							Month: 1,
							Day:   12,
						},
						StartTime: 10,
						EndTime:   20,
					},
					BookingSource: bookingv1.BookingSource_BOOKING_SOURCE_PLATFORM_APP,
				}, gomock.Any()).Return(nil)

			},
			output: outputParams{
				res: &bookingv1.RescheduleBookingResponse{
					BookingId: "booking-id-1",
				},
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, bookingDomain, bookingPublisher)

			expected, err := myAPIHandler.RescheduleBooking(ctx, tc.input.req)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(expected, tc.output.res, cmpopts.IgnoreUnexported(date.Date{}, bookingv1.RescheduleBookingResponse{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
