package gateway_test

import (
	"context"
	"errors"
	"fmt"
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
	mai := mock_gateways.NewMockMachineAuthInjector(ctrl)

	mai.EXPECT().ToCtx(ctx).Return(ctx)

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

	ctx := context.Background()

	defer ctrl.Finish()

	lbC := mock_gateways.NewMockLowriBeckClient(ctrl)
	mai := mock_gateways.NewMockMachineAuthInjector(ctrl)

	myGw := gateway.NewLowriBeckGateway(mai, lbC)

	availableSlotsRequest := &lowribeckv1.GetAvailableSlotsRequest{
		Postcode:  "E2 1ZZ",
		Reference: "booking-reference-1",
	}

	type testCases struct {
		description string
		setup       func(lbC *mock_gateways.MockLowriBeckClient, mai *mock_gateways.MockMachineAuthInjector)
		outputErr   error
	}

	tcs := []testCases{
		{
			description: "get available slots receives an internal status error",
			setup: func(lbC *mock_gateways.MockLowriBeckClient, mai *mock_gateways.MockMachineAuthInjector) {

				errorStatus := status.New(codes.Internal, "oops").Err()

				mai.EXPECT().ToCtx(ctx).Return(ctx)

				lbC.EXPECT().GetAvailableSlots(ctx, availableSlotsRequest).Return(&lowribeckv1.GetAvailableSlotsResponse{
					Slots: []*lowribeckv1.BookingSlot{},
				}, errorStatus)

			},
			outputErr: fmt.Errorf("failed to get available slots, %w", gateway.ErrInternal),
		},
		{
			description: "get available slots receives a not found error",
			setup: func(lbC *mock_gateways.MockLowriBeckClient, mai *mock_gateways.MockMachineAuthInjector) {

				errorStatus := status.New(codes.NotFound, "oops").Err()

				mai.EXPECT().ToCtx(ctx).Return(ctx)

				lbC.EXPECT().GetAvailableSlots(ctx, availableSlotsRequest).Return(&lowribeckv1.GetAvailableSlotsResponse{
					Slots: []*lowribeckv1.BookingSlot{},
				}, errorStatus)

			},
			outputErr: fmt.Errorf("failed to get available slots, %w", gateway.ErrNotFound),
		},
		{
			description: "get available slots receives an invalid argument error",
			setup: func(lbC *mock_gateways.MockLowriBeckClient, mai *mock_gateways.MockMachineAuthInjector) {

				errorStatus := status.New(codes.InvalidArgument, "oops").Err()

				mai.EXPECT().ToCtx(ctx).Return(ctx)

				lbC.EXPECT().GetAvailableSlots(ctx, availableSlotsRequest).Return(&lowribeckv1.GetAvailableSlotsResponse{
					Slots: []*lowribeckv1.BookingSlot{},
				}, errorStatus)

			},
			outputErr: fmt.Errorf("failed to get available slots, %w", gateway.ErrInvalidArgument),
		},
		{
			description: "get available slots receives an unhandled error",
			setup: func(lbC *mock_gateways.MockLowriBeckClient, mai *mock_gateways.MockMachineAuthInjector) {

				errorStatus := status.New(codes.Unauthenticated, "oops").Err()

				mai.EXPECT().ToCtx(ctx).Return(ctx)

				lbC.EXPECT().GetAvailableSlots(ctx, availableSlotsRequest).Return(&lowribeckv1.GetAvailableSlotsResponse{
					Slots: []*lowribeckv1.BookingSlot{},
				}, errorStatus)

			},
			outputErr: gateway.ErrUnhandledErrorCode,
		},
		{
			description: "get available slots receives an invalid argument status error with ",
			setup: func(lbC *mock_gateways.MockLowriBeckClient, mai *mock_gateways.MockMachineAuthInjector) {

				errorStatus, err := status.New(codes.InvalidArgument, "oops").WithDetails(&lowribeckv1.InvalidParameterResponse{
					Parameters: lowribeckv1.Parameters_PARAMETERS_POSTCODE,
				})

				if err != nil {
					t.Fatal(err)
				}

				mai.EXPECT().ToCtx(ctx).Return(ctx)

				lbC.EXPECT().GetAvailableSlots(ctx, availableSlotsRequest).Return(&lowribeckv1.GetAvailableSlotsResponse{
					Slots: []*lowribeckv1.BookingSlot{},
				}, errorStatus.Err())

			},
			outputErr: fmt.Errorf("failed to get available slots, %w", gateway.ErrInternalBadParameters),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {

			tc.setup(lbC, mai)

			_, err := myGw.GetAvailableSlots(ctx, "E2 1ZZ", "booking-reference-1")

			if diff := cmp.Diff(err.Error(), tc.outputErr.Error()); diff != "" {
				t.Fatal(diff)
			}
		})
	}

}

