package gateway_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
	mock_gateways "github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway/mocks"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	oops = errors.New("oops...")
)

func Test_GetAvailableSlots(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	lbC := mock_gateways.NewMockLowriBeckClient(ctrl)
	mai := fakeMachineAuthInjector{}
	mai.ctx = ctx

	myGw := gateway.NewLowriBeckGateway(mai, lbC)

	lbC.EXPECT().GetAvailableSlots(ctx, &lowribeckv1.GetAvailableSlotsRequest{
		Postcode:  "E2 1ZZ",
		Reference: "booking-reference-1",
	}).Return(&lowribeckv1.GetAvailableSlotsResponse{
		Slots: []*lowribeckv1.BookingSlot{
			{
				Date: &date.Date{
					Year:  2000,
					Month: 5,
					Day:   5,
				},
				StartTime: 10,
				EndTime:   15,
			},
			{
				Date: &date.Date{
					Year:  2001,
					Month: 5,
					Day:   5,
				},
				StartTime: 14,
				EndTime:   19,
			},
		},
	}, nil)

	actual := gateway.AvailableSlotsResponse{
		BookingSlots: []models.BookingSlot{
			{
				Date:      mustDate(t, "2000-05-05"),
				StartTime: 10,
				EndTime:   15,
			},
			{
				Date:      mustDate(t, "2001-05-05"),
				StartTime: 14,
				EndTime:   19,
			},
		},
	}

	expected, err := myGw.GetAvailableSlots(ctx, "E2 1ZZ", "booking-reference-1")
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, actual, cmpopts.IgnoreUnexported(date.Date{})) {
		t.Fatalf("expected: %+v, actual: %+v", expected, actual)
	}
}

func Test_GetAvailableSlots_HasError(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	ctx := context.Background()

	lbC := mock_gateways.NewMockLowriBeckClient(ctrl)
	mai := fakeMachineAuthInjector{}
	mai.ctx = ctx

	myGw := gateway.NewLowriBeckGateway(mai, lbC)

	availableSlotsRequest := &lowribeckv1.GetAvailableSlotsRequest{
		Postcode:  "E2 1ZZ",
		Reference: "booking-reference-1",
	}

	type testCases struct {
		description string
		setup       func(lbC *mock_gateways.MockLowriBeckClient)
		outputErr   error
	}

	tcs := []testCases{
		{
			description: "get available slots receives an internal status error",
			setup: func(lbC *mock_gateways.MockLowriBeckClient) {

				errorStatus := status.New(codes.Internal, "oops").Err()

				lbC.EXPECT().GetAvailableSlots(ctx, availableSlotsRequest).Return(&lowribeckv1.GetAvailableSlotsResponse{
					Slots: []*lowribeckv1.BookingSlot{},
				}, errorStatus)

			},
			outputErr: gateway.ErrInternal,
		},
		{
			description: "get available slots receives a not found error",
			setup: func(lbC *mock_gateways.MockLowriBeckClient) {

				errorStatus := status.New(codes.NotFound, "oops").Err()

				lbC.EXPECT().GetAvailableSlots(ctx, availableSlotsRequest).Return(&lowribeckv1.GetAvailableSlotsResponse{
					Slots: []*lowribeckv1.BookingSlot{},
				}, errorStatus)

			},
			outputErr: gateway.ErrNotFound,
		},
		{
			description: "get available slots receives an invalid argument error",
			setup: func(lbC *mock_gateways.MockLowriBeckClient) {

				errorStatus := status.New(codes.InvalidArgument, "oops").Err()

				lbC.EXPECT().GetAvailableSlots(ctx, availableSlotsRequest).Return(&lowribeckv1.GetAvailableSlotsResponse{
					Slots: []*lowribeckv1.BookingSlot{},
				}, errorStatus)

			},
			outputErr: gateway.ErrInvalidArgument,
		},
		{
			description: "get available slots receives an unhandled error",
			setup: func(lbC *mock_gateways.MockLowriBeckClient) {

				errorStatus := status.New(codes.Unauthenticated, "oops").Err()

				lbC.EXPECT().GetAvailableSlots(ctx, availableSlotsRequest).Return(&lowribeckv1.GetAvailableSlotsResponse{
					Slots: []*lowribeckv1.BookingSlot{},
				}, errorStatus)

			},
			outputErr: gateway.ErrUnhandledErrorCode,
		},
		{
			description: "get available slots receives an invalid argument status error with details - postcode",
			setup: func(lbC *mock_gateways.MockLowriBeckClient) {

				errorStatus, err := status.New(codes.InvalidArgument, "oops").WithDetails(&lowribeckv1.InvalidParameterResponse{
					Parameters: lowribeckv1.Parameters_PARAMETERS_POSTCODE,
				})

				if err != nil {
					t.Fatal(err)
				}

				lbC.EXPECT().GetAvailableSlots(ctx, availableSlotsRequest).Return(&lowribeckv1.GetAvailableSlotsResponse{
					Slots: []*lowribeckv1.BookingSlot{},
				}, errorStatus.Err())

			},
			outputErr: gateway.ErrInternalBadParameters,
		},
		{
			description: "get available slots receives an out of range error",
			setup: func(lbC *mock_gateways.MockLowriBeckClient) {

				errorStatus := status.New(codes.OutOfRange, "oops").Err()

				lbC.EXPECT().GetAvailableSlots(ctx, availableSlotsRequest).Return(&lowribeckv1.GetAvailableSlotsResponse{
					Slots: []*lowribeckv1.BookingSlot{},
				}, errorStatus)

			},
			outputErr: gateway.ErrOutOfRange,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(lbC)

			_, err := myGw.GetAvailableSlots(mai.ctx, "E2 1ZZ", "booking-reference-1")

			if diff := cmp.Diff(err.Error(), tc.outputErr.Error()); diff != "" {
				t.Fatal(diff)
			}
		})
	}

}

