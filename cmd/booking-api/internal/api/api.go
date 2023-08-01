package api

import (
	"context"
	"errors"
	"fmt"

	addressv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/address/v1"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type BookingDomain interface {
	GetCustomerContactDetails(ctx context.Context, accountID string) (models.Account, error)
	GetAccountAddressByAccountID(ctx context.Context, accountID string) (models.AccountAddress, error)
	GetCustomerBookings(ctx context.Context, accountID string) ([]*bookingv1.Booking, error)
	CreateBooking(ctx context.Context, params domain.CreateBookingParams) (proto.Message, error)
	GetAvailableSlots(ctx context.Context, params domain.GetAvailableSlotsParams) (domain.GetAvailableSlotsResponse, error)
}

type BookingAPI struct {
	bookingDomain BookingDomain
	bookingv1.UnimplementedBookingAPIServer
}

type accountIder interface {
	GetAccountId() string
}

func New(bookingDomain BookingDomain) *BookingAPI {
	return &BookingAPI{
		bookingDomain: bookingDomain,
	}
}

var (
	ErrNotImplemented = status.Error(codes.Internal, "not implemented")
)

func validateRequest(req accountIder) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "no request provided")
	}

	if req.GetAccountId() == "" {
		return status.Error(codes.InvalidArgument, "no account id provided")
	}

	return nil
}

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
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	bookings, err := b.bookingDomain.GetCustomerBookings(ctx, req.GetAccountId())
	if err != nil {
		return nil, err
	}

	return &bookingv1.GetCustomerBookingsResponse{Bookings: bookings}, nil
}

func (b *BookingAPI) GetAvailableSlots(ctx context.Context, req *bookingv1.GetAvailableSlotsRequest) (*bookingv1.GetAvailableSlotsResponse, error) { // nolint:revive

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "no request provided")
	}

	if req.GetAccountId() == "" {
		return nil, status.Error(codes.InvalidArgument, "no account ID provided")
	}

	if req.From == nil {
		return nil, status.Error(codes.InvalidArgument, "no date From provided")
	}

	if req.To == nil {
		return nil, status.Error(codes.InvalidArgument, "no date To provided")
	}

	params := domain.GetAvailableSlotsParams{
		AccountID: req.AccountId,
		From:      *req.From,
		To:        *req.To,
	}

	availableSlots, err := b.bookingDomain.GetAvailableSlots(ctx, params)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get available slots, %s", err))
	}

	bookingSlots := make([]*bookingv1.BookingSlot, len(availableSlots.Slots))

	for _, slot := range availableSlots.Slots {
		bookingSlot := bookingv1.BookingSlot{
			Date:      &slot.Date,
			StartTime: slot.StartTime,
			EndTime:   slot.EndTime,
		}

		bookingSlots = append(bookingSlots, &bookingSlot)
	}

	return &bookingv1.GetAvailableSlotsResponse{
		Slots: bookingSlots,
	}, nil
}

func (b *BookingAPI) CreateBooking(ctx context.Context, req *bookingv1.CreateBookingRequest) (*bookingv1.CreateBookingResponse, error) { // nolint:revive

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "no request provided")
	}

	if req.GetAccountId() == "" {
		return nil, status.Error(codes.InvalidArgument, "no account ID provided")
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
		Slot: models.Slot{
			Date:      *req.Slot.Date,
			StartTime: req.Slot.StartTime,
			EndTime:   req.Slot.EndTime,
		},
		VulnerabilityDetails: req.VulnerabilityDetails,
	}

	createBookingEvent, err := b.bookingDomain.CreateBooking(ctx, params)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create booking, %s", err))
	}

	return &bookingv1.CreateBookingResponse{
		BookingId: createBookingEvent.(*bookingv1.BookingCreatedEvent).BookingId,
	}, nil
}

func (b *BookingAPI) RescheduleBooking(ctx context.Context, req *bookingv1.RescheduleBookingRequest) (*bookingv1.RescheduleBookingResponse, error) { // nolint:revive

	return nil, ErrNotImplemented
}
