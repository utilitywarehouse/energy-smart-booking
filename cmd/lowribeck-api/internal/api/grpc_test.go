//go:generate mockgen -source=grpc.go -destination ./mocks/grpc_mocks.go

package api_test

import (
	"context"
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
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

func Test_GetAvailableSlots(t *testing.T) {
	now := time.Now().UTC().Format("02/01/2006 15:04:05")

	testCases := []struct {
		desc          string
		req           *lowribeck.GetCalendarAvailabilityRequest
		clientResp    *lowribeck.GetCalendarAvailabilityResponse
		clientErr     error
		expected      *contract.GetAvailableSlotsResponse
		expectedError error
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
		},
		{
			desc:          "Invalid postcode",
			clientErr:     mapper.NewInvalidRequestError(mapper.InvalidPostcode),
			expectedError: status.Error(codes.InvalidArgument, "error making get available slots request: invalid request [postcode]"),
		},
		{
			desc:          "Invalid reference",
			clientErr:     mapper.NewInvalidRequestError(mapper.InvalidReference),
			expectedError: status.Error(codes.InvalidArgument, "error making get available slots request: invalid request [reference]"),
		},
		{
			desc:          "Appointment not found",
			clientErr:     mapper.ErrAppointmentNotFound,
			expectedError: status.Error(codes.NotFound, "error making get available slots request: no appointments found"),
		},
		{
			desc:          "Appointment out of range",
			clientErr:     mapper.ErrAppointmentOutOfRange,
			expectedError: status.Error(codes.OutOfRange, "error making get available slots request: appointment out of range"),
		},
		{
			desc:          "Unknown bad parameter",
			clientErr:     mapper.ErrInvalidRequest,
			expectedError: status.Error(codes.InvalidArgument, "error making get available slots request: invalid request"),
		},
		{
			desc:          "Unknown error",
			clientErr:     fmt.Errorf("unknown"),
			expectedError: status.Error(codes.Internal, "error making get available slots request: unknown"),
		},
	}

	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	mapper := &fakeMapper{}

	myAPIHandler := api.New(client, mapper)

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mapper.availabilityRequest = tc.req
			mapper.availabilityResponse = tc.expected

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

func Test_CreateBooking(t *testing.T) {
	now := time.Now().UTC().Format("02/01/2006 15:04:05")

	testCases := []struct {
		desc          string
		req           *lowribeck.CreateBookingRequest
		clientResp    *lowribeck.CreateBookingResponse
		clientErr     error
		expected      *contract.CreateBookingResponse
		expectedError error
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
		},
		{
			desc:          "Invalid postcode",
			clientErr:     mapper.NewInvalidRequestError(mapper.InvalidPostcode),
			expectedError: status.Error(codes.InvalidArgument, "error making booking request: invalid request [postcode]"),
		},
		{
			desc:          "Invalid reference",
			clientErr:     mapper.NewInvalidRequestError(mapper.InvalidReference),
			expectedError: status.Error(codes.InvalidArgument, "error making booking request: invalid request [reference]"),
		},
		{
			desc:          "Invalid site",
			clientErr:     mapper.NewInvalidRequestError(mapper.InvalidSite),
			expectedError: status.Error(codes.InvalidArgument, "error making booking request: invalid request [site]"),
		},
		{
			desc:          "Appointment not found",
			clientErr:     mapper.ErrAppointmentNotFound,
			expectedError: status.Error(codes.NotFound, "error making booking request: no appointments found"),
		},
		{
			desc:          "Appointment out of range",
			clientErr:     mapper.ErrAppointmentOutOfRange,
			expectedError: status.Error(codes.OutOfRange, "error making booking request: appointment out of range"),
		},
		{
			desc:          "Appointment already exists",
			clientErr:     mapper.ErrAppointmentAlreadyExists,
			expectedError: status.Error(codes.AlreadyExists, "error making booking request: appointment already exists"),
		},
		{
			desc:          "Unknown bad parameter",
			clientErr:     mapper.ErrInvalidRequest,
			expectedError: status.Error(codes.InvalidArgument, "error making booking request: invalid request"),
		},
		{
			desc:          "Unknown error",
			clientErr:     fmt.Errorf("unknown"),
			expectedError: status.Error(codes.Internal, "error making booking request: unknown"),
		},
	}

	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	mapper := &fakeMapper{}

	myAPIHandler := api.New(client, mapper)

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mapper.bookingRequest = tc.req
			mapper.booingResponse = tc.expected

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

type fakeMapper struct {
	availabilityRequest  *lowribeck.GetCalendarAvailabilityRequest
	availabilityResponse *contract.GetAvailableSlotsResponse
	bookingRequest       *lowribeck.CreateBookingRequest
	booingResponse       *contract.CreateBookingResponse
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
	return f.booingResponse, nil
}
