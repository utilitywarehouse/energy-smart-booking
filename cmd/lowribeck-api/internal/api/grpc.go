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
	"github.com/utilitywarehouse/energy-smart-booking/internal/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrUserUnauthorised = errors.New("user does not have required access")
)

type Auth interface {
	Authorize(ctx context.Context, params *auth.PolicyParams) (bool, error)
}

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
	auth   Auth
	contract.UnimplementedLowriBeckAPIServer
}

func New(c Client, m Mapper, a Auth) *LowriBeckAPI {
	return &LowriBeckAPI{
		client: c,
		mapper: m,
		auth:   a,
	}
}

func (l *LowriBeckAPI) GetAvailableSlots(ctx context.Context, req *contract.GetAvailableSlotsRequest) (*contract.GetAvailableSlotsResponse, error) {

	err := l.validateCredentials(ctx, auth.GetAction, auth.LowribeckAPIResource, "lowribeck-api")
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", err)
		default:
			return nil, status.Error(codes.Internal, "failed to validate credentials")
		}
	}

	requestID := uuid.New().ID()
	availabilityReq := l.mapper.AvailabilityRequest(requestID, req)
	resp, err := l.client.GetCalendarAvailability(ctx, availabilityReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error making get available slots request: %v", err)
	}

	mappedResp, mappedErr := l.mapper.AvailableSlotsResponse(resp)
	if mappedErr != nil {
		logrus.Errorf("error making get available slots request(%d) for reference(%s) and postcode(%s): %v", requestID, req.GetReference(), req.GetPostcode(), mappedErr)
		switch {
		case errors.Is(mappedErr, mapper.ErrAppointmentNotFound):
			return nil, status.Errorf(codes.NotFound, "error making get available slots request: %v", mappedErr)

		case errors.Is(mappedErr, mapper.ErrAppointmentOutOfRange):
			return nil, status.Errorf(codes.OutOfRange, "error making get available slots request: %v", mappedErr)

		case errors.Is(mappedErr, mapper.ErrInvalidRequest):
			return nil, status.Errorf(codes.InvalidArgument, "error making get available slots request: %v", mappedErr)
		default:
			if invErr, ok := mappedErr.(*mapper.InvalidRequestError); ok {
				invReqError, err := createInvalidRequestError("error making get available slots request: %v", invErr)
				if err != nil {
					logrus.Errorf("internal error for reference(%s) and postcode(%s): %v", req.GetReference(), req.GetPostcode(), err)
					return nil, status.Errorf(codes.Internal, "error making get available slots request: %v", err)
				}
				return nil, invReqError
			}
		}
		return nil, status.Errorf(codes.Internal, "error making get available slots request: %v", mappedErr)
	}
	return mappedResp, nil
}

func (l *LowriBeckAPI) CreateBooking(ctx context.Context, req *contract.CreateBookingRequest) (*contract.CreateBookingResponse, error) {

	err := l.validateCredentials(ctx, auth.CreateAction, auth.LowribeckAPIResource, "lowribeck-api")
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", err)
		default:
			return nil, status.Errorf(codes.Internal, "failed to validate credentials")
		}
	}

	requestID := uuid.New().ID()
	bookingReq, err := l.mapper.BookingRequest(requestID, req)
	if err != nil {
		logrus.Errorf("error mapping booking request for reference(%s) and postcode(%s): %v", req.GetReference(), req.GetPostcode(), err)
		return nil, status.Errorf(codes.InvalidArgument, "error mapping booking request: %v", err)
	}
	resp, err := l.client.CreateBooking(ctx, bookingReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error making booking request: %v", err)
	}

	mappedResp, mappedErr := l.mapper.BookingResponse(resp)
	if mappedErr != nil {
		logrus.Errorf("error making booking request(%d) for reference(%s) and postcode(%s): %v", requestID, req.GetReference(), req.GetPostcode(), mappedErr)
		switch {
		case errors.Is(mappedErr, mapper.ErrAppointmentNotFound):
			return nil, status.Errorf(codes.NotFound, "error making booking request: %v", mappedErr)

		case errors.Is(mappedErr, mapper.ErrAppointmentAlreadyExists):
			return nil, status.Errorf(codes.AlreadyExists, "error making booking request: %v", mappedErr)

		case errors.Is(mappedErr, mapper.ErrAppointmentOutOfRange):
			return nil, status.Errorf(codes.OutOfRange, "error making booking request: %v", mappedErr)

		case errors.Is(mappedErr, mapper.ErrInvalidRequest):
			return nil, status.Errorf(codes.InvalidArgument, "error making booking request: %v", mappedErr)

		default:
			if invErr, ok := mappedErr.(*mapper.InvalidRequestError); ok {
				invReqError, err := createInvalidRequestError("error making booking request: %v", invErr)
				if err != nil {
					logrus.Errorf("internal error for reference(%s) and postcode(%s): %v", req.GetReference(), req.GetPostcode(), err)
					return nil, status.Errorf(codes.Internal, "error making booking request: %v", err)
				}
				return nil, invReqError
			}
		}
		return nil, status.Errorf(codes.Internal, "error making booking request: %v", mappedErr)
	}
	return mappedResp, nil
}

func createInvalidRequestError(msg string, invErr *mapper.InvalidRequestError) (error, error) {
	var param contract.Parameters
	switch invErr.GetParameter() {
	case mapper.InvalidPostcode:
		param = contract.Parameters_PARAMETERS_POSTCODE
	case mapper.InvalidReference:
		param = contract.Parameters_PARAMETERS_REFERENCE
	case mapper.InvalidSite:
		param = contract.Parameters_PARAMETERS_SITE
	case mapper.InvalidAppointmentDate:
		param = contract.Parameters_PARAMETERS_APPOINTMENT_DATE
	case mapper.InvalidAppointmentTime:
		param = contract.Parameters_PARAMETERS_APPOINTMENT_TIME
	default:
		param = contract.Parameters_PARAMETERS_UNKNOWN
	}
	invReqError, err := status.New(codes.InvalidArgument, fmt.Sprintf(msg, invErr)).WithDetails(&contract.InvalidParameterResponse{
		Parameters: param,
	})
	if err != nil {
		return nil, err
	}
	return invReqError.Err(), nil
}

func (l *LowriBeckAPI) validateCredentials(ctx context.Context, action, resource, requestAccountID string) error {

	authorised, err := l.auth.Authorize(ctx, &auth.PolicyParams{
		Action:     action,
		Resource:   resource,
		ResourceID: requestAccountID,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"action":   action,
			"resource": resource,
		}).Error("Authorize error: ", err)
		return err
	}
	if !authorised {
		return ErrUserUnauthorised
	}

	return nil
}
