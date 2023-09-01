//go:generate mockgen -source=api.go -destination ./mocks/api_mocks.go

package api_test

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
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/api"
	mocks "github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/api/mocks"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/internal/auth"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
)

var oops = errors.New("oops...")

func Test_GetCustomerContactDetails(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	bookingDomain := mocks.NewMockBookingDomain(ctrl)
	bookingPublisher := mocks.NewMockBookingPublisher(ctrl)
	mockAuth := mocks.NewMockAuth(ctrl)

	myAPIHandler := api.New(bookingDomain, bookingPublisher, mockAuth)

	type inputParams struct {
		req *bookingv1.GetCustomerContactDetailsRequest
	}

	type outputParams struct {
		res *bookingv1.GetCustomerContactDetailsResponse
		err error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, domain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, auth *mocks.MockAuth)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the account details by account id from a non-customer account",
			input: inputParams{
				req: &bookingv1.GetCustomerContactDetailsRequest{
					AccountId: "account-id-1",
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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
		{
			description: "should get the account details by account id from a customer account",
			input: inputParams{
				req: &bookingv1.GetCustomerContactDetailsRequest{
					AccountId: "account-id-1",
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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
		{
			description: "should fail to get account details because unauthorised",
			input: inputParams{
				req: &bookingv1.GetCustomerContactDetailsRequest{
					AccountId: "account-id-1",
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(false, nil)

			},
			output: outputParams{
				res: nil,
				err: status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", api.ErrUserUnauthorised),
			},
		},
		{
			description: "should fail to get account details because some error occurred in auth checks",
			input: inputParams{
				req: &bookingv1.GetCustomerContactDetailsRequest{
					AccountId: "account-id-1",
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(false, oops)

			},
			output: outputParams{
				res: nil,
				err: status.Errorf(codes.Internal, "failed to validate credentials"),
			},
		},
		{
			description: "should fail to get account details customer requesting not his own",
			input: inputParams{
				req: &bookingv1.GetCustomerContactDetailsRequest{
					AccountId: "account-id-2",
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-2",
				}).Return(false, nil)

			},
			output: outputParams{
				res: nil,
				err: status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", api.ErrUserUnauthorised),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, bookingDomain, bookingPublisher, mockAuth)

			expected, err := myAPIHandler.GetCustomerContactDetails(ctx, tc.input.req)
			if tc.output.err != nil {
				if diff := cmp.Diff(err.Error(), tc.output.err.Error()); diff != "" {
					t.Fatal(diff)
				}
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
	mockAuth := mocks.NewMockAuth(ctrl)

	myAPIHandler := api.New(bookingDomain, bookingPublisher, mockAuth)

	type inputParams struct {
		req *bookingv1.GetCustomerSiteAddressRequest
	}

	type outputParams struct {
		res *bookingv1.GetCustomerSiteAddressResponse
		err error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, domain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get account site address by account id for a non-customer account",
			input: inputParams{
				req: &bookingv1.GetCustomerSiteAddressRequest{
					AccountId: "account-id-1",
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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
		{
			description: "should get account site address by account id for a customer account",
			input: inputParams{
				req: &bookingv1.GetCustomerSiteAddressRequest{
					AccountId: "account-id-1",
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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
		{
			description: "should fail to get account site address because user is not authorised",
			input: inputParams{
				req: &bookingv1.GetCustomerSiteAddressRequest{
					AccountId: "account-id-1",
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(false, nil)

			},
			output: outputParams{
				res: nil,
				err: status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", api.ErrUserUnauthorised),
			},
		},
		{
			description: "should fail to get account site address because customer is trying to access another customer account id",
			input: inputParams{
				req: &bookingv1.GetCustomerSiteAddressRequest{
					AccountId: "account-id-2",
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-2",
				}).Return(false, nil)

			},
			output: outputParams{
				res: nil,
				err: status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", api.ErrUserUnauthorised),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, bookingDomain, bookingPublisher, mockAuth)

			expected, err := myAPIHandler.GetCustomerSiteAddress(ctx, tc.input.req)
			if tc.output.err != nil {
				if diff := cmp.Diff(err.Error(), tc.output.err.Error()); diff != "" {
					t.Fatal(diff)
				}
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
	mockAuth := mocks.NewMockAuth(ctrl)

	myAPIHandler := api.New(bookingDomain, bookingPublisher, mockAuth)

	type inputParams struct {
		req *bookingv1.GetCustomerBookingsRequest
	}

	type outputParams struct {
		res *bookingv1.GetCustomerBookingsResponse
		err error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, domain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the customer bookings on a non-customer requester",
			input: inputParams{
				req: &bookingv1.GetCustomerBookingsRequest{
					AccountId: "account-id-1",
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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
		{
			description: "should error when getting the customer bookings",
			input: inputParams{
				req: &bookingv1.GetCustomerBookingsRequest{
					AccountId: "account-id-1",
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

				bkDomain.EXPECT().GetCustomerBookings(ctx, "account-id-1").Return(nil, oops)
			},
			output: outputParams{
				res: &bookingv1.GetCustomerBookingsResponse{
					Bookings: nil,
				},
				err: status.Errorf(codes.Internal, "failed to get customer bookings, %s", oops),
			},
		},
		{
			description: "should fail to get customer bookings because requester is unauthorised",
			input: inputParams{
				req: &bookingv1.GetCustomerBookingsRequest{
					AccountId: "account-id-1",
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(false, nil)

			},
			output: outputParams{
				res: nil,
				err: status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", api.ErrUserUnauthorised),
			},
		},
		{
			description: "should fail to get customer bookings because customer tries to access another customer resource",
			input: inputParams{
				req: &bookingv1.GetCustomerBookingsRequest{
					AccountId: "account-id-2",
				},
			},
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-2",
				}).Return(false, nil)

			},
			output: outputParams{
				res: nil,
				err: status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", api.ErrUserUnauthorised),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, bookingDomain, bookingPublisher, mockAuth)

			expected, err := myAPIHandler.GetCustomerBookings(ctx, tc.input.req)

			if err != nil {
				if tc.output.err != nil {
					if diff := cmp.Diff(err.Error(), tc.output.err.Error()); diff != "" {
						t.Fatal(diff)
					}
				} else {
					t.Fatal(err)
				}
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
	mockAuth := mocks.NewMockAuth(ctrl)

	myAPIHandler := api.New(bookingDomain, bookingPublisher, mockAuth)

	type inputParams struct {
		req *bookingv1.GetAvailableSlotsRequest
	}

	type outputParams struct {
		res *bookingv1.GetAvailableSlotsResponse
		err error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, domain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth)
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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
		{
			description: "available slots returns a gateway.ErrInvalidArgument error",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

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

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

				bkDomain.EXPECT().GetAvailableSlots(ctx, params).Return(domain.GetAvailableSlotsResponse{
					Slots: []models.BookingSlot{},
				}, gateway.ErrInvalidArgument)

			},
			output: outputParams{
				res: &bookingv1.GetAvailableSlotsResponse{
					Slots: nil,
				},
				err: status.Errorf(codes.Internal, "failed to get available slots, %s", gateway.ErrInvalidArgument.Error()),
			},
		},
		{
			description: "available slots returns a gateway.ErrInternalBadParameters error",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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
					Slots: []models.BookingSlot{},
				}, gateway.ErrInternalBadParameters)

			},
			output: outputParams{
				res: &bookingv1.GetAvailableSlotsResponse{
					Slots: nil,
				},
				err: status.Errorf(codes.Internal, "failed to get available slots, %s", gateway.ErrInternalBadParameters.Error()),
			},
		},
		{
			description: "call to available slots returns an internal error",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

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

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

				bkDomain.EXPECT().GetAvailableSlots(ctx, params).Return(domain.GetAvailableSlotsResponse{
					Slots: []models.BookingSlot{},
				}, gateway.ErrInternal)

			},
			output: outputParams{
				res: &bookingv1.GetAvailableSlotsResponse{
					Slots: nil,
				},
				err: status.Errorf(codes.Internal, "failed to get available slots, %s", gateway.ErrInternal.Error()),
			},
		},
		{
			description: "available slots returns a gateway.ErrNotFound error",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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
					Slots: []models.BookingSlot{},
				}, gateway.ErrNotFound)

			},
			output: outputParams{
				res: &bookingv1.GetAvailableSlotsResponse{
					Slots: nil,
				},
				err: status.Errorf(codes.NotFound, "failed to get available slots, %s", gateway.ErrNotFound.Error()),
			},
		},
		{
			description: "available slots call returns a domain.ErrNoAvailableSlotsForProvidedDates",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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
					Slots: []models.BookingSlot{},
				}, domain.ErrNoAvailableSlotsForProvidedDates)

			},
			output: outputParams{
				res: &bookingv1.GetAvailableSlotsResponse{
					Slots: nil,
				},
				err: status.Errorf(codes.OutOfRange, "failed to get available slots, %s", domain.ErrNoAvailableSlotsForProvidedDates.Error()),
			},
		},
		{
			description: "available slots call returns a gateway.ErrOutOfRange",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

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

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

				bkDomain.EXPECT().GetAvailableSlots(ctx, params).Return(domain.GetAvailableSlotsResponse{
					Slots: []models.BookingSlot{},
				}, gateway.ErrOutOfRange)

			},
			output: outputParams{
				res: &bookingv1.GetAvailableSlotsResponse{
					Slots: nil,
				},
				err: status.Errorf(codes.OutOfRange, "failed to get available slots, %s", gateway.ErrOutOfRange.Error()),
			},
		},
		{
			description: "available slots call returns a generic error",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

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

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

				bkDomain.EXPECT().GetAvailableSlots(ctx, params).Return(domain.GetAvailableSlotsResponse{
					Slots: []models.BookingSlot{},
				}, oops)

			},
			output: outputParams{
				res: &bookingv1.GetAvailableSlotsResponse{
					Slots: nil,
				},
				err: status.Errorf(codes.Internal, "failed to get available slots, %s", oops),
			},
		},
		{
			description: "should fail because requester is unauthorised",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(false, nil)

			},
			output: outputParams{
				res: nil,
				err: status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", api.ErrUserUnauthorised),
			},
		},
		{
			description: "should fail because customer is trying to access another user's account id",
			input: inputParams{
				req: &bookingv1.GetAvailableSlotsRequest{
					AccountId: "account-id-2",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "get",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-2",
				}).Return(false, nil)

			},
			output: outputParams{
				res: nil,
				err: status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", api.ErrUserUnauthorised),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, bookingDomain, bookingPublisher, mockAuth)

			expected, err := myAPIHandler.GetAvailableSlots(ctx, tc.input.req)
			if tc.output.err != nil {
				if diff := cmp.Diff(err.Error(), tc.output.err.Error()); diff != "" {
					t.Fatal(diff)
				}
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
	mockAuth := mocks.NewMockAuth(ctrl)

	myAPIHandler := api.New(bookingDomain, bookingPublisher, mockAuth)

	type inputParams struct {
		req *bookingv1.CreateBookingRequest
	}

	type outputParams struct {
		res *bookingv1.CreateBookingResponse
		err error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, domain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth)
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "create",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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
		{
			description: "create booking call returns a gateway.ErrInvalidArgument",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "create",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().CreateBooking(ctx, params).Return(domain.CreateBookingResponse{}, gateway.ErrInvalidArgument)

			},
			output: outputParams{
				res: &bookingv1.CreateBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.Internal, "failed to create booking, %s", gateway.ErrInvalidArgument.Error()),
			},
		},
		{
			description: "create booking call returns a gateway.ErrInternalBadParameters",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "create",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().CreateBooking(ctx, params).Return(domain.CreateBookingResponse{}, gateway.ErrInternalBadParameters)

			},
			output: outputParams{
				res: &bookingv1.CreateBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.Internal, "failed to create booking, %s", gateway.ErrInternalBadParameters.Error()),
			},
		},
		{
			description: "create booking call returns a gateway.ErrInternal",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "create",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().CreateBooking(ctx, params).Return(domain.CreateBookingResponse{}, gateway.ErrInternal)

			},
			output: outputParams{
				res: &bookingv1.CreateBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.Internal, "failed to create booking, %s", gateway.ErrInternal.Error()),
			},
		},
		{
			description: "create booking call returns a gateway.ErrNotFound",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "create",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().CreateBooking(ctx, params).Return(domain.CreateBookingResponse{}, gateway.ErrNotFound)

			},
			output: outputParams{
				res: &bookingv1.CreateBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.NotFound, "failed to create booking, %s", gateway.ErrNotFound.Error()),
			},
		},
		{
			description: "create booking call returns a gateway.ErrOutOfRange",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "create",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().CreateBooking(ctx, params).Return(domain.CreateBookingResponse{}, gateway.ErrOutOfRange)

			},
			output: outputParams{
				res: &bookingv1.CreateBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.OutOfRange, "failed to create booking, %s", gateway.ErrOutOfRange.Error()),
			},
		},
		{
			description: "create booking call returns a gateway.ErrAlreadyExists",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "create",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().CreateBooking(ctx, params).Return(domain.CreateBookingResponse{}, gateway.ErrAlreadyExists)

			},
			output: outputParams{
				res: &bookingv1.CreateBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.AlreadyExists, "failed to create booking, %s", gateway.ErrAlreadyExists.Error()),
			},
		},
		{
			description: "create booking call returns a gateway.ErrAlreadyExists",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "create",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().CreateBooking(ctx, params).Return(domain.CreateBookingResponse{}, gateway.ErrAlreadyExists)

			},
			output: outputParams{
				res: &bookingv1.CreateBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.AlreadyExists, "failed to create booking, %s", gateway.ErrAlreadyExists.Error()),
			},
		},
		{
			description: "create booking call returns a gateway.ErrInvalidAppointmentDate",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "create",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().CreateBooking(ctx, params).Return(domain.CreateBookingResponse{}, gateway.ErrInvalidAppointmentDate)

			},
			output: outputParams{
				res: &bookingv1.CreateBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.InvalidArgument, "failed to create booking, %s", gateway.ErrInvalidAppointmentDate.Error()),
			},
		},
		{
			description: "create booking call returns a gateway.ErrInvalidAppointmentTime",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "create",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().CreateBooking(ctx, params).Return(domain.CreateBookingResponse{}, gateway.ErrInvalidAppointmentTime)

			},
			output: outputParams{
				res: &bookingv1.CreateBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.InvalidArgument, "failed to create booking, %s", gateway.ErrInvalidAppointmentTime.Error()),
			},
		},
		{
			description: "create booking call returns an error",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "create",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().CreateBooking(ctx, params).Return(domain.CreateBookingResponse{}, oops)

			},
			output: outputParams{
				res: &bookingv1.CreateBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.Internal, "failed to create booking, %s", oops.Error()),
			},
		},
		{
			description: "should fail to create booking because user is unauthorised",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "create",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(false, nil)
			},
			output: outputParams{
				res: nil,
				err: status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", api.ErrUserUnauthorised),
			},
		},
		{
			description: "should fail to create booking because user is trying to access another user",
			input: inputParams{
				req: &bookingv1.CreateBookingRequest{
					AccountId: "account-id-2",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "create",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-2",
				}).Return(false, nil)
			},
			output: outputParams{
				res: nil,
				err: status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", api.ErrUserUnauthorised),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, bookingDomain, bookingPublisher, mockAuth)

			expected, err := myAPIHandler.CreateBooking(ctx, tc.input.req)
			if tc.output.err != nil {
				if diff := cmp.Diff(err.Error(), tc.output.err.Error()); diff != "" {
					t.Fatal(diff)
				}
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
	mockAuth := mocks.NewMockAuth(ctrl)

	myAPIHandler := api.New(bookingDomain, bookingPublisher, mockAuth)

	type inputParams struct {
		req *bookingv1.RescheduleBookingRequest
	}

	type outputParams struct {
		res *bookingv1.RescheduleBookingResponse
		err error
	}

	type testSetup struct {
		description string
		setup       func(ctx context.Context, domain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth)
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should reschedule a booking",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "update",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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
		{
			description: "reschedule booking call returns a gateway.ErrInvalidArgument",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "update",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().RescheduleBooking(ctx, params).Return(domain.RescheduleBookingResponse{}, gateway.ErrInvalidArgument)

			},
			output: outputParams{
				res: &bookingv1.RescheduleBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.Internal, "failed to reschedule booking, %s", gateway.ErrInvalidArgument.Error()),
			},
		},
		{
			description: "reschedule booking call returns a gateway.ErrInternalBadParameters",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "update",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().RescheduleBooking(ctx, params).Return(domain.RescheduleBookingResponse{}, gateway.ErrInternalBadParameters)

			},
			output: outputParams{
				res: &bookingv1.RescheduleBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.Internal, "failed to reschedule booking, %s", gateway.ErrInternalBadParameters.Error()),
			},
		},
		{
			description: "reschedule booking call returns a gateway.ErrInternal",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "update",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().RescheduleBooking(ctx, params).Return(domain.RescheduleBookingResponse{}, gateway.ErrInternal)

			},
			output: outputParams{
				res: &bookingv1.RescheduleBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.Internal, "failed to reschedule booking, %s", gateway.ErrInternal.Error()),
			},
		},
		{
			description: "reschedule booking call returns a gateway.ErrNotFound",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "update",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().RescheduleBooking(ctx, params).Return(domain.RescheduleBookingResponse{}, gateway.ErrNotFound)

			},
			output: outputParams{
				res: &bookingv1.RescheduleBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.NotFound, "failed to reschedule booking, %s", gateway.ErrNotFound.Error()),
			},
		},
		{
			description: "reschedule booking call returns a gateway.ErrOutOfRange",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "update",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().RescheduleBooking(ctx, params).Return(domain.RescheduleBookingResponse{}, gateway.ErrOutOfRange)

			},
			output: outputParams{
				res: &bookingv1.RescheduleBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.OutOfRange, "failed to reschedule booking, %s", gateway.ErrOutOfRange.Error()),
			},
		},
		{
			description: "reschedule booking call returns a gateway.ErrAlreadyExists",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "update",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().RescheduleBooking(ctx, params).Return(domain.RescheduleBookingResponse{}, gateway.ErrAlreadyExists)

			},
			output: outputParams{
				res: &bookingv1.RescheduleBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.AlreadyExists, "failed to reschedule booking, %s", gateway.ErrAlreadyExists.Error()),
			},
		},
		{
			description: "reschedule booking call returns a gateway.ErrInvalidAppointmentDate",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "update",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().RescheduleBooking(ctx, params).Return(domain.RescheduleBookingResponse{}, gateway.ErrInvalidAppointmentDate)

			},
			output: outputParams{
				res: &bookingv1.RescheduleBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.InvalidArgument, "failed to reschedule booking, %s", gateway.ErrInvalidAppointmentDate.Error()),
			},
		},
		{
			description: "reschedule booking call returns a gateway.ErrInvalidAppointmentTime",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "update",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().RescheduleBooking(ctx, params).Return(domain.RescheduleBookingResponse{}, gateway.ErrInvalidAppointmentTime)

			},
			output: outputParams{
				res: &bookingv1.RescheduleBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.InvalidArgument, "failed to reschedule booking, %s", gateway.ErrInvalidAppointmentTime.Error()),
			},
		},
		{
			description: "reschedule booking call returns an error",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "update",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(true, nil)

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

				bkDomain.EXPECT().RescheduleBooking(ctx, params).Return(domain.RescheduleBookingResponse{}, oops)

			},
			output: outputParams{
				res: &bookingv1.RescheduleBookingResponse{
					BookingId: "",
				},
				err: status.Errorf(codes.Internal, "failed to reschedule booking, %s", oops.Error()),
			},
		},
		{
			description: "should fail to do a reschedule because user is unauthorised",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "update",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-1",
				}).Return(false, nil)

			},
			output: outputParams{
				res: nil,
				err: status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", api.ErrUserUnauthorised),
			},
		},
		{
			description: "should fail to do a reschedule because customer is trying to access another customer therefore is unauthorised",
			input: inputParams{
				req: &bookingv1.RescheduleBookingRequest{
					AccountId: "account-id-2",
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
			setup: func(ctx context.Context, bkDomain *mocks.MockBookingDomain, publisher *mocks.MockBookingPublisher, mAuth *mocks.MockAuth) {

				mAuth.EXPECT().Authorize(ctx, &auth.PolicyParams{
					Action:     "update",
					Resource:   "uw.energy-smart.v1.booking-api",
					ResourceID: "account-id-2",
				}).Return(false, nil)

			},
			output: outputParams{
				res: nil,
				err: status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", api.ErrUserUnauthorised),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(ctx, bookingDomain, bookingPublisher, mockAuth)

			expected, err := myAPIHandler.RescheduleBooking(ctx, tc.input.req)
			if tc.output.err != nil {
				if diff := cmp.Diff(err.Error(), tc.output.err.Error()); diff != "" {
					t.Fatal(diff)
				}
			}

			if diff := cmp.Diff(expected, tc.output.res, cmpopts.IgnoreUnexported(date.Date{}, bookingv1.RescheduleBookingResponse{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
