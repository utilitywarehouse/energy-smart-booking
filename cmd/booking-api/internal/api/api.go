package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	addressv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/address/v1"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/internal/auth"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
	"github.com/utilitywarehouse/uwos-go/v1/telemetry/tracing"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var (
	ErrUserUnauthorised = errors.New("user does not have required access")
)

type BookingDomain interface {
	GetCustomerContactDetails(ctx context.Context, accountID string) (models.Account, error)
	GetAccountAddressByAccountID(ctx context.Context, accountID string) (models.AccountAddress, error)
	GetCustomerBookings(ctx context.Context, accountID string) ([]*bookingv1.Booking, error)
	CreateBooking(ctx context.Context, params domain.CreateBookingParams) (domain.CreateBookingResponse, error)
	GetAvailableSlots(ctx context.Context, params domain.GetAvailableSlotsParams) (domain.GetAvailableSlotsResponse, error)
	RescheduleBooking(ctx context.Context, params domain.RescheduleBookingParams) (domain.RescheduleBookingResponse, error)
}

type BookingPublisher interface {
	Sink(ctx context.Context, proto proto.Message, at time.Time) error
}

type Auth interface {
	Authorize(ctx context.Context, params *auth.PolicyParams) (bool, error)
}

type BookingAPI struct {
	bookingDomain BookingDomain
	publisher     BookingPublisher
	auth          Auth
	bookingv1.UnimplementedBookingAPIServer
}

type accountIder interface {
	GetAccountId() string
}

func New(bookingDomain BookingDomain, publisher BookingPublisher, auth Auth) *BookingAPI {
	return &BookingAPI{
		bookingDomain: bookingDomain,
		publisher:     publisher,
		auth:          auth,
	}
}

func (b *BookingAPI) GetCustomerContactDetails(ctx context.Context, req *bookingv1.GetCustomerContactDetailsRequest) (_ *bookingv1.GetCustomerContactDetailsResponse, err error) { // nolint:revive
	ctx, span := tracing.Tracer().Start(ctx, "BookingAPI.GetCustomerContactDetails")
	defer func() {
		tracing.RecordSpanError(span, err) // nolint: errcheck
		span.End()
	}()

	err = b.validateCredentials(ctx, auth.GetAction, auth.AccountResource, req.AccountId)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", err)
		default:
			return nil, status.Errorf(codes.Internal, "failed to validate credentials")
		}
	}

	if err := validateRequest(req); err != nil {
		return nil, err
	}

	account, err := b.bookingDomain.GetCustomerContactDetails(ctx, req.GetAccountId())
	if err != nil {
		switch {
		case errors.Is(err, gateway.ErrAccountNotFound):
			return nil, status.Error(codes.NotFound, fmt.Sprintf("failed to get customer contact details, %s", err))
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get customer contact details, %s", err))
	}

	return &bookingv1.GetCustomerContactDetailsResponse{
		Title:     account.Details.Title,
		FirstName: account.Details.FirstName,
		LastName:  account.Details.LastName,
		Phone:     account.Details.Mobile,
		Email:     account.Details.Email,
	}, nil
}

func (b *BookingAPI) GetCustomerSiteAddress(ctx context.Context, req *bookingv1.GetCustomerSiteAddressRequest) (_ *bookingv1.GetCustomerSiteAddressResponse, err error) {
	ctx, span := tracing.Tracer().Start(ctx, "BookingAPI.GetCustomerSiteAddress")
	defer func() {
		tracing.RecordSpanError(span, err) // nolint: errcheck
		span.End()
	}()

	err = b.validateCredentials(ctx, auth.GetAction, auth.AccountResource, req.AccountId)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", err)
		default:
			return nil, status.Errorf(codes.Internal, "failed to validate credentials")
		}
	}

	if err := validateRequest(req); err != nil {
		return nil, err
	}

	accountAddress, err := b.bookingDomain.GetAccountAddressByAccountID(ctx, req.GetAccountId())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNoEligibleOccupanciesFound):
			return nil, status.Error(codes.NotFound, fmt.Sprintf("failed to get account address by account id %s", err))
		default:
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get account address by account id %s", err))
		}
	}

	return &bookingv1.GetCustomerSiteAddressResponse{
		SiteAddress: &addressv1.Address{
			Uprn: accountAddress.UPRN,
			Paf: &addressv1.Address_PAF{
				Organisation:            accountAddress.PAF.Organisation,
				Department:              accountAddress.PAF.Department,
				SubBuilding:             accountAddress.PAF.SubBuilding,
				BuildingName:            accountAddress.PAF.BuildingName,
				BuildingNumber:          accountAddress.PAF.BuildingNumber,
				DependentThoroughfare:   accountAddress.PAF.DependentThoroughfare,
				Thoroughfare:            accountAddress.PAF.Thoroughfare,
				DoubleDependentLocality: accountAddress.PAF.DoubleDependentLocality,
				DependentLocality:       accountAddress.PAF.DependentLocality,
				PostTown:                accountAddress.PAF.PostTown,
				Postcode:                accountAddress.PAF.Postcode,
			},
		},
	}, nil
}

