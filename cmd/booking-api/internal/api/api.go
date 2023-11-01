package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/account-platform/pkg/id"
	addressv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/address/v1"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/internal/auth"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/helpers"
	"github.com/utilitywarehouse/uwos-go/v1/telemetry/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

	// POS Journey
	CreateBookingPointOfSale(ctx context.Context, params domain.CreatePOSBookingParams) (domain.CreateBookingResponse, error)
	GetAvailableSlotsPointOfSale(ctx context.Context, params domain.GetPOSAvailableSlotsParams) (domain.GetAvailableSlotsResponse, error)
	GetCustomerDetailsPointOfSale(ctx context.Context, accountNumber string) (*models.PointOfSaleCustomerDetails, error)

	// Process Eligibility
	ProcessEligibility(context.Context, domain.ProcessEligibilityParams) (domain.ProcessEligibilityResult, error)
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
	useTracing bool
}

type accountIder interface {
	GetAccountId() string
}

type accountNumberer interface {
	GetAccountNumber() string
}

func New(bookingDomain BookingDomain, publisher BookingPublisher, auth Auth, useTracing bool) *BookingAPI {
	return &BookingAPI{
		bookingDomain: bookingDomain,
		publisher:     publisher,
		auth:          auth,
		useTracing:    useTracing,
	}
}

func (b *BookingAPI) GetCustomerContactDetails(ctx context.Context, req *bookingv1.GetCustomerContactDetailsRequest) (_ *bookingv1.GetCustomerContactDetailsResponse, err error) {
	var span trace.Span
	if b.useTracing {
		ctx, span = tracing.Tracer().Start(ctx, "BookingAPI.GetCustomerContactDetails",
			trace.WithAttributes(attribute.String("account.id", req.GetAccountId())),
		)
		defer func() {
			tracing.RecordSpanError(span, err)
			span.End()
		}()
	}

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

	contactDetails := &bookingv1.GetCustomerContactDetailsResponse{
		Title:     account.Details.Title,
		FirstName: account.Details.FirstName,
		LastName:  account.Details.LastName,
		Phone:     account.Details.Mobile,
		Email:     account.Details.Email,
	}

	if b.useTracing {
		contactAttr := helpers.CreateSpanAttribute(contactDetails, "contact", span)
		span.AddEvent("response", trace.WithAttributes(contactAttr))
	}

	return contactDetails, nil
}

func (b *BookingAPI) GetCustomerSiteAddress(ctx context.Context, req *bookingv1.GetCustomerSiteAddressRequest) (_ *bookingv1.GetCustomerSiteAddressResponse, err error) {
	var span trace.Span
	if b.useTracing {
		ctx, span = tracing.Tracer().Start(ctx, "BookingAPI.GetCustomerSiteAddress",
			trace.WithAttributes(attribute.String("account.id", req.GetAccountId())),
		)
		defer func() {
			tracing.RecordSpanError(span, err)
			span.End()
		}()
	}

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

	siteAddress := &addressv1.Address{
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
	}

	if b.useTracing {
		addressAttr := helpers.CreateSpanAttribute(siteAddress, "address", span)
		span.AddEvent("response", trace.WithAttributes(addressAttr))
	}

	return &bookingv1.GetCustomerSiteAddressResponse{
		SiteAddress: siteAddress,
	}, nil
}

func (b *BookingAPI) GetCustomerBookings(ctx context.Context, req *bookingv1.GetCustomerBookingsRequest) (_ *bookingv1.GetCustomerBookingsResponse, err error) {
	var span trace.Span
	if b.useTracing {
		ctx, span = tracing.Tracer().Start(ctx, "BookingAPI.GetCustomerBookings",
			trace.WithAttributes(attribute.String("account.id", req.GetAccountId())),
		)
		defer func() {
			tracing.RecordSpanError(span, err)
			span.End()
		}()
	}

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

	if b.useTracing {
		bookingAttr := helpers.CreateSpanAttribute(bookings, "bookings", span)
		span.AddEvent("response", trace.WithAttributes(bookingAttr))
	}

	return &bookingv1.GetCustomerBookingsResponse{Bookings: bookings}, nil
}

