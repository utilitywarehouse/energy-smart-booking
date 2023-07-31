package api

import (
	"context"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	contract "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/lowribeck"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client interface {
	GetCalendarAvailability(context.Context, *lowribeck.GetCalendarAvailabilityRequest) (*lowribeck.GetCalendarAvailabilityResponse, error)
	CreateBooking(context.Context, *lowribeck.CreateBookingRequest) (*lowribeck.CreateBookingResponse, error)
}

type Mapper interface {
	AvailabilityRequest(uint32, *contract.GetAvailableSlotsRequest) *lowribeck.GetCalendarAvailabilityRequest
	AvailableSlotsResponse(*lowribeck.GetCalendarAvailabilityResponse) (*contract.GetAvailableSlotsResponse, error)
	BookingRequest(req *contract.CreateBookingRequest) *lowribeck.CreateBookingRequest
	BookingResponse(_ *lowribeck.CreateBookingResponse) *contract.CreateBookingResponse
}

type LowriBeckAPI struct {
	client Client
	mapper Mapper
	contract.UnimplementedLowriBeckAPIServer
}

func New(c Client, m Mapper) *LowriBeckAPI {
	return &LowriBeckAPI{
		client: c,
		mapper: m,
	}
}

func (l *LowriBeckAPI) GetAvailableSlots(ctx context.Context, req *contract.GetAvailableSlotsRequest) (*contract.GetAvailableSlotsResponse, error) {
	requestID := uuid.New().ID()
	availabilityReq := l.mapper.AvailabilityRequest(requestID, req)
	resp, err := l.client.GetCalendarAvailability(ctx, availabilityReq)
	if err != nil {
		logrus.Errorf("error making request(%d) for reference(%s): %v", requestID, req.GetReference(), err)
		return nil, status.Errorf(codes.Internal, "error making request: %s", err.Error())
	}
	return l.mapper.AvailableSlotsResponse(resp)
}

func (l *LowriBeckAPI) CreateBooking(ctx context.Context, req *contract.CreateBookingRequest) (*contract.CreateBookingResponse, error) {
	bookingReq := l.mapper.BookingRequest(req)

	resp, err := l.client.CreateBooking(ctx, bookingReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error making request: %s", err.Error())
	}

	return l.mapper.BookingResponse(resp), nil
}
