package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	contract "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/lowribeck"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/mapper"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/metrics"
	"github.com/utilitywarehouse/energy-smart-booking/internal/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const resourceID = "lowribeck-api"

var (
	ErrUserUnauthorised = errors.New("user does not have required access")
)

type Auth interface {
	Authorize(ctx context.Context, params *auth.PolicyParams) (bool, error)
}

type Client interface {
	GetCalendarAvailability(context.Context, *lowribeck.GetCalendarAvailabilityRequest) (*lowribeck.GetCalendarAvailabilityResponse, error)
	CreateBooking(context.Context, *lowribeck.CreateBookingRequest) (*lowribeck.CreateBookingResponse, error)
	UpdateContactDetails(context.Context, *lowribeck.UpdateContactDetailsRequest) (*lowribeck.UpdateContactDetailsResponse, error)

	// Point Of Sale Methods
	GetCalendarAvailabilityPointOfSale(context.Context, *lowribeck.GetCalendarAvailabilityRequest) (*lowribeck.GetCalendarAvailabilityResponse, error)
	CreateBookingPointOfSale(context.Context, *lowribeck.CreateBookingRequest) (*lowribeck.CreateBookingResponse, error)
}

type Mapper interface {
	AvailabilityRequest(uint32, *contract.GetAvailableSlotsRequest) *lowribeck.GetCalendarAvailabilityRequest
	AvailableSlotsResponse(*lowribeck.GetCalendarAvailabilityResponse) (*contract.GetAvailableSlotsResponse, error)
	BookingRequest(uint32, *contract.CreateBookingRequest) (*lowribeck.CreateBookingRequest, error)
	BookingResponse(*lowribeck.CreateBookingResponse) (*contract.CreateBookingResponse, error)
	UpdateContactDetailsRequest(uint32, *contract.UpdateContactDetailsRequest) *lowribeck.UpdateContactDetailsRequest
	UpdateContactDetailsResponse(*lowribeck.UpdateContactDetailsResponse) (*contract.UpdateContactDetailsResponse, error)

	//Point Of Sale Methods
	AvailabilityRequestPointOfSale(uint32, *contract.GetAvailableSlotsPointOfSaleRequest) (*lowribeck.GetCalendarAvailabilityRequest, error)
	BookingRequestPointOfSale(uint32, *contract.CreateBookingPointOfSaleRequest) (*lowribeck.CreateBookingRequest, error)
	AvailableSlotsPointOfSaleResponse(resp *lowribeck.GetCalendarAvailabilityResponse) (*contract.GetAvailableSlotsPointOfSaleResponse, error)
	BookingResponsePointOfSale(resp *lowribeck.CreateBookingResponse) (*contract.CreateBookingPointOfSaleResponse, error)
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

	err := l.validateCredentials(ctx, auth.GetAction)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.PermissionDenied, "user does not have access to this action, %s", err)
		default:
			return nil, status.Error(codes.Internal, "failed to validate credentials")
		}
	}

	requestID := uuid.New().ID()
	availabilityReq := l.mapper.AvailabilityRequest(requestID, req)
	resp, err := l.client.GetCalendarAvailability(ctx, availabilityReq)
	if err != nil {
		slog.Error("error making get available slots request", "error", err, "request_id", requestID, "reference", req.GetReference(), "postcode", req.GetPostcode())
		return nil, status.Errorf(codes.Internal, "error making get available slots request: %v", err)
	}

	mappedResp, mappedErr := l.mapper.AvailableSlotsResponse(resp)
	if mappedErr != nil {
		slog.Error("error in get available slots response", "error", mappedErr, "request_id", requestID, "reference", req.GetReference(), "postcode", req.GetPostcode())
		return nil, getStatusFromError("error making get available slots request: %v", metrics.GetAvailableSlots, mappedErr)
	}
	return mappedResp, nil
}

