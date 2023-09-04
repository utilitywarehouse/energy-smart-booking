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

var oops = errors.New("oops")

func Test_GetAvailableSlots(t *testing.T) {
	now := time.Now().UTC().Format("02/01/2006 15:04:05")

	testCases := []struct {
		desc          string
		req           *lowribeck.GetCalendarAvailabilityRequest
		clientResp    *lowribeck.GetCalendarAvailabilityResponse
		clientErr     error
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
			clientErr:     mapper.NewInvalidRequestError(mapper.InvalidPostcode),
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
			clientErr:     mapper.NewInvalidRequestError(mapper.InvalidReference),
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
			clientErr:     mapper.ErrAppointmentNotFound,
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
			clientErr:     mapper.ErrAppointmentOutOfRange,
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
			clientErr:     mapper.ErrInvalidRequest,
			expectedError: status.Error(codes.InvalidArgument, "error making get available slots request: invalid request"),
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
			clientErr:     fmt.Errorf("unknown"),
			expectedError: status.Error(codes.Internal, "error making get available slots request: unknown"),
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
		t.Run(tc.desc, func(t *testing.T) {
			mapper.availabilityRequest = tc.req
			mapper.availabilityResponse = tc.expected

			tc.setup(ctx, mAuth)

			client.EXPECT().GetCalendarAvailability(ctx, tc.req).Return(tc.clientResp, tc.clientErr)

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

func Test_GetAvailableSlots_Unauthorised(t *testing.T) {

	testCases := []struct {
		desc          string
		expectedError error
		setup         func(context.Context, *mocks.MockAuth)
	}{
		{
			desc:          "Unauthorised",
			expectedError: status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", api.ErrUserUnauthorised),
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
					}).Return(false, oops)
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
		t.Run(tc.desc, func(t *testing.T) {

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
		clientErr     error
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
			clientErr:     mapper.NewInvalidRequestError(mapper.InvalidPostcode),
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
			clientErr:     mapper.NewInvalidRequestError(mapper.InvalidReference),
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
			clientErr:     mapper.NewInvalidRequestError(mapper.InvalidSite),
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
			clientErr:     mapper.ErrAppointmentNotFound,
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
			clientErr:     mapper.ErrAppointmentOutOfRange,
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
			clientErr:     mapper.ErrAppointmentAlreadyExists,
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
			clientErr:     mapper.ErrInvalidRequest,
			expectedError: status.Error(codes.InvalidArgument, "error making booking request: invalid request"),
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
			clientErr:     fmt.Errorf("unknown"),
			expectedError: status.Error(codes.Internal, "error making booking request: unknown"),
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
		t.Run(tc.desc, func(t *testing.T) {
			mapper.bookingRequest = tc.req
			mapper.booking = tc.expected

			tc.setup(ctx, mAuth)

			client.EXPECT().CreateBooking(ctx, tc.req).Return(tc.clientResp, tc.clientErr)

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

func Test_CreateBooking_Unauthorised(t *testing.T) {

	testCases := []struct {
		desc          string
		expectedError error
		setup         func(context.Context, *mocks.MockAuth)
	}{
		{
			desc:          "Unauthorised",
			expectedError: status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", api.ErrUserUnauthorised),
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
					}).Return(false, oops)
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
		t.Run(tc.desc, func(t *testing.T) {
			tc.setup(ctx, mAuth)

			_, err := myAPIHandler.CreateBooking(ctx, &contract.CreateBookingRequest{
				Postcode:  "postcode",
				Reference: "reference",
			})

			assert.EqualError(err, tc.expectedError.Error(), tc.desc)
		})
	}
}

type fakeMapper struct {
	availabilityRequest  *lowribeck.GetCalendarAvailabilityRequest
	availabilityResponse *contract.GetAvailableSlotsResponse
	bookingRequest       *lowribeck.CreateBookingRequest
	booking              *contract.CreateBookingResponse
}

func (f *fakeMapper) AvailabilityRequest(_ uint32, req *contract.GetAvailableSlotsRequest) *lowribeck.GetCalendarAvailabilityRequest {
	return f.availabilityRequest
}

func (f *fakeMapper) AvailableSlotsResponse(resp *lowribeck.GetCalendarAvailabilityResponse) (*contract.GetAvailableSlotsResponse, error) {
	return f.availabilityResponse, nil
}
func (f *fakeMapper) BookingRequest(_ uint32, resp *contract.CreateBookingRequest) (*lowribeck.CreateBookingRequest, error) {
	return f.bookingRequest, nil
}
func (f *fakeMapper) BookingResponse(resp *lowribeck.CreateBookingResponse) (*contract.CreateBookingResponse, error) {
	return f.booking, nil
}