// TODO - Seperate PR
// func Test_GetCreateBooking(t *testing.T) {
// 	ctrl := gomock.NewController(t)

// 	ctx := context.Background()

// 	defer ctrl.Finish()

// 	lbC := mock_gateways.NewMockLowriBeckClient(ctrl)
// 	mai := mock_gateways.NewMockMachineAuthInjector(ctrl)

// 	mai.EXPECT().ToCtx(ctx).Return(ctx)

// 	myGw := gateway.NewLowriBeckGateway(mai, lbC)

	lbC.EXPECT().CreateBooking(ctx, &lowribeckv1.CreateBookingRequest{
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
	}).Return(&lowribeckv1.CreateBookingResponse{
		Success: true,
	}, nil)

	actual := gateway.CreateBookingResponse{
		Success: true,
	}

// 	expected, err := myGw.CreateBooking(ctx, "E2 1ZZ", "booking-reference-1", models.BookingSlot{
// 		Date:      mustDate(t, "2020-12-20"),
// 		StartTime: 15,
// 		EndTime:   19,
// 	}, models.AccountDetails{
// 		Title:     "Mr",
// 		FirstName: "John",
// 		LastName:  "Doe",
// 		Email:     "jdoe@example.com",
// 		Mobile:    "555-0777",
// 	}, []lowribeckv1.Vulnerability{
// 		lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
// 	}, "Bad Knee")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if !cmp.Equal(expected, actual, cmpopts.IgnoreUnexported(date.Date{})) {
// 		t.Fatalf("expected: %+v, actual: %+v", expected, actual)
// 	}
// }

// func Test_GetCreateBooking_HasErrors(t *testing.T) {
// 	ctrl := gomock.NewController(t)

// 	ctx := context.Background()

// 	defer ctrl.Finish()

// 	lbC := mock_gateways.NewMockLowriBeckClient(ctrl)
// 	mai := mock_gateways.NewMockMachineAuthInjector(ctrl)

<<<<<<< HEAD
// 	mai.EXPECT().ToCtx(ctx).Return(ctx)

// 	myGw := gateway.NewLowriBeckGateway(mai, lbC)
=======
	myGw := gateway.NewLowriBeckGateway(mai, lbC)
