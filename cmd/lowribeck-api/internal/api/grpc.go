package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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

type Mapper interface {
	AvailabilityRequest(uint32, *contract.GetAvailableSlotsRequest) *lowribeck.GetCalendarAvailabilityRequest
	AvailableSlotsResponse(*lowribeck.GetCalendarAvailabilityResponse) (*contract.GetAvailableSlotsResponse, error)
	BookingRequest(uint32, *contract.CreateBookingRequest) (*lowribeck.CreateBookingRequest, error)
	BookingResponse(*lowribeck.CreateBookingResponse) (*contract.CreateBookingResponse, error)
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
		logrus.Errorf("invalid available slots request(%d) for reference(%s) and postcode(%s): %v", requestID, req.GetReference(), req.GetPostcode(), err)
		switch {
		case errors.Is(err, mapper.ErrInvalidRequest):
			if invErr, ok := err.(*mapper.InvalidRequestError); ok {
				invReqError, err := createInvalidRequestError(invErr)
				if err != nil {
					logrus.Errorf("internal error for reference(%s) and postcode(%s): %v", req.GetReference(), req.GetPostcode(), err)
					return nil, status.Errorf(codes.Internal, "error making booking request: %s", err.Error())
				}
				return nil, invReqError
			}
		case errors.Is(err, mapper.ErrAppointmentNotFound):
			return nil, status.Errorf(codes.NotFound, "error making get available slots request: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error making get available slots request: %v", err)
	}
	return l.mapper.AvailableSlotsResponse(resp)
}

func (l *LowriBeckAPI) CreateBooking(ctx context.Context, req *contract.CreateBookingRequest) (*contract.CreateBookingResponse, error) {
	requestID := uuid.New().ID()
	bookingReq, err := l.mapper.BookingRequest(requestID, req)
	if err != nil {
		logrus.Errorf("error mapping booking request for reference(%s) and postcode(%s): %v", req.GetReference(), req.GetPostcode(), err)
		return nil, status.Errorf(codes.InvalidArgument, "error mapping booking request: %v", err)
	}
	resp, err := l.client.CreateBooking(ctx, bookingReq)
	if err != nil {
		logrus.Errorf("error making booking request(%d) for reference(%s) and postcode(%s): %v", requestID, req.GetReference(), req.GetPostcode(), err)
		switch {
		case errors.Is(err, mapper.ErrInvalidRequest):
			if invErr, ok := err.(*mapper.InvalidRequestError); ok {
				invReqError, err := createInvalidRequestError(invErr)
				if err != nil {
					logrus.Errorf("internal error for reference(%s) and postcode(%s): %v", req.GetReference(), req.GetPostcode(), err)
					return nil, status.Errorf(codes.Internal, "error making booking request: %s", err.Error())
				}
				return nil, invReqError
			}
		case errors.Is(err, mapper.ErrAppointmentNotFound):
			return nil, status.Errorf(codes.NotFound, "error making booking request: %v", err)

		case errors.Is(err, mapper.ErrAppointmentAlreadyExists):
			return nil, status.Errorf(codes.AlreadyExists, "error making booking request: %v", err)

		case errors.Is(err, mapper.ErrAppointmentOutOfRange):
			return nil, status.Errorf(codes.OutOfRange, "error making booking request: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error making booking request: %s", err.Error())
	}
	return l.mapper.BookingResponse(resp)
}

func createInvalidRequestError(invErr *mapper.InvalidRequestError) (error, error) {
	var param contract.Parameters
	switch invErr.GetParameter() {
	case mapper.InvalidPostcode:
		param = contract.Parameters_PARAMETERS_POSTCODE
	case mapper.InvalidReference:
		param = contract.Parameters_PARAMETERS_REFERENCE
	case mapper.InvalidSite:
		param = contract.Parameters_PARAMETERS_SITE
	case mapper.InvalidAppointmentDate:
		param = contract.Parameters_PARAMETERS_APPPOINTMENT_DATE
	case mapper.InvalidAppointmentTime:
		param = contract.Parameters_PARAMETERS_APPPOINTMENT_TIME
	default:
		param = contract.Parameters_PARAMETERS_UNKNOWN
	}
	invReqError, err := status.New(codes.InvalidArgument, fmt.Sprintf("error making get available slots request: %v", invErr)).WithDetails(&contract.InvalidParameterResponse{
		Parameters: param,
	})
	if err != nil {
		return nil, err
	}
	return invReqError.Err(), nil
}
