package api

import (
	"context"

	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BookingAPI struct {
	client client.UpstreamClient
	bookingv1.UnimplementedBookingAPIServer
}

func New(c client.UpstreamClient) *BookingAPI {
	return &BookingAPI{
		client: c,
	}
}

var (
	ErrNotImplemented = status.Error(codes.Internal, "not implemented")
)

func (b *BookingAPI) GetCustomerContactDetails(ctx context.Context, req *bookingv1.GetCustomerContactDetailsRequest) (*bookingv1.GetCustomerContactDetailsResponse, error) { // nolint:revive
	return nil, ErrNotImplemented
}

func (b *BookingAPI) GetCustomerBookings(ctx context.Context, req *bookingv1.GetCustomerBookingsRequest) (*bookingv1.GetCustomerBookingsResponse, error) { // nolint:revive
	return nil, ErrNotImplemented
}

func (b *BookingAPI) GetAvailableSlots(ctx context.Context, req *bookingv1.GetAvailableSlotsRequest) (*bookingv1.GetAvailableSlotsResponse, error) { // nolint:revive
	return nil, ErrNotImplemented
}

func (b *BookingAPI) CreateBooking(ctx context.Context, req *bookingv1.CreateBookingRequest) (*bookingv1.CreateBookingResponse, error) { // nolint:revive
	return nil, ErrNotImplemented
}

func (b *BookingAPI) RescheduleBooking(ctx context.Context, req *bookingv1.RescheduleBookingRequest) (*bookingv1.RescheduleBookingResponse, error) { // nolint:revive
	return nil, ErrNotImplemented
}
