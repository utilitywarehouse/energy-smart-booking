package gateway

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/helpers"

	"github.com/utilitywarehouse/uwos-go/telemetry/tracing"
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

type CreateBookingPointOfSaleResponse struct {
	Success     bool
	ReferenceID string
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

	availableSlots, err := g.client.GetAvailableSlots(g.mai.ToCtx(ctx), &lowribeckv1.GetAvailableSlotsRequest{
		Postcode:  postcode,
		Reference: reference,
	})
	if err != nil {
		return AvailableSlotsResponse{}, mapAvailableSlotsError(err)
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

func (g LowriBeckGateway) CreateBooking(ctx context.Context, postcode, reference string, slot models.BookingSlot, contactDetails models.AccountDetails, vulnerabilities []lowribeckv1.Vulnerability, other string) (_ CreateBookingResponse, err error) {
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
			Title:     contactDetails.Title,
			FirstName: contactDetails.FirstName,
			LastName:  contactDetails.LastName,
			Phone:     contactDetails.Mobile,
		},
	}

	reqAttr := helpers.CreateSpanAttribute(req, "CreateBookingRequest", span)
	span.AddEvent("request", trace.WithAttributes(reqAttr))

	bookingResponse, err := g.client.CreateBooking(g.mai.ToCtx(ctx), req)
	if err != nil {
		return CreateBookingResponse{Success: false}, mapCreateBookingError(err)
	}

	span.AddEvent("response", trace.WithAttributes(attribute.Bool("resp", bookingResponse.Success)))
	return CreateBookingResponse{
		Success: bookingResponse.Success,
	}, nil
}

func (g LowriBeckGateway) GetAvailableSlotsPointOfSale(ctx context.Context, postcode, mpan, mprn string, tariffElectricity, tariffGas lowribeckv1.TariffType) (_ AvailableSlotsResponse, err error) {
	ctx, span := tracing.Tracer().Start(ctx, "BookingAPI.LowriBeckGateway.GetPOSAvailableSlots",
		trace.WithAttributes(attribute.String("postcode", postcode)),
		trace.WithAttributes(attribute.String("lowribeck.mpan", mpan)),
		trace.WithAttributes(attribute.String("lowribeck.mprn", mprn)),
		trace.WithAttributes(attribute.String("lowribeck.electric.tariff", tariffElectricity.String())),
		trace.WithAttributes(attribute.String("lowribeck.gas.tariff", tariffGas.String())),
	)

	defer func() {
		tracing.RecordSpanError(span, err)
		span.End()
	}()

	availableSlots, err := g.client.GetAvailableSlotsPointOfSale(g.mai.ToCtx(ctx), &lowribeckv1.GetAvailableSlotsPointOfSaleRequest{
		Postcode:              postcode,
		Mpan:                  mpan,
		ElectricityTariffType: tariffElectricity,
		Mprn:                  mprn,
		GasTariffType:         tariffGas,
	})
	if err != nil {
		return AvailableSlotsResponse{}, mapAvailableSlotsPointOfSaleError(err)
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

func (g LowriBeckGateway) CreateBookingPointOfSale(ctx context.Context, postcode, mpan, mprn string, tariffElectricity, tariffGas lowribeckv1.TariffType, slot models.BookingSlot, contactDetails models.AccountDetails, vulnerabilities []lowribeckv1.Vulnerability, other string) (_ CreateBookingPointOfSaleResponse, err error) {
	ctx, span := tracing.Tracer().Start(ctx, "BookingAPI.CreatePOSBooking",
		trace.WithAttributes(attribute.String("postcode", postcode)),
		trace.WithAttributes(attribute.String("lowribeck.mpan", mpan)),
		trace.WithAttributes(attribute.String("lowribeck.mprn", mprn)),
		trace.WithAttributes(attribute.String("lowribeck.electric.tariff", tariffElectricity.String())),
		trace.WithAttributes(attribute.String("lowribeck.gas.tariff", tariffGas.String())),
	)
	defer func() {
		tracing.RecordSpanError(span, err)
		span.End()
	}()

	req := &lowribeckv1.CreateBookingPointOfSaleRequest{
		Postcode:              postcode,
		Mpan:                  mpan,
		ElectricityTariffType: tariffElectricity,
		Mprn:                  mprn,
		GasTariffType:         tariffGas,
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
			Title:     contactDetails.Title,
			FirstName: contactDetails.FirstName,
			LastName:  contactDetails.LastName,
			Phone:     contactDetails.Mobile,
		},
	}

	reqAttr := helpers.CreateSpanAttribute(req, "CreateBookingRequest", span)
	span.AddEvent("request", trace.WithAttributes(reqAttr))

	bookingResponse, err := g.client.CreateBookingPointOfSale(g.mai.ToCtx(ctx), req)
	if err != nil {
		return CreateBookingPointOfSaleResponse{Success: false}, mapCreateBookingPointOfSaleError(err)
	}

	span.AddEvent("response", trace.WithAttributes(attribute.Bool("resp", bookingResponse.Success)))

	return CreateBookingPointOfSaleResponse{
		Success:     bookingResponse.Success,
		ReferenceID: bookingResponse.Reference,
	}, nil
}

func mapAvailableSlotsError(err error) error {
	logrus.Errorf("failed to get available slots, %s, %s", ErrInternal, err)

	switch status.Convert(err).Code() {
	case codes.Internal:
		return ErrInternal
	case codes.NotFound:
		return ErrNotFound
	case codes.OutOfRange:
		return ErrOutOfRange
	case codes.InvalidArgument:

		details := status.Convert(err).Details()

		for _, detail := range details {
			switch x := detail.(type) {
			case *lowribeckv1.InvalidParameterResponse:
				logrus.Debugf("Found details in invalid argument error code, %s", x.GetParameters().String())

				switch x.GetParameters() {
				case lowribeckv1.Parameters_PARAMETERS_POSTCODE,
					lowribeckv1.Parameters_PARAMETERS_REFERENCE:
					return ErrInternalBadParameters
				}
			}
		}
		return ErrInvalidArgument

	default:
		return ErrUnhandledErrorCode
	}
}

func mapCreateBookingError(err error) error {

	switch status.Convert(err).Code() {
	case codes.Internal:
		return ErrInternal
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
					return ErrInternalBadParameters
				case lowribeckv1.Parameters_PARAMETERS_APPOINTMENT_DATE:
					return ErrInvalidAppointmentDate

				case lowribeckv1.Parameters_PARAMETERS_APPOINTMENT_TIME:
					return ErrInvalidAppointmentTime
				}
			}
		}
		return ErrInvalidArgument
	case codes.AlreadyExists:
		return ErrAlreadyExists
	case codes.OutOfRange:
		return ErrOutOfRange
	case codes.NotFound:
		return ErrNotFound
	default:
		return ErrUnhandledErrorCode
	}

}

