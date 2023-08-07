package gateway_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
	mock_gateways "github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway/mocks"
	"google.golang.org/genproto/googleapis/type/date"
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
		Success:   true,
		ErrorCode: nil,
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
		Success:    false,
		ErrorCodes: *lowribeckv1.BookingErrorCodes_BOOKING_DUPLICATE_JOB_EXISTS.Enum(),
	}, nil)

	actual := gateway.CreateBookingResponse{
		Success:   false,
		ErrorCode: bookingv1.BookingErrorCodes_BOOKING_DUPLICATE_JOB_EXISTS.Enum(),
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

	if err != nil {
		t.Fatal(err)
	}

// 	if !cmp.Equal(expected, actual, cmpopts.IgnoreUnexported(date.Date{})) {
// 		t.Fatalf("expected: %+v, actual: %+v", expected, actual)
// 	}
// }

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()
	d, err := time.ParseInLocation(time.DateOnly, value, time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	return d
}