func (l *LowriBeckAPI) CreateBooking(ctx context.Context, req *contract.CreateBookingRequest) (*contract.CreateBookingResponse, error) {

	err := l.validateCredentials(ctx, auth.CreateAction)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.PermissionDenied, "user does not have access to this action, %s", err)
		default:
			return nil, status.Errorf(codes.Internal, "failed to validate credentials")
		}
	}

	requestID := uuid.New().ID()
	bookingReq, err := l.mapper.BookingRequest(requestID, req)
	if err != nil {
		slog.Error("error mapping booking request", "error", err, "reference", req.GetReference(), "postcode", req.GetPostcode())
		return nil, status.Errorf(codes.InvalidArgument, "error mapping booking request: %v", err)
	}
	resp, err := l.client.CreateBooking(ctx, bookingReq)
	if err != nil {
		slog.Error("error making booking request", "error", err, "request_id", requestID, "reference", req.GetReference(), "postcode", req.GetPostcode())
		return nil, status.Errorf(codes.Internal, "error making booking request: %v", err)
	}

	mappedResp, mappedErr := l.mapper.BookingResponse(resp)
	if mappedErr != nil {
		slog.Error("error in booking response", "request_id", requestID, "reference", req.GetReference(), "postcode", req.GetPostcode(), "error", mappedErr)
		return nil, getStatusFromError("error making booking request: %v", metrics.CreateBooking, mappedErr)
	}
	return mappedResp, nil
}

func (l *LowriBeckAPI) GetAvailableSlotsPointOfSale(ctx context.Context, req *contract.GetAvailableSlotsPointOfSaleRequest) (*contract.GetAvailableSlotsPointOfSaleResponse, error) {

	err := l.validateCredentials(ctx, auth.GetAction)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.PermissionDenied, "user does not have access to this action, %s", err)
		default:
			return nil, status.Error(codes.Internal, "failed to validate credentials")
		}
	}

	requestID := uuid.New().ID()
	availableSlotsRequest, err := l.mapper.AvailabilityRequestPointOfSale(requestID, req)
	if err != nil {
		if errors.Is(err, mapper.ErrInvalidElectricityTariffType) ||
			errors.Is(err, mapper.ErrInvalidGasTariffType) {
			slog.Error("error mapping booking point of sale request", "mpan", req.Mpan, "mprn", req.Mprn, "electricity_tariff", req.ElectricityTariffType.String(), "gas_tariff", req.GasTariffType.String(), "postcode", req.GetPostcode(), "error", err)
			return nil, status.Errorf(codes.Internal, "error making get available slots point of sale: %v", err)
		}
	}

	resp, err := l.client.GetCalendarAvailabilityPointOfSale(ctx, availableSlotsRequest)
	if err != nil {
		slog.Error("error making get available slots for point of sale", "request_id", requestID, "mpan", req.Mpan, "mprn", req.Mprn, "electricity_tariff", req.ElectricityTariffType.String(), "gas_tariff", req.GasTariffType.String(), "postcode", req.GetPostcode(), "error", err)
		return nil, status.Errorf(codes.Internal, "error making get available slots point of sale request: %v", err)
	}

	mappedResp, mappedErr := l.mapper.AvailableSlotsPointOfSaleResponse(resp)
	if mappedErr != nil {
		slog.Error("error in get available slots for point of sale", "request_id", requestID, "mpan", req.Mpan, "mprn", req.Mprn, "electricity_tariff", req.ElectricityTariffType.String(), "gas_tariff", req.GasTariffType.String(), "postcode", req.GetPostcode(), "error", mappedErr)
		return nil, getStatusFromError("error making get available slots point of sale request: %v", metrics.GetAvailableSlots, mappedErr)
	}
	return mappedResp, nil
}