func mapAvailableSlotsPointOfSaleError(err error) error {

	logrus.Errorf("failed to get available slots, %s, %s", ErrInternal, err)

	switch status.Convert(err).Code() {
	case codes.Internal:
		return ErrInternal
	case codes.NotFound:
		return ErrNotFound
	case codes.OutOfRange:
		return ErrOutOfRange
	case codes.InvalidArgument:

		details := status.Convert(err).Details()

		for _, detail := range details {
			switch x := detail.(type) {
			case *lowribeckv1.InvalidParameterResponse:
				logrus.Debugf("Found details in invalid argument error code, %s", x.GetParameters().String())

				switch x.GetParameters() {
				case lowribeckv1.Parameters_PARAMETERS_POSTCODE,
					lowribeckv1.Parameters_PARAMETERS_MPAN,
					lowribeckv1.Parameters_PARAMETERS_MPRN,
					lowribeckv1.Parameters_PARAMETERS_ELECTRICITY_TARIFF,
					lowribeckv1.Parameters_PARAMETERS_GAS_TARIFF:
					return ErrInternalBadParameters
				}
			}
		}
		return ErrInvalidArgument

	default:
		return ErrUnhandledErrorCode
	}

}

func mapCreateBookingPointOfSaleError(err error) error {

	switch status.Convert(err).Code() {
	case codes.Internal:
		return ErrInternal
	case codes.InvalidArgument:

		details := status.Convert(err).Details()

		for _, detail := range details {

			switch x := detail.(type) {
			case *lowribeckv1.InvalidParameterResponse:
				logrus.Debugf("Found details in invalid argument error code, %s", x.GetParameters().String())

				switch x.GetParameters() {
				case lowribeckv1.Parameters_PARAMETERS_POSTCODE,
					lowribeckv1.Parameters_PARAMETERS_MPAN,
					lowribeckv1.Parameters_PARAMETERS_MPRN,
					lowribeckv1.Parameters_PARAMETERS_ELECTRICITY_TARIFF,
					lowribeckv1.Parameters_PARAMETERS_GAS_TARIFF:
					return ErrInternalBadParameters
				case lowribeckv1.Parameters_PARAMETERS_APPOINTMENT_DATE:
					return ErrInvalidAppointmentDate

				case lowribeckv1.Parameters_PARAMETERS_APPOINTMENT_TIME:
					return ErrInvalidAppointmentTime
				}
			}
		}
		return ErrInvalidArgument
	case codes.AlreadyExists:
		return ErrAlreadyExists
	case codes.OutOfRange:
		return ErrOutOfRange
	case codes.NotFound:
		return ErrNotFound
	default:
		return ErrUnhandledErrorCode
	}
}
