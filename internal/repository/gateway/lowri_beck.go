package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/uwos-go/v1/telemetry/tracing"
	"go.opentelemetry.io/otel/attribute"
	tracecodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidArgument        = errors.New("invalid arguments")
	ErrInvalidAppointmentDate = errors.New("invalid appointment date")
	ErrInvalidAppointmentTime = errors.New("invalid appointment time")
	ErrInternalBadParameters  = errors.New("internal bad parameters")
	ErrNotFound               = errors.New("not found")
	ErrInternal               = errors.New("internal error")
	ErrUnhandledErrorCode     = errors.New("error code not handled")
	ErrAlreadyExists          = errors.New("already exists")
	ErrOutOfRange             = errors.New("out of range")
)

type LowriBeckGateway struct {
	mai    MachineAuthInjector
	client LowriBeckClient
}

func NewLowriBeckGateway(mai MachineAuthInjector, client LowriBeckClient) LowriBeckGateway {
	return LowriBeckGateway{mai, client}
}

type AvailableSlotsResponse struct {
	BookingSlots []models.BookingSlot
}

type CreateBookingResponse struct {
	Success bool
}

func (g LowriBeckGateway) GetAvailableSlots(ctx context.Context, postcode, reference string) (AvailableSlotsResponse, error) {
	req := &lowribeckv1.GetAvailableSlotsRequest{
		Postcode:  postcode,
		Reference: reference,
	}

	ctx, span := tracing.Tracer().Start(ctx, fmt.Sprintf("BookingAPI.%s", "GetAvailableSLots"),
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.AddEvent("request", trace.WithAttributes(attribute.String("request", fmt.Sprintf("%v", req))))

	availableSlots, err := g.client.GetAvailableSlots(g.mai.ToCtx(ctx), req)
	if err != nil {
		logrus.Errorf("failed to get available slots, %s, %s", ErrInternal, err)
		span.SetStatus(tracecodes.Error, err.Error())
		span.RecordError(err)

		switch status.Convert(err).Code() {
		case codes.Internal:
			span.SetAttributes(attribute.String("code", ErrInternal.Error()))
			return AvailableSlotsResponse{}, ErrInternal
		case codes.NotFound:
			span.SetAttributes(attribute.String("code", ErrNotFound.Error()))
			return AvailableSlotsResponse{}, ErrNotFound
		case codes.OutOfRange:
			span.SetAttributes(attribute.String("code", ErrOutOfRange.Error()))
			return AvailableSlotsResponse{}, ErrOutOfRange
		case codes.InvalidArgument:

			details := status.Convert(err).Details()

			for _, detail := range details {
				switch x := detail.(type) {
				case *lowribeckv1.InvalidParameterResponse:
					logrus.Debugf("Found details in invalid argument error code, %s", x.GetParameters().String())

					switch x.GetParameters() {
					case lowribeckv1.Parameters_PARAMETERS_POSTCODE,
						lowribeckv1.Parameters_PARAMETERS_REFERENCE:
						span.SetAttributes(attribute.String("code", ErrInternalBadParameters.Error()))
						return AvailableSlotsResponse{}, ErrInternalBadParameters
					}
				}
			}
			span.SetAttributes(attribute.String("code", ErrInvalidArgument.Error()))
			return AvailableSlotsResponse{}, ErrInvalidArgument
		}
		span.SetAttributes(attribute.String("code", ErrUnhandledErrorCode.Error()))
		return AvailableSlotsResponse{}, ErrUnhandledErrorCode
	}
	span.AddEvent("response", trace.WithAttributes(attribute.String("resp", fmt.Sprintf("%v", availableSlots.GetSlots()))))

	slots := []models.BookingSlot{}

	for _, elem := range availableSlots.GetSlots() {
		slots = append(slots, models.BookingSlot{
			Date:      time.Date(int(elem.Date.Year), time.Month(elem.Date.Month), int(elem.Date.Day), 0, 0, 0, 0, time.UTC),
			StartTime: int(elem.GetStartTime()),
			EndTime:   int(elem.GetEndTime()),
		})
	}

	return AvailableSlotsResponse{
		BookingSlots: slots,
	}, nil
}

func (g LowriBeckGateway) CreateBooking(ctx context.Context, postcode, reference string, slot models.BookingSlot, accountDetails models.AccountDetails, vulnerabilities []lowribeckv1.Vulnerability, other string) (CreateBookingResponse, error) {
	ctx, span := tracing.Tracer().Start(ctx, fmt.Sprintf("BookingAPI.%s", "CreateBooking"),
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	req := &lowribeckv1.CreateBookingRequest{
		Postcode:  postcode,
		Reference: reference,
		Slot: &lowribeckv1.BookingSlot{
			Date: &date.Date{
				Year:  int32(slot.Date.Year()),
				Month: int32(slot.Date.Month()),
				Day:   int32(slot.Date.Day()),
			},
			StartTime: int32(slot.StartTime),
			EndTime:   int32(slot.EndTime),
		},
		VulnerabilityDetails: &lowribeckv1.VulnerabilityDetails{
			Vulnerabilities: vulnerabilities,
			Other:           other,
		},
		ContactDetails: &lowribeckv1.ContactDetails{
			Title:     accountDetails.Title,
			FirstName: accountDetails.FirstName,
			LastName:  accountDetails.LastName,
			Phone:     accountDetails.Mobile,
		},
	}

	span.AddEvent("request", trace.WithAttributes(attribute.String("req", fmt.Sprintf("%v", req))))
	bookingResponse, err := g.client.CreateBooking(g.mai.ToCtx(ctx), req)
	if err != nil {
		span.SetStatus(tracecodes.Error, err.Error())
		span.RecordError(err)

		switch status.Convert(err).Code() {
		case codes.Internal:
			span.SetAttributes(attribute.String("code", ErrInternal.Error()))
			return CreateBookingResponse{Success: false}, ErrInternal
		case codes.InvalidArgument:

			details := status.Convert(err).Details()

			for _, detail := range details {

				switch x := detail.(type) {
				case *lowribeckv1.InvalidParameterResponse:
					logrus.Debugf("Found details in invalid argument error code, %s", x.GetParameters().String())

					switch x.GetParameters() {
					case lowribeckv1.Parameters_PARAMETERS_POSTCODE,
						lowribeckv1.Parameters_PARAMETERS_REFERENCE,
						lowribeckv1.Parameters_PARAMETERS_SITE:
						span.SetAttributes(attribute.String("code", ErrInternalBadParameters.Error()))
						return CreateBookingResponse{
							Success: false,
						}, ErrInternalBadParameters
					case lowribeckv1.Parameters_PARAMETERS_APPOINTMENT_DATE:
						span.SetAttributes(attribute.String("code", ErrInvalidAppointmentDate.Error()))
						return CreateBookingResponse{
							Success: false,
						}, ErrInvalidAppointmentDate

					case lowribeckv1.Parameters_PARAMETERS_APPOINTMENT_TIME:
						span.SetAttributes(attribute.String("code", ErrInvalidAppointmentTime.Error()))
						return CreateBookingResponse{
							Success: false,
						}, ErrInvalidAppointmentTime
					}
				}
			}
			span.SetAttributes(attribute.String("code", ErrInvalidArgument.Error()))
			return CreateBookingResponse{Success: false}, ErrInvalidArgument
		case codes.AlreadyExists:
			span.SetAttributes(attribute.String("code", ErrAlreadyExists.Error()))
			return CreateBookingResponse{Success: false}, ErrAlreadyExists
		case codes.OutOfRange:
			span.SetAttributes(attribute.String("code", ErrOutOfRange.Error()))
			return CreateBookingResponse{Success: false}, ErrOutOfRange
		case codes.NotFound:
			span.SetAttributes(attribute.String("code", ErrNotFound.Error()))
			return CreateBookingResponse{Success: false}, ErrNotFound
		default:
			span.SetAttributes(attribute.String("code", ErrUnhandledErrorCode.Error()))
			return CreateBookingResponse{Success: false}, ErrUnhandledErrorCode
		}
	}
	span.AddEvent("response", trace.WithAttributes(attribute.String("resp", fmt.Sprintf("%v", bookingResponse))))
	return CreateBookingResponse{
		Success: bookingResponse.Success,
	}, nil
}