func Test_GetCreateBooking(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	lbC := mock_gateways.NewMockLowriBeckClient(ctrl)

	ctx := context.Background()
	mai := fakeMachineAuthInjector{}
	mai.ctx = ctx

	myGw := gateway.NewLowriBeckGateway(mai, lbC)

	postcode, bookingreference := "E2 1ZZ", "booking-reference-1"

	lbC.EXPECT().CreateBooking(ctx, &lowribeckv1.CreateBookingRequest{
		Postcode:  postcode,
		Reference: bookingreference,
		Slot: &lowribeckv1.BookingSlot{
			Date: &date.Date{
				Year:  2020,
				Month: 12,
				Day:   20,
			},
			StartTime: 15,
			EndTime:   19,
		},
		VulnerabilityDetails: &lowribeckv1.VulnerabilityDetails{
			Vulnerabilities: []lowribeckv1.Vulnerability{
				lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
			},
			Other: "Bad Knee",
		},
		ContactDetails: &lowribeckv1.ContactDetails{
			Title:     "Mr",
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "555-0777",
		},
	}).Return(&lowribeckv1.CreateBookingResponse{
		Success: true,
	}, nil)

	actual := gateway.CreateBookingResponse{
		Success: true,
	}

	expected, err := myGw.CreateBooking(ctx, postcode, bookingreference,
		models.BookingSlot{
			Date:      mustDate(t, "2020-12-20"),
			StartTime: 15,
			EndTime:   19,
		}, models.AccountDetails{
			Title:     "Mr",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "jdoe@example.com",
			Mobile:    "555-0777",
		}, []lowribeckv1.Vulnerability{
			lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
		}, "Bad Knee")
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, actual, cmpopts.IgnoreUnexported(date.Date{})) {
		t.Fatalf("expected: %+v, actual: %+v", expected, actual)
	}
}

