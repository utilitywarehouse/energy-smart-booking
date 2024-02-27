//go:generate mockgen -source=partial_booking.go -destination ./mocks/partial_booking_mocks.go

package workers_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/workers"
	mocks "github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/workers/mocks"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

func Test_Run(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	mockPBStore := mocks.NewMockPartialBookingStore(ctrl)
	mockOccupancyStore := mocks.NewMockOccupancyStore(ctrl)
	mockPublisher := mocks.NewMockBookingPublisher(ctrl)
	alertThreshold := time.Hour

	worker := workers.NewPartialBookingWorker(mockPBStore, mockOccupancyStore, mockPublisher, alertThreshold)

	mockPBStore.EXPECT().GetPending(ctx).Return([]*models.PartialBooking{
		{
			CreatedAt: time.Now(),
			BookingID: "booking-id-1",
			Event: &bookingv1.BookingCreatedEvent{
				BookingId:   "booking-id-1",
				OccupancyId: "",
				Details: &bookingv1.Booking{
					Id:        "booking-id-1",
					AccountId: "account-id-1",
				},
			},
		},
		{
			CreatedAt: time.Now(),
			BookingID: "booking-id-2",
			Event: &bookingv1.BookingCreatedEvent{
				BookingId:   "booking-id-2",
				OccupancyId: "",
				Details: &bookingv1.Booking{
					Id:        "booking-id-2",
					AccountId: "account-id-2",
				},
			},
		},
	}, nil)

	mockOccupancyStore.EXPECT().GetOccupancyByAccountID(ctx, "account-id-1").Return(&models.Occupancy{
		OccupancyID: "occupancy-id-1",
	}, nil)
	mockOccupancyStore.EXPECT().GetOccupancyByAccountID(ctx, "account-id-2").Return(nil, store.ErrOccupancyNotFound)

	mockPublisher.EXPECT().Sink(ctx, &bookingv1.BookingCreatedEvent{
		BookingId:   "booking-id-1",
		OccupancyId: "occupancy-id-1",
		Details: &bookingv1.Booking{
			Id:        "booking-id-1",
			AccountId: "account-id-1",
		},
	}, gomock.Any()).Return(nil)

	mockPBStore.EXPECT().MarkAsDeleted(ctx, "booking-id-1", models.DeletionReasonBookingCreated).Return(nil)

	mockPBStore.EXPECT().UpdateRetries(ctx, "booking-id-2", 0).Return(nil)

	err := worker.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