func (l *LowriBeckAPI) CreateBookingPointOfSale(ctx context.Context, req *contract.CreateBookingPointOfSaleRequest) (*contract.CreateBookingPointOfSaleResponse, error) {

	err := l.validateCredentials(ctx, auth.CreateAction)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.PermissionDenied, "user does not have access to this action, %s", err)
		default:
			return nil, status.Errorf(codes.Internal, "failed to validate credentials")
		}
	}

	requestID := uuid.New().ID()
	bookingReq, err := l.mapper.BookingRequestPointOfSale(requestID, req)
	if err != nil {
		slog.Error("error mapping booking point of sale request", "mpan", req.Mpan, "mprn", req.Mprn, "electricity_tariff", req.ElectricityTariffType.String(), "gas_tariff", req.GasTariffType.String(), "postcode", req.SiteAddress.Paf.GetPostcode(), "error", err)
		if errors.Is(err, mapper.ErrInvalidElectricityTariffType) ||
			errors.Is(err, mapper.ErrInvalidGasTariffType) {
			return nil, status.Errorf(codes.Internal, "error mapping point of sale booking request: %v", err)
		}
		return nil, status.Errorf(codes.InvalidArgument, "error mapping point of sale booking request: %v", err)
	}
	resp, err := l.client.CreateBookingPointOfSale(ctx, bookingReq)
	if err != nil {
		slog.Error("error making booking point of sale request", "request_id", requestID, "mpan", req.Mpan, "mprn", req.Mprn, "elec_tariff", req.ElectricityTariffType.String(), "gas_tariff", req.GasTariffType.String(), "postcode", req.SiteAddress.Paf.GetPostcode(), "error", err)
		return nil, status.Errorf(codes.Internal, "error making booking point of sale request: %v", err)
	}

	mappedResp, mappedErr := l.mapper.BookingResponsePointOfSale(resp)
	if mappedErr != nil {
		slog.Error("error in booking point of sale request", "request_id", requestID, "mpan", req.Mpan, "mprn", req.Mprn, "elec_tariff", req.ElectricityTariffType.String(), "gas_tariff", req.GasTariffType.String(), "postcode", req.SiteAddress.Paf.GetPostcode(), "error", mappedErr)
		return nil, getStatusFromError("error making booking point of sale request: %v", metrics.CreateBooking, mappedErr)
	}
	return mappedResp, nil
}

func (l *LowriBeckAPI) UpdateContactDetails(ctx context.Context, req *contract.UpdateContactDetailsRequest) (*contract.UpdateContactDetailsResponse, error) {

	err := l.validateCredentials(ctx, auth.UpdateAction)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.PermissionDenied, "user does not have access to this action, %s", err)
		default:
			return nil, status.Errorf(codes.Internal, "failed to validate credentials")
		}
	}

	requestID := uuid.New().ID()
	updateContactReq := l.mapper.UpdateContactDetailsRequest(requestID, req)

	resp, err := l.client.UpdateContactDetails(ctx, updateContactReq)
	if err != nil {
		slog.Error("error making update contact details request", "request_id", requestID, "reference", req.GetReference(), "error", err)
		return nil, status.Errorf(codes.Internal, "error making update contact detail request: %v", err)
	}

	mappedResp, mappedErr := l.mapper.UpdateContactDetailsResponse(resp)
	if mappedErr != nil {
		slog.Error("error in update contact details response", "request_id", requestID, "reference", req.GetReference(), "error", mappedErr)
		return nil, getStatusFromError("error making update contact detail request: %v", metrics.UpdateContactDetails, mappedErr)
	}
	return mappedResp, nil
}