func Test_GetCreateBooking_HasErrors(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	lbC := mock_gateways.NewMockLowriBeckClient(ctrl)
	mai := fakeMachineAuthInjector{}
	mai.ctx = ctx

	myGw := gateway.NewLowriBeckGateway(mai, lbC)

	type testCases struct {
		description string
		setup       func(lbC *mock_gateways.MockLowriBeckClient)
		outputErr   error
	}

	lbcreatebookingRequest := &lowribeckv1.CreateBookingRequest{
		Postcode:  "E2 1ZZ",
		Reference: "booking-reference-1",
		Slot: &lowribeckv1.BookingSlot{
			Date: &date.Date{
				Year:  2020,
				Month: 12,
				Day:   20,
			},
			StartTime: 15,
			EndTime:   19,
		},
		VulnerabilityDetails: &lowribeckv1.VulnerabilityDetails{
			Vulnerabilities: []lowribeckv1.Vulnerability{
				lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
			},
			Other: "Bad Knee",
		},
		ContactDetails: &lowribeckv1.ContactDetails{
			Title:     "Mr",
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "555-0777",
		},
	}

	tcs := []testCases{
		{
			description: "Create booking returns internal error status code",
			setup: func(lbC *mock_gateways.MockLowriBeckClient) {

				errorStatus := status.New(codes.Internal, "oops").Err()

				lbC.EXPECT().CreateBooking(ctx, lbcreatebookingRequest).Return(&lowribeckv1.CreateBookingResponse{
					Success: false,
				}, errorStatus)
			},
			outputErr: gateway.ErrInternal,
		},
		{
			description: "Create booking returns invalid argument status code",
			setup: func(lbC *mock_gateways.MockLowriBeckClient) {

				errorStatus := status.New(codes.InvalidArgument, "oops").Err()

				lbC.EXPECT().CreateBooking(ctx, lbcreatebookingRequest).Return(&lowribeckv1.CreateBookingResponse{
					Success: false,
				}, errorStatus)
			},
			outputErr: gateway.ErrInvalidArgument,
		},
		{
			description: "Create booking returns already exists status code",
			setup: func(lbC *mock_gateways.MockLowriBeckClient) {

				errorStatus := status.New(codes.AlreadyExists, "oops").Err()

				lbC.EXPECT().CreateBooking(ctx, lbcreatebookingRequest).Return(&lowribeckv1.CreateBookingResponse{
					Success: false,
				}, errorStatus)
			},
			outputErr: gateway.ErrAlreadyExists,
		},
		{
			description: "Create booking returns out of range status code",
			setup: func(lbC *mock_gateways.MockLowriBeckClient) {

				errorStatus := status.New(codes.OutOfRange, "oops").Err()

				lbC.EXPECT().CreateBooking(ctx, lbcreatebookingRequest).Return(&lowribeckv1.CreateBookingResponse{
					Success: false,
				}, errorStatus)
			},
			outputErr: gateway.ErrOutOfRange,
		},
		{
			description: "Create booking returns not found status code",
			setup: func(lbC *mock_gateways.MockLowriBeckClient) {

				errorStatus := status.New(codes.NotFound, "oops").Err()

				lbC.EXPECT().CreateBooking(ctx, lbcreatebookingRequest).Return(&lowribeckv1.CreateBookingResponse{
					Success: false,
				}, errorStatus)
			},
			outputErr: gateway.ErrNotFound,
		},
		{
			description: "Create booking returns an unhandled status code",
			setup: func(lbC *mock_gateways.MockLowriBeckClient) {

				errorStatus := status.New(codes.Unimplemented, "oops").Err()

				lbC.EXPECT().CreateBooking(ctx, lbcreatebookingRequest).Return(&lowribeckv1.CreateBookingResponse{
					Success: false,
				}, errorStatus)
			},
			outputErr: gateway.ErrUnhandledErrorCode,
		},
		{
			description: "Create booking returns an invalid argument status code and with details",
			setup: func(lbC *mock_gateways.MockLowriBeckClient) {

				errorStatus, err := status.New(codes.InvalidArgument, "oops").WithDetails(&lowribeckv1.InvalidParameterResponse{
					Parameters: lowribeckv1.Parameters_PARAMETERS_POSTCODE,
				})

				if err != nil {
					t.Fatal(err)
				}

				lbC.EXPECT().CreateBooking(ctx, lbcreatebookingRequest).Return(&lowribeckv1.CreateBookingResponse{
					Success: false,
				}, errorStatus.Err())
			},
			outputErr: gateway.ErrInternalBadParameters,
		},
		{
			description: "Create booking returns an invalid argument status code and with details - date",
			setup: func(lbC *mock_gateways.MockLowriBeckClient) {

				errorStatus, err := status.New(codes.InvalidArgument, "oops").WithDetails(&lowribeckv1.InvalidParameterResponse{
					Parameters: lowribeckv1.Parameters_PARAMETERS_APPOINTMENT_DATE,
				})

				if err != nil {
					t.Fatal(err)
				}

				lbC.EXPECT().CreateBooking(ctx, lbcreatebookingRequest).Return(&lowribeckv1.CreateBookingResponse{
					Success: false,
				}, errorStatus.Err())
			},
			outputErr: gateway.ErrInvalidAppointmentDate,
		},
		{
			description: "Create booking returns an invalid argument status code and with details - time",
			setup: func(lbC *mock_gateways.MockLowriBeckClient) {

				errorStatus, err := status.New(codes.InvalidArgument, "oops").WithDetails(&lowribeckv1.InvalidParameterResponse{
					Parameters: lowribeckv1.Parameters_PARAMETERS_APPOINTMENT_TIME,
				})

				if err != nil {
					t.Fatal(err)
				}

				lbC.EXPECT().CreateBooking(ctx, lbcreatebookingRequest).Return(&lowribeckv1.CreateBookingResponse{
					Success: false,
				}, errorStatus.Err())
			},
			outputErr: gateway.ErrInvalidAppointmentTime,
		},
	}

	actual := gateway.CreateBookingResponse{
		Success: false,
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			tc.setup(lbC)

			expected, err := myGw.CreateBooking(ctx, "E2 1ZZ", "booking-reference-1", models.BookingSlot{
				Date:      mustDate(t, "2020-12-20"),
				StartTime: 15,
				EndTime:   19,
			}, models.AccountDetails{
				Title:     "Mr",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "jdoe@example.com",
				Mobile:    "555-0777",
			}, []lowribeckv1.Vulnerability{
				lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
			}, "Bad Knee")

			if diff := cmp.Diff(err.Error(), tc.outputErr.Error()); diff != "" {
				t.Fatal(diff)
			}

			if !cmp.Equal(expected, actual, cmpopts.IgnoreUnexported(date.Date{})) {
				t.Fatalf("expected: %+v, actual: %+v", expected, actual)
			}
		})
	}
}

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()
	d, err := time.ParseInLocation(time.DateOnly, value, time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

type fakeMachineAuthInjector struct {
	ctx context.Context
}

func (fmai fakeMachineAuthInjector) ToCtx(_ context.Context) context.Context {
	return fmai.ctx
}