func (b *BookingAPI) GetCustomerBookings(ctx context.Context, req *bookingv1.GetCustomerBookingsRequest) (_ *bookingv1.GetCustomerBookingsResponse, err error) { // nolint:revive
	ctx, span := tracing.Tracer().Start(ctx, "BookingAPI.GetCustomerBookings")
	defer func() {
		tracing.RecordSpanError(span, err) // nolint: errcheck
		span.End()
	}()

	err = b.validateCredentials(ctx, auth.GetAction, auth.AccountBookingResource, req.AccountId)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", err)
		default:
			return nil, status.Errorf(codes.Internal, "failed to validate credentials")
		}
	}

	bookings, err := b.bookingDomain.GetCustomerBookings(ctx, req.GetAccountId())
	if err != nil {
		return &bookingv1.GetCustomerBookingsResponse{Bookings: nil}, status.Errorf(codes.Internal, "failed to get customer bookings, %s", err)
	}

	return &bookingv1.GetCustomerBookingsResponse{Bookings: bookings}, nil
}

func (b *BookingAPI) GetAvailableSlots(ctx context.Context, req *bookingv1.GetAvailableSlotsRequest) (_ *bookingv1.GetAvailableSlotsResponse, err error) { // nolint:revive
	ctx, span := tracing.Tracer().Start(ctx, "BookingAPI.GetAvailableSlots")
	defer func() {
		tracing.RecordSpanError(span, err) // nolint: errcheck
		span.End()
	}()

	err = b.validateCredentials(ctx, auth.GetAction, auth.AccountBookingResource, req.AccountId)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", err)
		default:
			return nil, status.Error(codes.Internal, "failed to validate credentials")
		}
	}

	if err := validateRequest(req); err != nil {
		return nil, err
	}

	if req.From == nil {
		return nil, status.Error(codes.InvalidArgument, "no date From provided")
	}

	if req.To == nil {
		return nil, status.Error(codes.InvalidArgument, "no date To provided")
	}

	params := domain.GetAvailableSlotsParams{
		AccountID: req.AccountId,
		From:      req.From,
		To:        req.To,
	}

	availableSlotsResponse, err := b.bookingDomain.GetAvailableSlots(ctx, params)
	if err != nil {
		switch {

		case errors.Is(err, gateway.ErrInvalidArgument):
			return &bookingv1.GetAvailableSlotsResponse{
				Slots: nil,
			}, status.Errorf(codes.Internal, "failed to get available slots, %s", err)

		case errors.Is(err, gateway.ErrInternalBadParameters):
			return &bookingv1.GetAvailableSlotsResponse{
				Slots: nil,
			}, status.Errorf(codes.Internal, "failed to get available slots, %s", err)

		case errors.Is(err, gateway.ErrInternal):
			return &bookingv1.GetAvailableSlotsResponse{
				Slots: nil,
			}, status.Errorf(codes.Internal, "failed to get available slots, %s", err)

		case errors.Is(err, gateway.ErrNotFound):
			return &bookingv1.GetAvailableSlotsResponse{
				Slots: nil,
			}, status.Errorf(codes.NotFound, "failed to get available slots, %s", err)

		case errors.Is(err, gateway.ErrOutOfRange):
			return &bookingv1.GetAvailableSlotsResponse{
				Slots: nil,
			}, status.Errorf(codes.OutOfRange, "failed to get available slots, %s", err)

		case errors.Is(err, domain.ErrNoAvailableSlotsForProvidedDates):
			return &bookingv1.GetAvailableSlotsResponse{
				Slots: nil,
			}, status.Errorf(codes.OutOfRange, "failed to get available slots, %s", err)

		default:
			return &bookingv1.GetAvailableSlotsResponse{
				Slots: nil,
			}, status.Errorf(codes.Internal, "failed to get available slots, %s", err)
		}
	}

	bookingSlots := make([]*bookingv1.BookingSlot, len(availableSlotsResponse.Slots))

	for index, slot := range availableSlotsResponse.Slots {

		bookingSlot := bookingv1.BookingSlot{
			Date: &date.Date{
				Year:  int32(slot.Date.Year()),
				Month: int32(slot.Date.Month()),
				Day:   int32(slot.Date.Day()),
			},
			StartTime: int32(slot.StartTime),
			EndTime:   int32(slot.EndTime),
		}

		bookingSlots[index] = &bookingSlot
	}

	return &bookingv1.GetAvailableSlotsResponse{
		Slots: bookingSlots,
	}, nil
}

