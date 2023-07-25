package api

import (
	"context"
	"fmt"

	addressv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/address/v1"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CustomerDomain interface {
	GetCustomerContactDetails(ctx context.Context, accountID string) (models.Account, error)
	GetAccountAddressByAccountID(ctx context.Context, accountID string) (models.AccountAddress, error)
}

type BookingAPI struct {
	customerDomain CustomerDomain
	bookingv1.UnimplementedBookingAPIServer
}

func New(customerDomain CustomerDomain) *BookingAPI {
	return &BookingAPI{
		customerDomain: customerDomain,
	}
}

var (
	ErrNotImplemented = status.Error(codes.Internal, "not implemented")
)

func (b *BookingAPI) GetCustomerContactDetails(ctx context.Context, req *bookingv1.GetCustomerContactDetailsRequest) (*bookingv1.GetCustomerContactDetailsResponse, error) { // nolint:revive

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "no request provided")
	}

	if req.GetAccountId() == "" {
		return nil, status.Error(codes.InvalidArgument, "no account id provided")
	}

	account, err := b.customerDomain.GetCustomerContactDetails(ctx, req.GetAccountId())
	if err != nil {
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
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "no request provided")
	}

	if req.GetAccountId() == "" {
		return nil, status.Error(codes.InvalidArgument, "no account id provided")
	}

	accountAddress, err := b.customerDomain.GetAccountAddressByAccountID(ctx, req.GetAccountId())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get customer contact details, %s", err))
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
	return nil, ErrNotImplemented
}

func (b *BookingAPI) GetAvailableSlots(ctx context.Context, req *bookingv1.GetAvailableSlotsRequest) (*bookingv1.GetAvailableSlotsResponse, error) { // nolint:revive
	return nil, ErrNotImplemented
}

func (b *BookingAPI) CreateBooking(ctx context.Context, req *bookingv1.CreateBookingRequest) (*bookingv1.CreateBookingResponse, error) { // nolint:revive
	return nil, ErrNotImplemented
}

func (b *BookingAPI) RescheduleBooking(ctx context.Context, req *bookingv1.RescheduleBookingRequest) (*bookingv1.RescheduleBookingResponse, error) { // nolint:revive
	return nil, ErrNotImplemented
}