func (b *BookingAPI) GetAvailableSlots(ctx context.Context, req *bookingv1.GetAvailableSlotsRequest) (_ *bookingv1.GetAvailableSlotsResponse, err error) {
	var span trace.Span
	if b.useTracing {
		ctx, span = tracing.Tracer().Start(ctx, "BookingAPI.GetAvailableSlots",
			trace.WithAttributes(attribute.String("account.id", req.GetAccountId())),
		)
		span.AddEvent("request", trace.WithAttributes(attribute.String("from", req.GetFrom().String()), attribute.String("to", req.GetTo().String())))
		defer func() {
			tracing.RecordSpanError(span, err)
			span.End()
		}()
	}

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
		return &bookingv1.GetAvailableSlotsResponse{
			Slots: nil,
		}, mapError("failed to get available slots, %s", err)
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

	if b.useTracing {
		bookingSlotsAttr := helpers.CreateSpanAttribute(bookingSlots, "bookingSlots", span)
		span.AddEvent("response", trace.WithAttributes(bookingSlotsAttr))
	}

	return &bookingv1.GetAvailableSlotsResponse{
		Slots: bookingSlots,
	}, nil
}

func (b *BookingAPI) CreateBooking(ctx context.Context, req *bookingv1.CreateBookingRequest) (_ *bookingv1.CreateBookingResponse, err error) {
	if b.useTracing {
		var span trace.Span
		ctx, span = tracing.Tracer().Start(ctx, "BookingAPI.CreateBooking",
			trace.WithAttributes(attribute.String("account.id", req.GetAccountId())),
		)
		defer func() {
			tracing.RecordSpanError(span, err)
			span.End()
		}()
	}

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
		return &bookingv1.CreateBookingResponse{
			BookingId: "",
		}, mapError("failed to create booking, %s", err)
	}

	err = b.publisher.Sink(ctx, createBookingResponse.Event, time.Now())
	if err != nil {
		logrus.Errorf("failed to sink create booking event: %+v", createBookingResponse.Event)
	}

	return &bookingv1.CreateBookingResponse{
		BookingId: createBookingResponse.Event.(*bookingv1.BookingCreatedEvent).BookingId,
	}, nil
}

func (b *BookingAPI) RescheduleBooking(ctx context.Context, req *bookingv1.RescheduleBookingRequest) (_ *bookingv1.RescheduleBookingResponse, err error) {
	if b.useTracing {
		var span trace.Span
		ctx, span = tracing.Tracer().Start(ctx, "BookingAPI.RescheduleBooking")
		defer func() {
			tracing.RecordSpanError(span, err)
			span.End()
		}()
	}

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

	if req.ContactDetails == nil {
		return nil, status.Error(codes.InvalidArgument, "no contact details provided")
	}

	if req.VulnerabilityDetails == nil {
		return nil, status.Error(codes.InvalidArgument, "no vulnerability details provided")
	}

	params := domain.RescheduleBookingParams{
		AccountID:            req.AccountId,
		BookingID:            req.BookingId,
		VulnerabilityDetails: req.VulnerabilityDetails,
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
		Source: models.PlatformSourceToBookingSource(req.Platform),
	}

	rescheduleBookingResponse, err := b.bookingDomain.RescheduleBooking(ctx, params)
	if err != nil {
		return &bookingv1.RescheduleBookingResponse{
			BookingId: "",
		}, mapError("failed to reschedule booking, %s", err)
	}

	err = b.publisher.Sink(ctx, rescheduleBookingResponse.Event, time.Now())
	if err != nil {
		logrus.Errorf("failed to sink reschedule booking event: %+v", rescheduleBookingResponse.Event)
	}

	return &bookingv1.RescheduleBookingResponse{
		BookingId: rescheduleBookingResponse.Event.(*bookingv1.BookingRescheduledEvent).BookingId,
	}, nil
}

