//go:generate mockgen -source=grpc.go -destination ./mocks/grpc_mocks.go

package api_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	addressv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/address/v1"
	contract "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"

	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/api"
	mocks "github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/api/mocks"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/lowribeck"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/mapper"
	"github.com/utilitywarehouse/energy-smart-booking/internal/auth"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

var errOops = errors.New("oops")

func Test_GetAvailableSlots(t *testing.T) {
	now := time.Now().UTC().Format("02/01/2006 15:04:05")

	testCases := []struct {
		desc          string
		req           *lowribeck.GetCalendarAvailabilityRequest
		clientResp    *lowribeck.GetCalendarAvailabilityResponse
		mapperErr     error
		expected      *contract.GetAvailableSlotsResponse
		expectedError error
		setup         func(context.Context, *mocks.MockAuth)
	}{
		{
			desc: "Valid",
			req: &lowribeck.GetCalendarAvailabilityRequest{
				PostCode:    "postcode",
				ReferenceID: "reference",
				CreatedDate: now,
			},
			clientResp: &lowribeck.GetCalendarAvailabilityResponse{
				CalendarAvailabilityResult: []lowribeck.AvailabilitySlot{
					{
						AppointmentDate: "01/12/2023",
						AppointmentTime: "10:00-12:00",
					},
				},
			},
			expected: &contract.GetAvailableSlotsResponse{
				Slots: []*contract.BookingSlot{
					{
						Date: &date.Date{
							Year:  2023,
							Month: 10,
							Day:   1,
						},
						StartTime: 10,
						EndTime:   12,
					},
				},
			},
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Invalid postcode",
			mapperErr:     mapper.NewInvalidRequestError(mapper.InvalidPostcode),
			expectedError: status.Error(codes.InvalidArgument, "error making get available slots request: invalid request [postcode]"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Invalid reference",
			mapperErr:     mapper.NewInvalidRequestError(mapper.InvalidReference),
			expectedError: status.Error(codes.InvalidArgument, "error making get available slots request: invalid request [reference]"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Appointment not found",
			mapperErr:     mapper.ErrAppointmentNotFound,
			expectedError: status.Error(codes.NotFound, "error making get available slots request: no appointments found"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Appointment out of range",
			mapperErr:     mapper.ErrAppointmentOutOfRange,
			expectedError: status.Error(codes.OutOfRange, "error making get available slots request: appointment out of range"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Unknown bad parameter",
			mapperErr:     mapper.NewInvalidRequestError("something else"),
			expectedError: status.Error(codes.InvalidArgument, "error making get available slots request: invalid request [something else]"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Internal error",
			mapperErr:     fmt.Errorf("%w [%s]", mapper.ErrInternalError, "Insufficient notice to rearrange this appointment."),
			expectedError: status.Error(codes.Internal, "error making get available slots request: internal server error [Insufficient notice to rearrange this appointment.]"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Unknown error",
			mapperErr:     mapper.ErrUnknownError,
			expectedError: status.Error(codes.Internal, "error making get available slots request: unknown error"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
	}

	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	mAuth := mocks.NewMockAuth(ctrl)
	mapper := &fakeMapper{}

	myAPIHandler := api.New(client, mapper, mAuth)

	for _, tc := range testCases {
		t.Run(tc.desc, func(_ *testing.T) {
			mapper.availabilityRequest = tc.req
			mapper.availabilityResponse = tc.expected
			mapper.availabilityError = tc.mapperErr

			tc.setup(ctx, mAuth)

			client.EXPECT().GetCalendarAvailability(ctx, tc.req).Return(tc.clientResp, nil)

			result, err := myAPIHandler.GetAvailableSlots(ctx, &contract.GetAvailableSlotsRequest{
				Postcode:  "postcode",
				Reference: "reference",
			})

			if tc.expectedError == nil {
				assert.NoError(err, tc.desc)
				diff := cmp.Diff(tc.expected, result, protocmp.Transform(), cmpopts.IgnoreUnexported())
				assert.Empty(diff, tc.desc)
			} else {
				assert.EqualError(err, tc.expectedError.Error(), tc.desc)
			}
		})
	}
}

func Test_GetAvailableSlots_ClientError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	mAuth := mocks.NewMockAuth(ctrl)
	mapper := &fakeMapper{}

	myAPIHandler := api.New(client, mapper, mAuth)

	errorMessage := "received status code [500] (expected 200): Internal error has occurred, could not complete appointmentManagement GetCalendarAvailability request. The error has been logged."

	mAuth.EXPECT().Authorize(ctx,
		&auth.PolicyParams{
			Action:     "get",
			Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
			ResourceID: "lowribeck-api",
		}).Return(true, nil)

	req := &lowribeck.GetCalendarAvailabilityRequest{
		PostCode:    "postcode",
		ReferenceID: "reference",
		CreatedDate: time.Now().UTC().Format("02/01/2006 15:04:05"),
	}
	mapper.availabilityRequest = req

	client.EXPECT().GetCalendarAvailability(ctx, req).Return(nil, fmt.Errorf(errorMessage))

	_, err := myAPIHandler.GetAvailableSlots(ctx, &contract.GetAvailableSlotsRequest{
		Postcode:  "postcode",
		Reference: "reference",
	})

	assert.EqualError(err, "rpc error: code = Internal desc = error making get available slots request: "+errorMessage)
}

func Test_GetAvailableSlots_Unauthorised(t *testing.T) {

	testCases := []struct {
		desc          string
		expectedError error
		setup         func(context.Context, *mocks.MockAuth)
	}{
		{
			desc:          "Unauthorised",
			expectedError: status.Errorf(codes.PermissionDenied, "user does not have access to this action, %s", api.ErrUserUnauthorised),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(false, nil)
			},
		},
		{
			desc:          "Internal error",
			expectedError: status.Errorf(codes.Internal, "failed to validate credentials"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(false, errOops)
			},
		},
	}

	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	mAuth := mocks.NewMockAuth(ctrl)
	mapper := &fakeMapper{}

	myAPIHandler := api.New(client, mapper, mAuth)

	for _, tc := range testCases {
		t.Run(tc.desc, func(_ *testing.T) {

			tc.setup(ctx, mAuth)

			_, err := myAPIHandler.GetAvailableSlots(ctx, &contract.GetAvailableSlotsRequest{
				Postcode:  "postcode",
				Reference: "reference",
			})

			assert.EqualError(err, tc.expectedError.Error(), tc.desc)
		})
	}
}

func Test_CreateBooking(t *testing.T) {
	now := time.Now().UTC().Format("02/01/2006 15:04:05")

	testCases := []struct {
		desc          string
		req           *lowribeck.CreateBookingRequest
		clientResp    *lowribeck.CreateBookingResponse
		mapperErr     error
		expected      *contract.CreateBookingResponse
		expectedError error
		setup         func(context.Context, *mocks.MockAuth)
	}{
		{
			desc: "Valid",
			req: &lowribeck.CreateBookingRequest{
				PostCode:    "postcode",
				ReferenceID: "reference",
				CreatedDate: now,
			},
			clientResp: &lowribeck.CreateBookingResponse{
				ResponseCode: "B01",
			},
			expected: &contract.CreateBookingResponse{
				Success: true,
			},
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Invalid postcode",
			mapperErr:     mapper.NewInvalidRequestError(mapper.InvalidPostcode),
			expectedError: status.Error(codes.InvalidArgument, "error making booking request: invalid request [postcode]"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Invalid reference",
			mapperErr:     mapper.NewInvalidRequestError(mapper.InvalidReference),
			expectedError: status.Error(codes.InvalidArgument, "error making booking request: invalid request [reference]"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Invalid site",
			mapperErr:     mapper.NewInvalidRequestError(mapper.InvalidSite),
			expectedError: status.Error(codes.InvalidArgument, "error making booking request: invalid request [site]"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Appointment not found",
			mapperErr:     mapper.ErrAppointmentNotFound,
			expectedError: status.Error(codes.NotFound, "error making booking request: no appointments found"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Appointment out of range",
			mapperErr:     mapper.ErrAppointmentOutOfRange,
			expectedError: status.Error(codes.OutOfRange, "error making booking request: appointment out of range"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Appointment already exists",
			mapperErr:     mapper.ErrAppointmentAlreadyExists,
			expectedError: status.Error(codes.AlreadyExists, "error making booking request: appointment already exists"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Unknown bad parameter",
			mapperErr:     mapper.NewInvalidRequestError("something else"),
			expectedError: status.Error(codes.InvalidArgument, "error making booking request: invalid request [something else]"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Internal error",
			mapperErr:     fmt.Errorf("%w [%s]", mapper.ErrInternalError, "Insufficient notice to rearrange this appointment."),
			expectedError: status.Error(codes.Internal, "error making booking request: internal server error [Insufficient notice to rearrange this appointment.]"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Unknown error",
			mapperErr:     mapper.ErrUnknownError,
			expectedError: status.Error(codes.Internal, "error making booking request: unknown error"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
	}

	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	mAuth := mocks.NewMockAuth(ctrl)
	mapper := &fakeMapper{}

	myAPIHandler := api.New(client, mapper, mAuth)

	for _, tc := range testCases {
		t.Run(tc.desc, func(_ *testing.T) {
			mapper.bookingRequest = tc.req
			mapper.bookingResponse = tc.expected
			mapper.bookingError = tc.mapperErr

			tc.setup(ctx, mAuth)

			client.EXPECT().CreateBooking(ctx, tc.req).Return(tc.clientResp, nil)

			result, err := myAPIHandler.CreateBooking(ctx, &contract.CreateBookingRequest{
				Postcode:  "postcode",
				Reference: "reference",
			})

			if tc.expectedError == nil {
				assert.NoError(err, tc.desc)
				diff := cmp.Diff(tc.expected, result, protocmp.Transform(), cmpopts.IgnoreUnexported())
				assert.Empty(diff, tc.desc)
			} else {
				assert.EqualError(err, tc.expectedError.Error(), tc.desc)
			}
		})
	}
}

func Test_CreateBooking_ClientError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	mAuth := mocks.NewMockAuth(ctrl)
	mapper := &fakeMapper{}

	myAPIHandler := api.New(client, mapper, mAuth)

	errorMessage := "received status code [500] (expected 200): Internal error has occurred, could not complete appointmentManagement CreateBooking request. The error has been logged."

	mAuth.EXPECT().Authorize(ctx,
		&auth.PolicyParams{
			Action:     "create",
			Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
			ResourceID: "lowribeck-api",
		}).Return(true, nil)

	req := &lowribeck.CreateBookingRequest{
		PostCode:    "postcode",
		ReferenceID: "reference",
		CreatedDate: time.Now().UTC().Format("02/01/2006 15:04:05"),
	}
	mapper.bookingRequest = req

	client.EXPECT().CreateBooking(ctx, req).Return(nil, fmt.Errorf(errorMessage))

	_, err := myAPIHandler.CreateBooking(ctx, &contract.CreateBookingRequest{
		Postcode:  "postcode",
		Reference: "reference",
	})

	assert.EqualError(err, "rpc error: code = Internal desc = error making booking request: "+errorMessage)
}

func Test_CreateBooking_Unauthorised(t *testing.T) {

	testCases := []struct {
		desc          string
		expectedError error
		setup         func(context.Context, *mocks.MockAuth)
	}{
		{
			desc:          "Unauthorised",
			expectedError: status.Errorf(codes.PermissionDenied, "user does not have access to this action, %s", api.ErrUserUnauthorised),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(false, nil)
			},
		},
		{
			desc:          "Internal error",
			expectedError: status.Error(codes.Internal, "failed to validate credentials"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(false, errOops)
			},
		},
	}

	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	mAuth := mocks.NewMockAuth(ctrl)
	mapper := &fakeMapper{}

	myAPIHandler := api.New(client, mapper, mAuth)

	for _, tc := range testCases {
		t.Run(tc.desc, func(_ *testing.T) {
			tc.setup(ctx, mAuth)

			_, err := myAPIHandler.CreateBooking(ctx, &contract.CreateBookingRequest{
				Postcode:  "postcode",
				Reference: "reference",
			})

			assert.EqualError(err, tc.expectedError.Error(), tc.desc)
		})
	}
}

func Test_GetAvailableSlots_PointOfSale(t *testing.T) {
	now := time.Now().UTC().Format("02/01/2006 15:04:05")

	testCases := []struct {
		desc          string
		req           *lowribeck.GetCalendarAvailabilityRequest
		clientResp    *lowribeck.GetCalendarAvailabilityResponse
		mapperErr     error
		expected      *contract.GetAvailableSlotsPointOfSaleResponse
		expectedError error
		setup         func(context.Context, *mocks.MockAuth, *mocks.MockClient)
	}{
		{
			desc: "Valid",
			req: &lowribeck.GetCalendarAvailabilityRequest{
				PostCode:        "postcode",
				Mpan:            "mpan-1",
				Mprn:            "mprn-1",
				ElecJobTypeCode: "elec-job-type-code",
				GasJobTypeCode:  "gas-job-type-code",
				CreatedDate:     now,
			},
			clientResp: &lowribeck.GetCalendarAvailabilityResponse{
				CalendarAvailabilityResult: []lowribeck.AvailabilitySlot{
					{
						AppointmentDate: "01/12/2023",
						AppointmentTime: "10:00-12:00",
					},
				},
			},
			expected: &contract.GetAvailableSlotsPointOfSaleResponse{
				Slots: []*contract.BookingSlot{
					{
						Date: &date.Date{
							Year:  2023,
							Month: 10,
							Day:   1,
						},
						StartTime: 10,
						EndTime:   12,
					},
				},
			},
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, client *mocks.MockClient) {
				client.EXPECT().GetCalendarAvailabilityPointOfSale(ctx, &lowribeck.GetCalendarAvailabilityRequest{
					PostCode:        "postcode",
					Mpan:            "mpan-1",
					Mprn:            "mprn-1",
					ElecJobTypeCode: "elec-job-type-code",
					GasJobTypeCode:  "gas-job-type-code",
					CreatedDate:     now,
				}).Return(&lowribeck.GetCalendarAvailabilityResponse{
					CalendarAvailabilityResult: []lowribeck.AvailabilitySlot{
						{
							AppointmentDate: "01/12/2023",
							AppointmentTime: "10:00-12:00",
						},
					},
				}, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Invalid postcode",
			mapperErr:     mapper.NewInvalidRequestError(mapper.InvalidPostcode),
			expectedError: status.Error(codes.InvalidArgument, "error making get available slots point of sale request: invalid request [postcode]"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, client *mocks.MockClient) {
				client.EXPECT().GetCalendarAvailabilityPointOfSale(ctx, gomock.Any()).Return(nil, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Appointment not found",
			mapperErr:     mapper.ErrAppointmentNotFound,
			expectedError: status.Error(codes.NotFound, "error making get available slots point of sale request: no appointments found"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, client *mocks.MockClient) {
				client.EXPECT().GetCalendarAvailabilityPointOfSale(ctx, gomock.Any()).Return(nil, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Appointment out of range",
			mapperErr:     mapper.ErrAppointmentOutOfRange,
			expectedError: status.Error(codes.OutOfRange, "error making get available slots point of sale request: appointment out of range"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, client *mocks.MockClient) {
				client.EXPECT().GetCalendarAvailabilityPointOfSale(ctx, gomock.Any()).Return(nil, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Unknown bad parameter",
			mapperErr:     mapper.NewInvalidRequestError("something else"),
			expectedError: status.Error(codes.InvalidArgument, "error making get available slots point of sale request: invalid request [something else]"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, client *mocks.MockClient) {
				client.EXPECT().GetCalendarAvailabilityPointOfSale(ctx, gomock.Any()).Return(nil, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Internal error",
			mapperErr:     fmt.Errorf("%w [%s]", mapper.ErrInternalError, "Insufficient notice to rearrange this appointment."),
			expectedError: status.Error(codes.Internal, "error making get available slots point of sale request: internal server error [Insufficient notice to rearrange this appointment.]"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, client *mocks.MockClient) {
				client.EXPECT().GetCalendarAvailabilityPointOfSale(ctx, gomock.Any()).Return(nil, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Unknown error",
			mapperErr:     mapper.ErrUnknownError,
			expectedError: status.Error(codes.Internal, "error making get available slots point of sale request: unknown error"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, client *mocks.MockClient) {
				client.EXPECT().GetCalendarAvailabilityPointOfSale(ctx, gomock.Any()).Return(nil, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
	}

	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	defer ctrl.Finish()

	mClient := mocks.NewMockClient(ctrl)
	mAuth := mocks.NewMockAuth(ctrl)
	mapper := &fakeMapper{}

	myAPIHandler := api.New(mClient, mapper, mAuth)

	for _, tc := range testCases {
		t.Run(tc.desc, func(_ *testing.T) {
			mapper.availabilityRequest = tc.req
			mapper.availabilityPointOfSaleResponse = tc.expected
			mapper.availabilityError = tc.mapperErr

			tc.setup(ctx, mAuth, mClient)

			result, err := myAPIHandler.GetAvailableSlotsPointOfSale(ctx, &contract.GetAvailableSlotsPointOfSaleRequest{
				Postcode:              "postcode",
				Mpan:                  "mpan-1",
				Mprn:                  "mprn-1",
				ElectricityTariffType: contract.TariffType_TARIFF_TYPE_CREDIT,
				GasTariffType:         contract.TariffType_TARIFF_TYPE_CREDIT,
			})

			if tc.expectedError == nil {
				assert.NoError(err, tc.desc)
				diff := cmp.Diff(tc.expected, result, protocmp.Transform(), cmpopts.IgnoreUnexported())
				assert.Empty(diff, tc.desc)
			} else {
				assert.EqualError(err, tc.expectedError.Error(), tc.desc)
			}
		})
	}
}

func Test_GetAvailableSlots_PointOfSale_ClientError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	mAuth := mocks.NewMockAuth(ctrl)
	mapper := &fakeMapper{}

	myAPIHandler := api.New(client, mapper, mAuth)

	errorMessage := "received status code [500] (expected 200): Internal error has occurred, could not complete appointmentManagement GetCalendarAvailability request. The error has been logged."

	mAuth.EXPECT().Authorize(ctx,
		&auth.PolicyParams{
			Action:     "get",
			Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
			ResourceID: "lowribeck-api",
		}).Return(true, nil)

	req := &lowribeck.GetCalendarAvailabilityRequest{
		PostCode:        "postcode",
		Mpan:            "mpan-1",
		Mprn:            "mprn-1",
		ElecJobTypeCode: "credit",
		GasJobTypeCode:  "credit",
		CreatedDate:     time.Now().UTC().Format("02/01/2006 15:04:05"),
	}
	mapper.availabilityRequest = req

	client.EXPECT().GetCalendarAvailabilityPointOfSale(ctx, req).Return(nil, fmt.Errorf(errorMessage))

	_, err := myAPIHandler.GetAvailableSlotsPointOfSale(ctx, &contract.GetAvailableSlotsPointOfSaleRequest{
		Postcode:              "postcode",
		Mpan:                  "mpan-1",
		Mprn:                  "mprn-1",
		ElectricityTariffType: contract.TariffType_TARIFF_TYPE_CREDIT,
		GasTariffType:         contract.TariffType_TARIFF_TYPE_CREDIT,
	})

	assert.EqualError(err, "rpc error: code = Internal desc = error making get available slots point of sale request: "+errorMessage)
}

func Test_GetAvailableSlots_PointOfSale_Unauthorised(t *testing.T) {

	testCases := []struct {
		desc          string
		expectedError error
		setup         func(context.Context, *mocks.MockAuth)
	}{
		{
			desc:          "Unauthorised",
			expectedError: status.Errorf(codes.PermissionDenied, "user does not have access to this action, %s", api.ErrUserUnauthorised),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(false, nil)
			},
		},
		{
			desc:          "Internal error",
			expectedError: status.Errorf(codes.Internal, "failed to validate credentials"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "get",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(false, errOops)
			},
		},
	}

	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	mAuth := mocks.NewMockAuth(ctrl)
	mapper := &fakeMapper{}

	myAPIHandler := api.New(client, mapper, mAuth)

	for _, tc := range testCases {
		t.Run(tc.desc, func(_ *testing.T) {

			tc.setup(ctx, mAuth)

			_, err := myAPIHandler.GetAvailableSlotsPointOfSale(ctx, &contract.GetAvailableSlotsPointOfSaleRequest{
				Postcode:              "postcode",
				Mpan:                  "mpan-1",
				Mprn:                  "mprn-1",
				ElectricityTariffType: contract.TariffType_TARIFF_TYPE_CREDIT,
				GasTariffType:         contract.TariffType_TARIFF_TYPE_CREDIT,
			})

			assert.EqualError(err, tc.expectedError.Error(), tc.desc)
		})
	}
}

func Test_CreateBooking_PointOfSale(t *testing.T) {
	now := time.Now().UTC().Format("02/01/2006 15:04:05")

	req := &lowribeck.CreateBookingRequest{
		SubBuildName:            "sub-1",
		BuildingName:            "bn-1",
		DependThroughfare:       "dt-1",
		Throughfare:             "tf-1",
		DoubleDependantLocality: "ddl-1",
		DependantLocality:       "dl-1",
		PostTown:                "pt",
		County:                  "", // There is no County in the PAF format
		PostCode:                "postcode",
		Mpan:                    "mpan-1",
		Mprn:                    "mprn-1",
		ElecJobTypeCode:         "credit",
		GasJobTypeCode:          "credit",
		CreatedDate:             now,
	}

	testCases := []struct {
		desc          string
		req           *lowribeck.CreateBookingRequest
		mapperErr     error
		expected      *contract.CreateBookingPointOfSaleResponse
		expectedError error
		setup         func(context.Context, *mocks.MockAuth, *mocks.MockClient)
	}{
		{
			desc: "Valid",
			expected: &contract.CreateBookingPointOfSaleResponse{
				Success: true,
			},
			req: req,
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, mClient *mocks.MockClient) {
				mClient.EXPECT().CreateBookingPointOfSale(ctx, req).Return(&lowribeck.CreateBookingResponse{
					ResponseCode: "B01",
				}, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Invalid postcode",
			mapperErr:     mapper.NewInvalidRequestError(mapper.InvalidPostcode),
			expectedError: status.Error(codes.InvalidArgument, "error making booking point of sale request: invalid request [postcode]"),
			req:           req,
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, mClient *mocks.MockClient) {
				mClient.EXPECT().CreateBookingPointOfSale(ctx, req).Return(&lowribeck.CreateBookingResponse{
					ResponseCode: "",
				}, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Invalid mpan",
			mapperErr:     mapper.NewInvalidRequestError(mapper.InvalidMPAN),
			expectedError: status.Error(codes.InvalidArgument, "error making booking point of sale request: invalid request [mpan]"),
			req:           req,
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, mClient *mocks.MockClient) {
				mClient.EXPECT().CreateBookingPointOfSale(ctx, req).Return(&lowribeck.CreateBookingResponse{
					ResponseCode: "",
				}, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Invalid mprn",
			mapperErr:     mapper.NewInvalidRequestError(mapper.InvalidMPRN),
			expectedError: status.Error(codes.InvalidArgument, "error making booking point of sale request: invalid request [mprn]"),
			req:           req,
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, mClient *mocks.MockClient) {
				mClient.EXPECT().CreateBookingPointOfSale(ctx, req).Return(&lowribeck.CreateBookingResponse{
					ResponseCode: "",
				}, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Invalid electricity job type code",
			mapperErr:     mapper.ErrInvalidElectricityJobTypeCode,
			expectedError: status.Error(codes.Internal, "error making booking point of sale request: invalid electricity job type code"),
			req:           req,
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, mClient *mocks.MockClient) {
				mClient.EXPECT().CreateBookingPointOfSale(ctx, req).Return(&lowribeck.CreateBookingResponse{
					ResponseCode: "",
				}, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Invalid gas job type code",
			mapperErr:     mapper.ErrInvalidGasJobTypeCode,
			expectedError: status.Error(codes.Internal, "error making booking point of sale request: invalid gas job type code"),
			req:           req,
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, mClient *mocks.MockClient) {
				mClient.EXPECT().CreateBookingPointOfSale(ctx, req).Return(&lowribeck.CreateBookingResponse{
					ResponseCode: "",
				}, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Invalid (unspecified)job type code",
			mapperErr:     mapper.ErrInvalidJobTypeCode,
			expectedError: status.Error(codes.Internal, "error making booking point of sale request: invalid job type code"),
			req:           req,
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, mClient *mocks.MockClient) {
				mClient.EXPECT().CreateBookingPointOfSale(ctx, req).Return(&lowribeck.CreateBookingResponse{
					ResponseCode: "",
				}, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Invalid site",
			mapperErr:     mapper.NewInvalidRequestError(mapper.InvalidSite),
			expectedError: status.Error(codes.InvalidArgument, "error making booking point of sale request: invalid request [site]"),
			req:           req,
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, mClient *mocks.MockClient) {
				mClient.EXPECT().CreateBookingPointOfSale(ctx, req).Return(&lowribeck.CreateBookingResponse{
					ResponseCode: "",
				}, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Appointment not found",
			mapperErr:     mapper.ErrAppointmentNotFound,
			expectedError: status.Error(codes.NotFound, "error making booking point of sale request: no appointments found"),
			req:           req,
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, mClient *mocks.MockClient) {
				mClient.EXPECT().CreateBookingPointOfSale(ctx, req).Return(&lowribeck.CreateBookingResponse{
					ResponseCode: "",
				}, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Appointment out of range",
			mapperErr:     mapper.ErrAppointmentOutOfRange,
			expectedError: status.Error(codes.OutOfRange, "error making booking point of sale request: appointment out of range"),
			req:           req,
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, mClient *mocks.MockClient) {
				mClient.EXPECT().CreateBookingPointOfSale(ctx, req).Return(&lowribeck.CreateBookingResponse{
					ResponseCode: "",
				}, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Appointment already exists",
			mapperErr:     mapper.ErrAppointmentAlreadyExists,
			expectedError: status.Error(codes.AlreadyExists, "error making booking point of sale request: appointment already exists"),
			req:           req,
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, mClient *mocks.MockClient) {
				mClient.EXPECT().CreateBookingPointOfSale(ctx, req).Return(&lowribeck.CreateBookingResponse{
					ResponseCode: "",
				}, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Unknown bad parameter",
			mapperErr:     mapper.NewInvalidRequestError("something else"),
			expectedError: status.Error(codes.InvalidArgument, "error making booking point of sale request: invalid request [something else]"),
			req:           req,
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, mClient *mocks.MockClient) {
				mClient.EXPECT().CreateBookingPointOfSale(ctx, req).Return(&lowribeck.CreateBookingResponse{
					ResponseCode: "",
				}, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Internal error",
			mapperErr:     fmt.Errorf("%w [%s]", mapper.ErrInternalError, "Insufficient notice to rearrange this appointment."),
			expectedError: status.Error(codes.Internal, "error making booking point of sale request: internal server error [Insufficient notice to rearrange this appointment.]"),
			req:           req,
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, mClient *mocks.MockClient) {
				mClient.EXPECT().CreateBookingPointOfSale(ctx, req).Return(&lowribeck.CreateBookingResponse{
					ResponseCode: "",
				}, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Unknown error",
			mapperErr:     mapper.ErrUnknownError,
			expectedError: status.Error(codes.Internal, "error making booking point of sale request: unknown error"),
			req:           req,
			setup: func(ctx context.Context, mAuth *mocks.MockAuth, mClient *mocks.MockClient) {
				mClient.EXPECT().CreateBookingPointOfSale(ctx, req).Return(&lowribeck.CreateBookingResponse{
					ResponseCode: "",
				}, nil)
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
	}

	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	defer ctrl.Finish()

	mClient := mocks.NewMockClient(ctrl)
	mAuth := mocks.NewMockAuth(ctrl)
	mapper := &fakeMapper{}

	myAPIHandler := api.New(mClient, mapper, mAuth)

	for _, tc := range testCases {
		t.Run(tc.desc, func(_ *testing.T) {
			mapper.bookingRequest = tc.req
			mapper.bookingPointOfSaleResponse = tc.expected
			mapper.bookingError = tc.mapperErr

			tc.setup(ctx, mAuth, mClient)

			result, err := myAPIHandler.CreateBookingPointOfSale(ctx, &contract.CreateBookingPointOfSaleRequest{
				Postcode:              "postcode",
				Mpan:                  "mpan-1",
				Mprn:                  "mprn-1",
				ElectricityTariffType: contract.TariffType_TARIFF_TYPE_CREDIT,
				GasTariffType:         contract.TariffType_TARIFF_TYPE_CREDIT,
				SiteAddress: &addressv1.Address{
					Uprn: "uprn-1",
					Paf: &addressv1.Address_PAF{
						Organisation:            "org",
						Department:              "department-1",
						SubBuilding:             "sub-1",
						BuildingName:            "bn-1",
						BuildingNumber:          "bnum-1",
						DependentThoroughfare:   "dt-1",
						Thoroughfare:            "tf-1",
						DoubleDependentLocality: "ddl-1",
						DependentLocality:       "dl-1",
						PostTown:                "pt",
						Postcode:                "postcode",
					},
				},
			})

			if tc.expectedError == nil {
				assert.NoError(err, tc.desc)
				diff := cmp.Diff(tc.expected, result, protocmp.Transform(), cmpopts.IgnoreUnexported())
				assert.Empty(diff, tc.desc)
			} else {
				assert.EqualError(err, tc.expectedError.Error(), tc.desc)
			}
		})
	}
}

func Test_CreateBooking_PointOfSale_ClientError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	mAuth := mocks.NewMockAuth(ctrl)
	mapper := &fakeMapper{}

	myAPIHandler := api.New(client, mapper, mAuth)

	errorMessage := "received status code [500] (expected 200): Internal error has occurred, could not complete appointmentManagement CreateBooking request. The error has been logged."

	mAuth.EXPECT().Authorize(ctx,
		&auth.PolicyParams{
			Action:     "create",
			Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
			ResourceID: "lowribeck-api",
		}).Return(true, nil)

	req := &lowribeck.CreateBookingRequest{
		SubBuildName:            "sub-1",
		BuildingName:            "bn-1",
		DependThroughfare:       "dt-1",
		Throughfare:             "tf-1",
		DoubleDependantLocality: "ddl-1",
		DependantLocality:       "dl-1",
		PostTown:                "pt",
		County:                  "", // There is no County in the PAF format
		PostCode:                "postcode",
		Mpan:                    "mpan-1",
		Mprn:                    "mprn-1",
		ElecJobTypeCode:         "credit",
		GasJobTypeCode:          "credit",
		CreatedDate:             time.Now().UTC().Format("02/01/2006 15:04:05"),
	}
	mapper.bookingRequest = req

	client.EXPECT().CreateBookingPointOfSale(ctx, req).Return(nil, fmt.Errorf(errorMessage))

	_, err := myAPIHandler.CreateBookingPointOfSale(ctx, &contract.CreateBookingPointOfSaleRequest{
		Postcode:              "postcode",
		Mpan:                  "mpan-1",
		Mprn:                  "mprn-1",
		ElectricityTariffType: contract.TariffType_TARIFF_TYPE_CREDIT,
		GasTariffType:         contract.TariffType_TARIFF_TYPE_CREDIT,
		SiteAddress: &addressv1.Address{
			Uprn: "uprn-1",
			Paf: &addressv1.Address_PAF{
				Organisation:            "org",
				Department:              "department-1",
				SubBuilding:             "sub-1",
				BuildingName:            "bn-1",
				BuildingNumber:          "bnum-1",
				DependentThoroughfare:   "dt-1",
				Thoroughfare:            "tf-1",
				DoubleDependentLocality: "ddl-1",
				DependentLocality:       "dl-1",
				PostTown:                "pt",
				Postcode:                "postcode",
			},
		},
	})

	assert.EqualError(err, "rpc error: code = Internal desc = error making booking point of sale request: "+errorMessage)
}

func Test_CreateBooking_PointOfSale_Unauthorised(t *testing.T) {

	testCases := []struct {
		desc          string
		expectedError error
		setup         func(context.Context, *mocks.MockAuth)
	}{
		{
			desc:          "Unauthorised",
			expectedError: status.Errorf(codes.PermissionDenied, "user does not have access to this action, %s", api.ErrUserUnauthorised),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(false, nil)
			},
		},
		{
			desc:          "Internal error",
			expectedError: status.Error(codes.Internal, "failed to validate credentials"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "create",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(false, errOops)
			},
		},
	}

	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	mAuth := mocks.NewMockAuth(ctrl)
	mapper := &fakeMapper{}

	myAPIHandler := api.New(client, mapper, mAuth)

	for _, tc := range testCases {
		t.Run(tc.desc, func(_ *testing.T) {
			tc.setup(ctx, mAuth)

			_, err := myAPIHandler.CreateBookingPointOfSale(ctx, &contract.CreateBookingPointOfSaleRequest{
				Postcode:              "postcode",
				Mpan:                  "mpan-1",
				Mprn:                  "mprn-1",
				ElectricityTariffType: contract.TariffType_TARIFF_TYPE_CREDIT,
				GasTariffType:         contract.TariffType_TARIFF_TYPE_CREDIT,
				SiteAddress: &addressv1.Address{
					Uprn: "uprn-1",
					Paf: &addressv1.Address_PAF{
						Organisation:            "org",
						Department:              "department-1",
						SubBuilding:             "sub-1",
						BuildingName:            "bn-1",
						BuildingNumber:          "bnum-1",
						DependentThoroughfare:   "dt-1",
						Thoroughfare:            "tf-1",
						DoubleDependentLocality: "ddl-1",
						DependentLocality:       "dl-1",
						PostTown:                "pt",
						Postcode:                "postcode",
					},
				},
			})

			assert.EqualError(err, tc.expectedError.Error(), tc.desc)
		})
	}
}

func Test_UpdateContactDetails(t *testing.T) {
	now := time.Now().UTC().Format("02/01/2006 15:04:05")

	testCases := []struct {
		desc          string
		req           *lowribeck.UpdateContactDetailsRequest
		clientResp    *lowribeck.UpdateContactDetailsResponse
		mapperErr     error
		expected      *contract.UpdateContactDetailsResponse
		expectedError error
		setup         func(context.Context, *mocks.MockAuth)
	}{
		{
			desc: "Valid",
			req: &lowribeck.UpdateContactDetailsRequest{
				SiteContactName: "Test User",
				ReferenceID:     "reference",
				CreatedDate:     now,
			},
			clientResp: &lowribeck.UpdateContactDetailsResponse{
				ResponseCode: "U01",
			},
			expected: &contract.UpdateContactDetailsResponse{
				Success: true,
			},
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "update",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Invalid reference",
			mapperErr:     mapper.NewInvalidRequestError(mapper.InvalidReference),
			expectedError: status.Error(codes.InvalidArgument, "error making update contact detail request: invalid request [reference]"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "update",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Appointment not found",
			mapperErr:     mapper.ErrAppointmentNotFound,
			expectedError: status.Error(codes.NotFound, "error making update contact detail request: no appointments found"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "update",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Unknown bad parameter",
			mapperErr:     mapper.NewInvalidRequestError("something else"),
			expectedError: status.Error(codes.InvalidArgument, "error making update contact detail request: invalid request [something else]"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "update",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Internal error",
			mapperErr:     fmt.Errorf("%w [%s]", mapper.ErrInternalError, "Insufficient notice to rearrange this appointment."),
			expectedError: status.Error(codes.Internal, "error making update contact detail request: internal server error [Insufficient notice to rearrange this appointment.]"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "update",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
		{
			desc:          "Unknown error",
			mapperErr:     mapper.ErrUnknownError,
			expectedError: status.Error(codes.Internal, "error making update contact detail request: unknown error"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "update",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(true, nil)
			},
		},
	}

	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	mAuth := mocks.NewMockAuth(ctrl)
	mapper := &fakeMapper{}

	myAPIHandler := api.New(client, mapper, mAuth)

	for _, tc := range testCases {
		t.Run(tc.desc, func(_ *testing.T) {
			mapper.updateContactRequest = tc.req
			mapper.updateContactResponse = tc.expected
			mapper.updateContactError = tc.mapperErr

			tc.setup(ctx, mAuth)

			client.EXPECT().UpdateContactDetails(ctx, tc.req).Return(tc.clientResp, nil)

			result, err := myAPIHandler.UpdateContactDetails(ctx, &contract.UpdateContactDetailsRequest{
				ContactDetails: &contract.ContactDetails{
					FirstName: "Test",
					LastName:  "User",
				},
				Reference: "reference",
			})

			if tc.expectedError == nil {
				assert.NoError(err, tc.desc)
				diff := cmp.Diff(tc.expected, result, protocmp.Transform(), cmpopts.IgnoreUnexported())
				assert.Empty(diff, tc.desc)
			} else {
				assert.EqualError(err, tc.expectedError.Error(), tc.desc)
			}
		})
	}
}

func Test_UpdateContactDetails_ClientError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	mAuth := mocks.NewMockAuth(ctrl)
	mapper := &fakeMapper{}

	myAPIHandler := api.New(client, mapper, mAuth)

	errorMessage := "received status code [500] (expected 200): Internal error has occurred, could not complete appointmentManagement UpdateContactDetails request. The error has been logged."

	mAuth.EXPECT().Authorize(ctx,
		&auth.PolicyParams{
			Action:     "update",
			Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
			ResourceID: "lowribeck-api",
		}).Return(true, nil)

	req := &lowribeck.UpdateContactDetailsRequest{
		SiteContactName: "Test User",
		ReferenceID:     "reference",
		CreatedDate:     time.Now().UTC().Format("02/01/2006 15:04:05"),
	}
	mapper.updateContactRequest = req

	client.EXPECT().UpdateContactDetails(ctx, req).Return(nil, fmt.Errorf(errorMessage))

	_, err := myAPIHandler.UpdateContactDetails(ctx, &contract.UpdateContactDetailsRequest{
		ContactDetails: &contract.ContactDetails{
			FirstName: "Test",
			LastName:  "User",
		},
		Reference: "reference",
	})

	assert.EqualError(err, "rpc error: code = Internal desc = error making update contact detail request: "+errorMessage)
}

func Test_UpdateContactDetails_Unauthorised(t *testing.T) {

	testCases := []struct {
		desc          string
		expectedError error
		setup         func(context.Context, *mocks.MockAuth)
	}{
		{
			desc:          "Unauthorised",
			expectedError: status.Errorf(codes.PermissionDenied, "user does not have access to this action, %s", api.ErrUserUnauthorised),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "update",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(false, nil)
			},
		},
		{
			desc:          "Internal error",
			expectedError: status.Error(codes.Internal, "failed to validate credentials"),
			setup: func(ctx context.Context, mAuth *mocks.MockAuth) {
				mAuth.EXPECT().Authorize(ctx,
					&auth.PolicyParams{
						Action:     "update",
						Resource:   "uw.energy-smart.v1.lowribeck-wrapper",
						ResourceID: "lowribeck-api",
					}).Return(false, errOops)
			},
		},
	}

	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	mAuth := mocks.NewMockAuth(ctrl)
	mapper := &fakeMapper{}

	myAPIHandler := api.New(client, mapper, mAuth)

	for _, tc := range testCases {
		t.Run(tc.desc, func(_ *testing.T) {
			tc.setup(ctx, mAuth)

			_, err := myAPIHandler.UpdateContactDetails(ctx, &contract.UpdateContactDetailsRequest{
				ContactDetails: &contract.ContactDetails{
					FirstName: "Test",
					LastName:  "User",
				},
				Reference: "reference",
			})

			assert.EqualError(err, tc.expectedError.Error(), tc.desc)
		})
	}
}

type fakeMapper struct {
	availabilityRequest  *lowribeck.GetCalendarAvailabilityRequest
	availabilityResponse *contract.GetAvailableSlotsResponse
	availabilityError    error
	bookingRequest       *lowribeck.CreateBookingRequest
	bookingResponse      *contract.CreateBookingResponse
	bookingError         error

	updateContactRequest  *lowribeck.UpdateContactDetailsRequest
	updateContactResponse *contract.UpdateContactDetailsResponse
	updateContactError    error

	availabilityPointOfSaleResponse *contract.GetAvailableSlotsPointOfSaleResponse
	bookingPointOfSaleResponse      *contract.CreateBookingPointOfSaleResponse
}

func (f *fakeMapper) AvailabilityRequest(_ uint32, _ *contract.GetAvailableSlotsRequest) *lowribeck.GetCalendarAvailabilityRequest {
	return f.availabilityRequest
}

func (f *fakeMapper) AvailableSlotsResponse(_ *lowribeck.GetCalendarAvailabilityResponse) (*contract.GetAvailableSlotsResponse, error) {
	if f.availabilityError != nil {
		return nil, f.availabilityError
	}
	return f.availabilityResponse, nil
}
func (f *fakeMapper) BookingRequest(_ uint32, _ *contract.CreateBookingRequest) (*lowribeck.CreateBookingRequest, error) {
	return f.bookingRequest, nil
}
func (f *fakeMapper) BookingResponse(_ *lowribeck.CreateBookingResponse) (*contract.CreateBookingResponse, error) {
	if f.bookingError != nil {
		return nil, f.bookingError
	}
	return f.bookingResponse, nil
}

func (f *fakeMapper) AvailabilityRequestPointOfSale(_ uint32, _ *contract.GetAvailableSlotsPointOfSaleRequest) (*lowribeck.GetCalendarAvailabilityRequest, error) {
	return f.availabilityRequest, nil
}

func (f *fakeMapper) BookingRequestPointOfSale(_ uint32, _ *contract.CreateBookingPointOfSaleRequest) (*lowribeck.CreateBookingRequest, error) {
	return f.bookingRequest, nil
}

func (f *fakeMapper) AvailableSlotsPointOfSaleResponse(_ *lowribeck.GetCalendarAvailabilityResponse) (*contract.GetAvailableSlotsPointOfSaleResponse, error) {
	if f.availabilityError != nil {
		return nil, f.availabilityError
	}
	return f.availabilityPointOfSaleResponse, nil
}

func (f *fakeMapper) BookingResponsePointOfSale(_ *lowribeck.CreateBookingResponse) (*contract.CreateBookingPointOfSaleResponse, error) {
	if f.bookingError != nil {
		return nil, f.bookingError
	}
	return f.bookingPointOfSaleResponse, nil
}

func (f *fakeMapper) UpdateContactDetailsRequest(_ uint32, _ *contract.UpdateContactDetailsRequest) *lowribeck.UpdateContactDetailsRequest {
	return f.updateContactRequest
}
func (f *fakeMapper) UpdateContactDetailsResponse(_ *lowribeck.UpdateContactDetailsResponse) (*contract.UpdateContactDetailsResponse, error) {
	if f.updateContactError != nil {
		return nil, f.updateContactError
	}
	return f.updateContactResponse, nil
}