>>>>>>> eb238c9 (wip)

	type testCases struct {
		description string
		setup       func(lbC *mock_gateways.MockLowriBeckClient, mai *mock_gateways.MockMachineAuthInjector)
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
			setup: func(lbC *mock_gateways.MockLowriBeckClient, mai *mock_gateways.MockMachineAuthInjector) {

				errorStatus := status.New(codes.Internal, "oops").Err()

				mai.EXPECT().ToCtx(ctx).Return(ctx)

				lbC.EXPECT().CreateBooking(ctx, lbcreatebookingRequest).Return(&lowribeckv1.CreateBookingResponse{
					Success: false,
				}, errorStatus)
			},
			outputErr: fmt.Errorf("failed to call lowribeck create booking, %w", gateway.ErrInternal),
		},
		{
			description: "Create booking returns invalid argument status code",
			setup: func(lbC *mock_gateways.MockLowriBeckClient, mai *mock_gateways.MockMachineAuthInjector) {

				errorStatus := status.New(codes.InvalidArgument, "oops").Err()

				mai.EXPECT().ToCtx(ctx).Return(ctx)

				lbC.EXPECT().CreateBooking(ctx, lbcreatebookingRequest).Return(&lowribeckv1.CreateBookingResponse{
					Success: false,
				}, errorStatus)
			},
			outputErr: fmt.Errorf("failed to call lowribeck create booking, %w", gateway.ErrInvalidArgument),
		},
		{
			description: "Create booking returns already exists status code",
			setup: func(lbC *mock_gateways.MockLowriBeckClient, mai *mock_gateways.MockMachineAuthInjector) {

				errorStatus := status.New(codes.AlreadyExists, "oops").Err()

				mai.EXPECT().ToCtx(ctx).Return(ctx)

				lbC.EXPECT().CreateBooking(ctx, lbcreatebookingRequest).Return(&lowribeckv1.CreateBookingResponse{
					Success: false,
				}, errorStatus)
			},
			outputErr: fmt.Errorf("failed to call lowribeck create booking, %w", gateway.ErrAlreadyExists),
		},
		{
			description: "Create booking returns out of range status code",
			setup: func(lbC *mock_gateways.MockLowriBeckClient, mai *mock_gateways.MockMachineAuthInjector) {

				errorStatus := status.New(codes.OutOfRange, "oops").Err()

				mai.EXPECT().ToCtx(ctx).Return(ctx)

				lbC.EXPECT().CreateBooking(ctx, lbcreatebookingRequest).Return(&lowribeckv1.CreateBookingResponse{
					Success: false,
				}, errorStatus)
			},
			outputErr: fmt.Errorf("failed to call lowribeck create booking, %w", gateway.ErrOutOfRange),
		},
		{
			description: "Create booking returns not found status code",
			setup: func(lbC *mock_gateways.MockLowriBeckClient, mai *mock_gateways.MockMachineAuthInjector) {

				errorStatus := status.New(codes.NotFound, "oops").Err()

				mai.EXPECT().ToCtx(ctx).Return(ctx)

				lbC.EXPECT().CreateBooking(ctx, lbcreatebookingRequest).Return(&lowribeckv1.CreateBookingResponse{
					Success: false,
				}, errorStatus)
			},
			outputErr: fmt.Errorf("failed to call lowribeck create booking, %w", gateway.ErrNotFound),
		},
		{
			description: "Create booking returns an unhandled status code",
			setup: func(lbC *mock_gateways.MockLowriBeckClient, mai *mock_gateways.MockMachineAuthInjector) {

				errorStatus := status.New(codes.Unimplemented, "oops").Err()

				mai.EXPECT().ToCtx(ctx).Return(ctx)

				lbC.EXPECT().CreateBooking(ctx, lbcreatebookingRequest).Return(&lowribeckv1.CreateBookingResponse{
					Success: false,
				}, errorStatus)
			},
			outputErr: gateway.ErrUnhandledErrorCode,
		},
		{
			description: "Create booking returns an invalid argument status code and with details",
			setup: func(lbC *mock_gateways.MockLowriBeckClient, mai *mock_gateways.MockMachineAuthInjector) {

				errorStatus, err := status.New(codes.InvalidArgument, "oops").WithDetails(&lowribeckv1.InvalidParameterResponse{
					Parameters: lowribeckv1.Parameters_PARAMETERS_POSTCODE,
				})

				if err != nil {
					t.Fatal(err)
				}

				mai.EXPECT().ToCtx(ctx).Return(ctx)

				lbC.EXPECT().CreateBooking(ctx, lbcreatebookingRequest).Return(&lowribeckv1.CreateBookingResponse{
					Success: false,
				}, errorStatus.Err())
			},
			outputErr: fmt.Errorf("failed to call lowribeck create booking, %w", gateway.ErrInternalBadParameters),
		},
	}

	actual := gateway.CreateBookingResponse{
		Success: false,
	}

<<<<<<< HEAD
// 	expected, err := myGw.CreateBooking(ctx, "E2 1ZZ", "booking-reference-1", models.BookingSlot{
// 		Date:      mustDate(t, "2020-12-20"),
// 		StartTime: 15,
// 		EndTime:   19,
// 	}, models.AccountDetails{
// 		Title:     "Mr",
// 		FirstName: "John",
// 		LastName:  "Doe",
// 		Email:     "jdoe@example.com",
// 		Mobile:    "555-0777",
// 	}, []lowribeckv1.Vulnerability{
// 		lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
// 	}, "Bad Knee")
=======
	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			tc.setup(lbC, mai)
>>>>>>> eb238c9 (wip)

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

<<<<<<< HEAD
// 	if !cmp.Equal(expected, actual, cmpopts.IgnoreUnexported(date.Date{})) {
// 		t.Fatalf("expected: %+v, actual: %+v", expected, actual)
// 	}
// }
=======
			if diff := cmp.Diff(err.Error(), tc.outputErr.Error()); diff != "" {
				t.Fatal(diff)
			}

			if !cmp.Equal(expected, actual, cmpopts.IgnoreUnexported(date.Date{})) {
				t.Fatalf("expected: %+v, actual: %+v", expected, actual)
			}
		})
	}
}
>>>>>>> eb238c9 (wip)

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()
	d, err := time.ParseInLocation(time.DateOnly, value, time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	return d
}