func (b *BookingAPI) CreateBooking(ctx context.Context, req *bookingv1.CreateBookingRequest) (_ *bookingv1.CreateBookingResponse, err error) { // nolint:revive
	ctx, span := tracing.Tracer().Start(ctx, "BookingAPI.CreateBooking")
	defer func() {
		tracing.RecordSpanError(span, err) // nolint: errcheck
		span.End()
	}()

	err = b.validateCredentials(ctx, auth.CreateAction, auth.AccountBookingResource, req.AccountId)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", err)
		default:
			return nil, status.Error(codes.Internal, "failed to validate credentials")
		}
	}

	if err := validateRequest(req); err != nil {
		return nil, err
	}

	if req.ContactDetails == nil {
		return nil, status.Error(codes.InvalidArgument, "no contact details provided")
	}

	if req.Platform == bookingv1.Platform_PLATFORM_UNKNOWN {
		return nil, status.Error(codes.InvalidArgument, "unknown platform provided")
	}

	if req.Slot == nil {
		return nil, status.Error(codes.InvalidArgument, "no slot provided")
	}

	if req.VulnerabilityDetails == nil {
		return nil, status.Error(codes.InvalidArgument, "no vulnerability details provided")
	}

	params := domain.CreateBookingParams{
		AccountID: req.AccountId,
		ContactDetails: models.AccountDetails{
			Title:     req.GetContactDetails().Title,
			FirstName: req.GetContactDetails().FirstName,
			LastName:  req.GetContactDetails().LastName,
			Email:     req.GetContactDetails().Email,
			Mobile:    req.GetContactDetails().Phone,
		},
		Slot: models.BookingSlot{
			Date:      time.Date(int(req.Slot.Date.Year), time.Month(req.Slot.Date.Month), int(req.Slot.Date.Day), 0, 0, 0, 0, time.UTC),
			StartTime: int(req.Slot.StartTime),
			EndTime:   int(req.Slot.EndTime),
		},
		VulnerabilityDetails: req.VulnerabilityDetails,
		Source:               models.PlatformSourceToBookingSource(req.Platform),
	}

	createBookingResponse, err := b.bookingDomain.CreateBooking(ctx, params)
	if err != nil {
		switch {
		case errors.Is(err, gateway.ErrInvalidArgument):
			return &bookingv1.CreateBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.Internal, "failed to create booking, %s", err)

		case errors.Is(err, gateway.ErrInternalBadParameters):
			return &bookingv1.CreateBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.Internal, "failed to create booking, %s", err)

		case errors.Is(err, gateway.ErrInternal):
			return &bookingv1.CreateBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.Internal, "failed to create booking, %s", err)

		case errors.Is(err, gateway.ErrNotFound):
			return &bookingv1.CreateBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.NotFound, "failed to create booking, %s", err)

		case errors.Is(err, gateway.ErrOutOfRange):
			return &bookingv1.CreateBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.OutOfRange, "failed to create booking, %s", err)

		case errors.Is(err, gateway.ErrAlreadyExists):
			return &bookingv1.CreateBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.AlreadyExists, "failed to create booking, %s", err)

		case errors.Is(err, gateway.ErrInvalidAppointmentDate):
			return &bookingv1.CreateBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.InvalidArgument, "failed to create booking, %s", err)

		case errors.Is(err, gateway.ErrInvalidAppointmentTime):
			return &bookingv1.CreateBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.InvalidArgument, "failed to create booking, %s", err)

		default:
			return &bookingv1.CreateBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.Internal, "failed to create booking, %s", err)

		}
	}

	err = b.publisher.Sink(ctx, createBookingResponse.Event, time.Now())
	if err != nil {
		logrus.Errorf("failed to sink create booking event: %+v", createBookingResponse.Event)
	}

	return &bookingv1.CreateBookingResponse{
		BookingId: createBookingResponse.Event.(*bookingv1.BookingCreatedEvent).BookingId,
	}, nil
}

