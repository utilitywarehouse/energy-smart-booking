package gateway

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/helpers"

	"github.com/utilitywarehouse/uwos-go/v1/telemetry/tracing"
	"go.opentelemetry.io/otel/attribute"
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

func (g LowriBeckGateway) GetAvailableSlots(ctx context.Context, postcode, reference string) (_ AvailableSlotsResponse, err error) {
	ctx, span := tracing.Tracer().Start(ctx, "BookingAPI.LowriBeckGateway.GetAvailableSlots",
		trace.WithAttributes(attribute.String("postcode", postcode)),
		trace.WithAttributes(attribute.String("lowribeck.reference", reference)),
	)

	defer func() {
		tracing.RecordSpanError(span, err)
		span.End()
	}()

	span.AddEvent("request", trace.WithAttributes(attribute.String("postcode", postcode), attribute.String("reference", reference)))

	availableSlots, err := g.client.GetAvailableSlots(g.mai.ToCtx(ctx), &lowribeckv1.GetAvailableSlotsRequest{
		Postcode:  postcode,
		Reference: reference,
	})
	if err != nil {
		logrus.Errorf("failed to get available slots, %s, %s", ErrInternal, err)

		switch status.Convert(err).Code() {
		case codes.Internal:
			return AvailableSlotsResponse{}, ErrInternal
		case codes.NotFound:
			return AvailableSlotsResponse{}, ErrNotFound
		case codes.OutOfRange:
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
						return AvailableSlotsResponse{}, ErrInternalBadParameters
					}
				}
			}
			return AvailableSlotsResponse{}, ErrInvalidArgument
		}
		return AvailableSlotsResponse{}, ErrUnhandledErrorCode
	}

	slots := []models.BookingSlot{}

	for _, elem := range availableSlots.GetSlots() {
		slots = append(slots, models.BookingSlot{
			Date:      time.Date(int(elem.Date.Year), time.Month(elem.Date.Month), int(elem.Date.Day), 0, 0, 0, 0, time.UTC),
			StartTime: int(elem.GetStartTime()),
			EndTime:   int(elem.GetEndTime()),
		})
	}

	slotsAttr := helpers.CreateSpanAttribute(slots, "slots", span)
	span.AddEvent("response", trace.WithAttributes(slotsAttr))

	return AvailableSlotsResponse{
		BookingSlots: slots,
	}, nil
}

func (g LowriBeckGateway) CreateBooking(ctx context.Context, postcode, reference string, slot models.BookingSlot, accountDetails models.AccountDetails, vulnerabilities []lowribeckv1.Vulnerability, other string) (_ CreateBookingResponse, err error) {
	ctx, span := tracing.Tracer().Start(ctx, "BookingAPI.CreateBooking",
		trace.WithAttributes(attribute.String("postcode", postcode)),
		trace.WithAttributes(attribute.String("lowribeck.reference", reference)),
	)
	defer func() {
		tracing.RecordSpanError(span, err)
		span.End()
	}()

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

	reqAttr := helpers.CreateSpanAttribute(req, "CreateBookingRequest", span)
	span.AddEvent("request", trace.WithAttributes(reqAttr))
	bookingResponse, err := g.client.CreateBooking(g.mai.ToCtx(ctx), req)
	if err != nil {
		switch status.Convert(err).Code() {
		case codes.Internal:
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
						return CreateBookingResponse{
							Success: false,
						}, ErrInternalBadParameters
					case lowribeckv1.Parameters_PARAMETERS_APPOINTMENT_DATE:
						return CreateBookingResponse{
							Success: false,
						}, ErrInvalidAppointmentDate

					case lowribeckv1.Parameters_PARAMETERS_APPOINTMENT_TIME:
						return CreateBookingResponse{
							Success: false,
						}, ErrInvalidAppointmentTime
					}
				}
			}
			return CreateBookingResponse{Success: false}, ErrInvalidArgument
		case codes.AlreadyExists:
			return CreateBookingResponse{Success: false}, ErrAlreadyExists
		case codes.OutOfRange:
			return CreateBookingResponse{Success: false}, ErrOutOfRange
		case codes.NotFound:
			return CreateBookingResponse{Success: false}, ErrNotFound
		default:
			return CreateBookingResponse{Success: false}, ErrUnhandledErrorCode
		}
	}
	span.AddEvent("response", trace.WithAttributes(attribute.Bool("resp", bookingResponse.Success)))
	return CreateBookingResponse{
		Success: bookingResponse.Success,
	}, nil
}
