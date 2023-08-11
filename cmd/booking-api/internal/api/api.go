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
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
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

type BookingAPI struct {
	bookingDomain BookingDomain
	publisher     BookingPublisher
	bookingv1.UnimplementedBookingAPIServer
}

type accountIder interface {
	GetAccountId() string
}

func New(bookingDomain BookingDomain, publisher BookingPublisher) *BookingAPI {
	return &BookingAPI{
		bookingDomain: bookingDomain,
		publisher:     publisher,
	}
}

var (
	ErrNotImplemented = status.Error(codes.Internal, "not implemented")
)

func (b *BookingAPI) GetCustomerContactDetails(ctx context.Context, req *bookingv1.GetCustomerContactDetailsRequest) (*bookingv1.GetCustomerContactDetailsResponse, error) { // nolint:revive
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

func (b *BookingAPI) GetCustomerSiteAddress(ctx context.Context, req *bookingv1.GetCustomerSiteAddressRequest) (*bookingv1.GetCustomerSiteAddressResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	accountAddress, err := b.bookingDomain.GetAccountAddressByAccountID(ctx, req.GetAccountId())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNoOccupanciesFound) ||
			errors.Is(err, domain.ErrNoEligibleOccupanciesFound):
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

func (b *BookingAPI) GetCustomerBookings(ctx context.Context, req *bookingv1.GetCustomerBookingsRequest) (*bookingv1.GetCustomerBookingsResponse, error) { // nolint:revive

	bookings, err := b.bookingDomain.GetCustomerBookings(ctx, req.GetAccountId())
	if err != nil {
		return nil, err
	}

	return &bookingv1.GetCustomerBookingsResponse{Bookings: bookings}, nil
}

func (b *BookingAPI) GetAvailableSlots(ctx context.Context, req *bookingv1.GetAvailableSlotsRequest) (*bookingv1.GetAvailableSlotsResponse, error) { // nolint:revive

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

		case errors.Is(err, gateway.ErrInvalidAppointmentDate):
			return &bookingv1.GetAvailableSlotsResponse{
				Slots: nil,
			}, status.Errorf(codes.InvalidArgument, "failed to get available slots, %s", err)

		case errors.Is(err, gateway.ErrInvalidAppointmentTime):
			return &bookingv1.GetAvailableSlotsResponse{
				Slots: nil,
			}, status.Errorf(codes.InvalidArgument, "failed to get available slots, %s", err)

		default:
			return nil, status.Errorf(codes.Internal, "failed to get available slots, %s", err)
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

func (b *BookingAPI) CreateBooking(ctx context.Context, req *bookingv1.CreateBookingRequest) (*bookingv1.CreateBookingResponse, error) { // nolint:revive

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

func (b *BookingAPI) RescheduleBooking(ctx context.Context, req *bookingv1.RescheduleBookingRequest) (*bookingv1.RescheduleBookingResponse, error) { // nolint:revive

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
			return nil, status.Errorf(codes.Internal, "failed to reschedule booking, %s", err)
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