func (b *BookingAPI) GetAvailableSlotsPointOfSale(ctx context.Context, req *bookingv1.GetAvailableSlotsPointOfSaleRequest) (_ *bookingv1.GetAvailableSlotsPointOfSaleResponse, err error) {
	if err := validatePOSRequest(req); err != nil {
		return nil, err
	}
	accountID := id.NewAccountID(req.GetAccountNumber())

	var span trace.Span
	if b.useTracing {
		ctx, span = tracing.Tracer().Start(ctx, "BookingAPI.GetAvailableSlots",
			trace.WithAttributes(attribute.String("account.id", accountID)),
		)
		span.AddEvent("request", trace.WithAttributes(attribute.String("from", req.GetFrom().String()), attribute.String("to", req.GetTo().String())))
		defer func() {
			tracing.RecordSpanError(span, err)
			span.End()
		}()
	}

	err = b.validateCredentials(ctx, auth.GetAction, auth.AccountBookingResource, accountID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", err)
		default:
			return nil, status.Error(codes.Internal, "failed to validate credentials")
		}
	}

	if req.From == nil {
		return nil, status.Error(codes.InvalidArgument, "no date From provided")
	}

	if req.To == nil {
		return nil, status.Error(codes.InvalidArgument, "no date To provided")
	}

	params := domain.GetPOSAvailableSlotsParams{
		AccountID: accountID,
		From:      req.From,
		To:        req.To,
	}

	availableSlotsResponse, err := b.bookingDomain.GetAvailableSlotsPointOfSale(ctx, params)
	if err != nil {
		return &bookingv1.GetAvailableSlotsPointOfSaleResponse{
			Slots: nil,
		}, mapError("failed to get available slots, %s", err)
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

	if b.useTracing {
		bookingSlotsAttr := helpers.CreateSpanAttribute(bookingSlots, "bookingSlots", span)
		span.AddEvent("response", trace.WithAttributes(bookingSlotsAttr))
	}

	return &bookingv1.GetAvailableSlotsPointOfSaleResponse{
		Slots: bookingSlots,
	}, nil
}