func (b *BookingAPI) RescheduleBooking(ctx context.Context, req *bookingv1.RescheduleBookingRequest) (_ *bookingv1.RescheduleBookingResponse, err error) { // nolint:revive
	ctx, span := tracing.Tracer().Start(ctx, "BookingAPI.GetCustomerBookings")
	defer func() {
		tracing.RecordSpanError(span, err) // nolint: errcheck
		span.End()
	}()

	err = b.validateCredentials(ctx, auth.UpdateAction, auth.AccountBookingResource, req.AccountId)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", err)
		default:
			return nil, status.Error(codes.Internal, "failed to validate credentials")
		}
	}

	if err := validateRequest(req); err != nil {
		return nil, err
	}

	if req.BookingId == "" {
		return nil, status.Error(codes.InvalidArgument, "no booking id provided")
	}

	if req.Platform == bookingv1.Platform_PLATFORM_UNKNOWN {
		return nil, status.Error(codes.InvalidArgument, "platform unknown")
	}

	if req.Slot == nil {
		return nil, status.Error(codes.InvalidArgument, "no slot was provided")
	}

	params := domain.RescheduleBookingParams{
		AccountID: req.AccountId,
		BookingID: req.BookingId,
		Slot: models.BookingSlot{
			Date:      time.Date(int(req.Slot.Date.Year), time.Month(req.Slot.Date.Month), int(req.Slot.Date.Day), 0, 0, 0, 0, time.UTC),
			StartTime: int(req.Slot.StartTime),
			EndTime:   int(req.Slot.EndTime),
		},
		Source: models.PlatformSourceToBookingSource(req.Platform),
	}

	rescheduleBookingResponse, err := b.bookingDomain.RescheduleBooking(ctx, params)
	if err != nil {
		switch {

		case errors.Is(err, gateway.ErrInvalidArgument):
			return &bookingv1.RescheduleBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.Internal, "failed to reschedule booking, %s", err)

		case errors.Is(err, gateway.ErrInternalBadParameters):
			return &bookingv1.RescheduleBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.Internal, "failed to reschedule booking, %s", err)

		case errors.Is(err, gateway.ErrInternal):
			return &bookingv1.RescheduleBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.Internal, "failed to reschedule booking, %s", err)

		case errors.Is(err, gateway.ErrNotFound):
			return &bookingv1.RescheduleBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.NotFound, "failed to reschedule booking, %s", err)

		case errors.Is(err, gateway.ErrOutOfRange):
			return &bookingv1.RescheduleBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.OutOfRange, "failed to reschedule booking, %s", err)

		case errors.Is(err, gateway.ErrAlreadyExists):
			return &bookingv1.RescheduleBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.AlreadyExists, "failed to reschedule booking, %s", err)

		case errors.Is(err, gateway.ErrInvalidAppointmentDate):
			return &bookingv1.RescheduleBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.InvalidArgument, "failed to reschedule booking, %s", err)

		case errors.Is(err, gateway.ErrInvalidAppointmentTime):
			return &bookingv1.RescheduleBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.InvalidArgument, "failed to reschedule booking, %s", err)

		default:
			return &bookingv1.RescheduleBookingResponse{
				BookingId: "",
			}, status.Errorf(codes.Internal, "failed to reschedule booking, %s", err)
		}
	}

	err = b.publisher.Sink(ctx, rescheduleBookingResponse.Event, time.Now())
	if err != nil {
		logrus.Errorf("failed to sink reschedule booking event: %+v", rescheduleBookingResponse.Event)
	}

	return &bookingv1.RescheduleBookingResponse{
		BookingId: rescheduleBookingResponse.Event.(*bookingv1.BookingRescheduledEvent).BookingId,
	}, nil
}

func validateRequest(req accountIder) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "no request provided")
	}

	if req.GetAccountId() == "" {
		return status.Error(codes.InvalidArgument, "no account id provided")
	}

	return nil
}

func (b *BookingAPI) validateCredentials(ctx context.Context, action, resource, requestAccountID string) error {

	authorised, err := b.auth.Authorize(ctx, &auth.PolicyParams{
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