func createInvalidRequestError(msg, endpoint string, invErr *mapper.InvalidRequestError) (error, error) {
	var param contract.Parameters
	switch invErr.GetParameter() {
	case mapper.InvalidPostcode:
		metrics.LBErrorsCount.WithLabelValues(metrics.InvalidPostcode, endpoint).Inc()
		param = contract.Parameters_PARAMETERS_POSTCODE
	case mapper.InvalidReference:
		metrics.LBErrorsCount.WithLabelValues(metrics.InvalidReference, endpoint).Inc()
		param = contract.Parameters_PARAMETERS_REFERENCE
	case mapper.InvalidSite:
		metrics.LBErrorsCount.WithLabelValues(metrics.InvalidSite, endpoint).Inc()
		param = contract.Parameters_PARAMETERS_SITE
	case mapper.InvalidAppointmentDate:
		metrics.LBErrorsCount.WithLabelValues(metrics.InvalidAppointmentDate, endpoint).Inc()
		param = contract.Parameters_PARAMETERS_APPOINTMENT_DATE
	case mapper.InvalidAppointmentTime:
		metrics.LBErrorsCount.WithLabelValues(metrics.InvalidAppointmentTime, endpoint).Inc()
		param = contract.Parameters_PARAMETERS_APPOINTMENT_TIME
	case mapper.InvalidMPAN:
		metrics.LBErrorsCount.WithLabelValues(metrics.InvalidMPAN, endpoint).Inc()
		param = contract.Parameters_PARAMETERS_MPAN
	case mapper.InvalidMPRN:
		metrics.LBErrorsCount.WithLabelValues(metrics.InvalidMPRN, endpoint).Inc()
		param = contract.Parameters_PARAMETERS_MPRN
	default:
		metrics.LBErrorsCount.WithLabelValues(metrics.InvalidUnknownParameter, endpoint).Inc()
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

func getStatusFromError(formatMessage, endpoint string, err error) error {
	switch {
	case errors.Is(err, mapper.ErrAppointmentNotFound):
		metrics.LBErrorsCount.WithLabelValues(metrics.AppointmentNotFound, endpoint).Inc()
		return status.Errorf(codes.NotFound, formatMessage, err)

	case errors.Is(err, mapper.ErrAppointmentAlreadyExists):
		metrics.LBErrorsCount.WithLabelValues(metrics.AppointmentAlreadyExists, endpoint).Inc()
		return status.Errorf(codes.AlreadyExists, formatMessage, err)

	case errors.Is(err, mapper.ErrAppointmentOutOfRange):
		metrics.LBErrorsCount.WithLabelValues(metrics.AppointmentOutOfRange, endpoint).Inc()
		return status.Errorf(codes.OutOfRange, formatMessage, err)

	case errors.Is(err, mapper.ErrInternalError):
		metrics.LBErrorsCount.WithLabelValues(metrics.Internal, endpoint).Inc()
		return status.Errorf(codes.Internal, formatMessage, err)

	case errors.Is(err, mapper.ErrInvalidJobTypeCode),
		errors.Is(err, mapper.ErrInvalidElectricityJobTypeCode),
		errors.Is(err, mapper.ErrInvalidGasJobTypeCode):
		metrics.LBErrorsCount.WithLabelValues(metrics.InvalidJobTypeCode, endpoint).Inc()
		return status.Errorf(codes.Internal, formatMessage, err)

	default:
		if invErr, ok := err.(*mapper.InvalidRequestError); ok {
			invReqError, err := createInvalidRequestError(formatMessage, endpoint, invErr)
			if err != nil {
				return status.Errorf(codes.Internal, formatMessage, err)
			}
			return invReqError
		}
	}
	metrics.LBErrorsCount.WithLabelValues(metrics.Unknown, endpoint).Inc()
	return status.Errorf(codes.Internal, formatMessage, err)
}

func (l *LowriBeckAPI) validateCredentials(ctx context.Context, action string) error {

	authorised, err := l.auth.Authorize(ctx, &auth.PolicyParams{
		Action:     action,
		Resource:   auth.LowribeckAPIResource,
		ResourceID: resourceID,
	})
	if err != nil {
		slog.Error("authorise error", "error", err, "action", action, "resource", auth.LowribeckAPIResource)
		return err
	}
	if !authorised {
		return ErrUserUnauthorised
	}

	return nil
}