func (b *BookingAPI) CreateBookingPointOfSale(ctx context.Context, req *bookingv1.CreateBookingPointOfSaleRequest) (_ *bookingv1.CreateBookingPointOfSaleResponse, err error) {
	if err := validatePOSRequest(req); err != nil {
		return nil, err
	}
	accountID := id.NewAccountID(req.GetAccountNumber())

	if b.useTracing {
		var span trace.Span
		ctx, span = tracing.Tracer().Start(ctx, "BookingAPI.CreatePOSBooking",
			trace.WithAttributes(attribute.String("account.id", accountID)),
		)
		defer func() {
			tracing.RecordSpanError(span, err)
			span.End()
		}()
	}

	err = b.validateCredentials(ctx, auth.CreateAction, auth.AccountBookingResource, accountID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", err)
		default:
			return nil, status.Error(codes.Internal, "failed to validate credentials")
		}
	}

	if req.SiteAddress == nil {
		return nil, status.Error(codes.InvalidArgument, "no site address provided")
	}

	if req.SiteAddress.Paf == nil {
		return nil, status.Error(codes.InvalidArgument, "no Postcode Address File(PAF) provided")
	}

	if req.SiteAddress.Paf.Postcode == "" {
		return nil, status.Error(codes.InvalidArgument, "no post code provided")
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

	params := domain.CreatePOSBookingParams{
		AccountID: accountID,
		SiteAddress: models.AccountAddress{
			UPRN: req.SiteAddress.Uprn,
			PAF: models.PAF{
				BuildingName:            req.SiteAddress.Paf.BuildingName,
				BuildingNumber:          req.SiteAddress.Paf.BuildingNumber,
				Department:              req.SiteAddress.Paf.Department,
				DependentLocality:       req.SiteAddress.Paf.DependentLocality,
				DependentThoroughfare:   req.SiteAddress.Paf.DependentThoroughfare,
				DoubleDependentLocality: req.SiteAddress.Paf.DoubleDependentLocality,
				Organisation:            req.SiteAddress.Paf.Organisation,
				PostTown:                req.SiteAddress.Paf.PostTown,
				Postcode:                req.SiteAddress.Paf.Postcode,
				SubBuilding:             req.SiteAddress.Paf.SubBuilding,
				Thoroughfare:            req.SiteAddress.Paf.Thoroughfare,
			},
		},
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

	createBookingResponse, err := b.bookingDomain.CreateBookingPointOfSale(ctx, params)
	if err != nil {
		switch err {
		case domain.ErrMissingOccupancyInBooking:
			logrus.Warnf("occupancy for the account_id: %s was not found! saving the partial booking created event in the database with booking_id: %s", createBookingResponse.Event.(*bookingv1.BookingCreatedEvent).Details.AccountId, createBookingResponse.Event.(*bookingv1.BookingCreatedEvent).BookingId)
			return &bookingv1.CreateBookingPointOfSaleResponse{
				BookingId: createBookingResponse.Event.(*bookingv1.BookingCreatedEvent).BookingId,
			}, nil
		default:
			return &bookingv1.CreateBookingPointOfSaleResponse{
				BookingId: "",
			}, mapError("failed to create booking, %s", err)
		}
	}

	err = b.publisher.Sink(ctx, createBookingResponse.Event, time.Now())
	if err != nil {
		logrus.Errorf("failed to sink create booking event: %+v, %s", createBookingResponse.Event, err)
	}

	return &bookingv1.CreateBookingPointOfSaleResponse{
		BookingId: createBookingResponse.Event.(*bookingv1.BookingCreatedEvent).BookingId,
	}, nil
}

func (b *BookingAPI) GetCustomerDetailsPointOfSale(ctx context.Context, req *bookingv1.GetCustomerDetailsPointOfSaleRequest) (_ *bookingv1.GetCustomerDetailsPointOfSaleResponse, err error) {

	if req.AccountNumber == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid account number provided (empty string)")
	}

	if b.useTracing {
		var span trace.Span
		ctx, span = tracing.Tracer().Start(ctx, "BookingAPI.GetCustomerDetailsPointOfSale",
			trace.WithAttributes(attribute.String("account.number", req.AccountNumber)),
		)
		defer func() {
			tracing.RecordSpanError(span, err)
			span.End()
		}()
	}

	err = b.validateCredentials(ctx, auth.GetAction, auth.AccountBookingResource, req.AccountNumber)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", err)
		default:
			return nil, status.Error(codes.Internal, "failed to validate credentials")
		}
	}

	customerDetails, err := b.bookingDomain.GetCustomerDetailsPointOfSale(ctx, req.AccountNumber)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPointOfSaleCustomerDetailsNotFound):
			return nil, status.Errorf(codes.NotFound, "did not find customer details for provided account number: %s", req.GetAccountNumber())
		}
		return nil, status.Errorf(codes.Internal, "failed to get customer details point of sale, %s", err)
	}

	return &bookingv1.GetCustomerDetailsPointOfSaleResponse{
		ContactDetails: &bookingv1.ContactDetails{
			Title:     customerDetails.Details.Title,
			FirstName: customerDetails.Details.FirstName,
			LastName:  customerDetails.Details.LastName,
			Phone:     customerDetails.Details.Mobile,
			Email:     customerDetails.Details.Email,
		},
		SiteAddress: &addressv1.Address{
			Uprn: customerDetails.Address.UPRN,
			Paf: &addressv1.Address_PAF{
				Organisation:            customerDetails.Address.PAF.Organisation,
				Department:              customerDetails.Address.PAF.Department,
				SubBuilding:             customerDetails.Address.PAF.SubBuilding,
				BuildingName:            customerDetails.Address.PAF.BuildingName,
				BuildingNumber:          customerDetails.Address.PAF.BuildingNumber,
				DependentThoroughfare:   customerDetails.Address.PAF.DependentThoroughfare,
				Thoroughfare:            customerDetails.Address.PAF.Thoroughfare,
				DoubleDependentLocality: customerDetails.Address.PAF.DoubleDependentLocality,
				DependentLocality:       customerDetails.Address.PAF.DependentLocality,
				PostTown:                customerDetails.Address.PAF.PostTown,
				Postcode:                customerDetails.Address.PAF.Postcode,
			},
		},
	}, nil
}

