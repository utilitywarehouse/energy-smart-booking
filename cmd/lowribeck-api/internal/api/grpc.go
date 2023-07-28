package api

import (
	"context"

	"github.com/google/uuid"
	contract "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/lowribeck"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/mapper"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client interface {
	GetCalendarAvailability(context.Context, *lowribeck.GetCalendarAvailabilityRequest) (*lowribeck.GetCalendarAvailabilityResponse, error)
	CreateBooking(context.Context, *lowribeck.CreateBookingRequest) (*lowribeck.CreateBookingResponse, error)
}

type LowriBeckAPI struct {
	client Client
	contract.UnimplementedLowriBeckAPIServer
}

func New(c Client) *LowriBeckAPI {
	return &LowriBeckAPI{
		client: c,
	}
}

func (l *LowriBeckAPI) GetAvailableSlots(ctx context.Context, req *contract.GetAvailableSlotsRequest) (*contract.GetAvailableSlotsResponse, error) {
	requestID := uuid.New().ID()
	availabilityReq := mapper.MapAvailabilityRequest(requestID, req)
	resp, err := l.client.GetCalendarAvailability(ctx, availabilityReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error making request(%s) for reference(%s): %v", requestID, requestID, err)
	}
	return mapper.MapAvailableSlotsResponse(resp)
}

func (l *LowriBeckAPI) CreateBooking(ctx context.Context, req *contract.CreateBookingRequest) (*contract.CreateBookingResponse, error) {
	bookingReq := mapper.MapBookingRequest(req)

	resp, err := l.client.CreateBooking(ctx, bookingReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error making request: %s", err.Error())
	}

	return mapper.MapBookingResponse(resp), nil
}