// TODO: ADD test
func (b *BookingAPI) GetEligibilityPointOfSaleJourney(ctx context.Context, req *bookingv1.GetEligibilityPointOfSaleJourneyRequest) (_ *bookingv1.GetEligibilityPointOfSaleJourneyResponse, err error) {

	if req.ContactDetails == nil {
		return nil, status.Error(codes.InvalidArgument, "provided contact details is missing")
	}

	if req.SiteAddress == nil {
		return nil, status.Error(codes.InvalidArgument, "provided site address is missing")
	}

	if req.SiteAddress.Paf == nil {
		return nil, status.Error(codes.InvalidArgument, "provided PAF is missing")
	}

	if req.SiteAddress.Paf.Postcode == "" {
		return nil, status.Error(codes.InvalidArgument, "provided post code is missing")
	}

	if req.AccountNumber == "" {
		return nil, status.Error(codes.InvalidArgument, "provided account number is missing")
	}

	if req.Mpan == "" {
		return nil, status.Error(codes.InvalidArgument, "provided mpan is missing")
	}

	if req.ElectricityTariffType == bookingv1.TariffType_TARIFF_TYPE_UNKNOWN {
		return nil, status.Error(codes.InvalidArgument, "provided electricity type is missing")
	}

	if req.Mprn != "" && req.GasTariffType == bookingv1.TariffType_TARIFF_TYPE_UNKNOWN {
		return nil, status.Error(codes.InvalidArgument, "provided mprn is not empty, but gas tariff type is unknown")
	}

	err = b.validateCredentials(ctx, auth.GetAction, auth.EligibilityResource, req.AccountNumber)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.Unauthenticated, "user does not have access to this action, %s", err)
		default:
			return nil, status.Error(codes.Internal, "failed to validate credentials")
		}
	}

	result, err := b.bookingDomain.ProcessEligibility(ctx, domain.ProcessEligibilityParams{
		AccountNumber: req.AccountNumber,
		Details: models.PointOfSaleCustomerDetails{
			AccountNumber: req.AccountNumber,
			Details: models.AccountDetails{
				Title:     req.ContactDetails.Title,
				FirstName: req.ContactDetails.FirstName,
				LastName:  req.ContactDetails.LastName,
				Email:     req.ContactDetails.Email,
				Mobile:    req.ContactDetails.Phone,
			},
			Address: models.AccountAddress{
				UPRN: req.SiteAddress.Uprn,
				PAF: models.PAF{
					BuildingName:            req.SiteAddress.Paf.BuildingName,
					BuildingNumber:          req.SiteAddress.Paf.BuildingNumber,
					Department:              req.SiteAddress.Paf.Department,
					DependentLocality:       req.SiteAddress.Paf.DependentLocality,
					DependentThoroughfare:   req.SiteAddress.Paf.DependentThoroughfare,
					DoubleDependentLocality: req.SiteAddress.Paf.DoubleDependentLocality,
					Organisation:            req.SiteAddress.Paf.Organisation,
					PostTown:                req.SiteAddress.Paf.PostTown,
					Postcode:                req.SiteAddress.Paf.Postcode,
					SubBuilding:             req.SiteAddress.Paf.SubBuilding,
					Thoroughfare:            req.SiteAddress.Paf.Thoroughfare,
				},
			},
			Meterpoints: []models.Meterpoint{
				{
					MPXN:       req.Mpan,
					TariffType: req.ElectricityTariffType,
				},
				{
					MPXN:       req.Mprn,
					TariffType: req.GasTariffType,
				},
			},
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to process eligibility for meterpoint, %s", err)
	}

	return &bookingv1.GetEligibilityPointOfSaleJourneyResponse{
		Eligible: result.Eligible,
		Link:     result.Link,
	}, nil
}

func validatePOSRequest(req accountNumberer) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "no request provided")
	}

	if req.GetAccountNumber() == "" {
		return status.Error(codes.InvalidArgument, "no account number provided")
	}

	return nil
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

func mapError(message string, err error) error {
	switch {
	case errors.Is(err, gateway.ErrInvalidArgument):
		return status.Errorf(codes.Internal, message, err)

	case errors.Is(err, gateway.ErrInternalBadParameters):
		return status.Errorf(codes.Internal, message, err)

	case errors.Is(err, gateway.ErrInternal):
		return status.Errorf(codes.Internal, message, err)

	case errors.Is(err, gateway.ErrNotFound):
		return status.Errorf(codes.NotFound, message, err)

	case errors.Is(err, gateway.ErrOutOfRange):
		return status.Errorf(codes.OutOfRange, message, err)

	case errors.Is(err, domain.ErrNoAvailableSlotsForProvidedDates):
		return status.Errorf(codes.OutOfRange, message, err)

	case errors.Is(err, gateway.ErrAlreadyExists):
		return status.Errorf(codes.AlreadyExists, message, err)

	case errors.Is(err, gateway.ErrInvalidAppointmentDate):
		return status.Errorf(codes.InvalidArgument, message, err)

	case errors.Is(err, gateway.ErrInvalidAppointmentTime):
		return status.Errorf(codes.InvalidArgument, message, err)

	default:
		return status.Errorf(codes.Internal, message, err)
	}
}
