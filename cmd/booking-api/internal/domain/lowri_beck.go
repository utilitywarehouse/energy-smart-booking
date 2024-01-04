package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	addressv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/address/v1"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	commsv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/comms/v1"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/helpers"
	"github.com/utilitywarehouse/uwos-go/v1/telemetry/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/protobuf/proto"
)

var (
	ErrNoAvailableSlotsForProvidedDates = errors.New("no available slots for provided dates")
	ErrMissingOccupancyInBooking        = errors.New("no occupancy id was found, can not publish create booking event")
	ErrUnsucessfulBooking               = errors.New("create booking point of sale did not return success")
	ErrUnsuccessfulReschedule           = errors.New("reschedule booking did not return success")
)

type GetAvailableSlotsParams struct {
	AccountID string
	From      *date.Date
	To        *date.Date
}

type CreateBookingParams struct {
	AccountID            string
	ContactDetails       models.AccountDetails
	Slot                 models.BookingSlot
	Source               bookingv1.BookingSource
	VulnerabilityDetails *bookingv1.VulnerabilityDetails
}

type RescheduleBookingParams struct {
	AccountID            string
	BookingID            string
	Source               bookingv1.BookingSource
	VulnerabilityDetails *bookingv1.VulnerabilityDetails
	ContactDetails       models.AccountDetails
	Slot                 models.BookingSlot
}

type GetPOSAvailableSlotsParams struct {
	AccountNumber string
	From          *date.Date
	To            *date.Date
}

type CreatePOSBookingParams struct {
	AccountNumber        string
	AccountID            string
	ContactDetails       models.AccountDetails
	Slot                 models.BookingSlot
	Source               bookingv1.BookingSource
	VulnerabilityDetails *bookingv1.VulnerabilityDetails
}

type ReschedulePOSBookingParams struct {
	AccountID         string
	Postcode          string
	Mpan              string
	Mprn              string
	TariffElectricity bookingv1.TariffType
	TariffGas         bookingv1.TariffType
	BookingID         string
	Source            bookingv1.BookingSource
	Slot              models.BookingSlot
}

type GetAvailableSlotsResponse struct {
	Slots []models.BookingSlot
}

type CreateBookingResponse struct {
	Event proto.Message
}

type CreateBookingPointOfSaleResponse struct {
	CommsEvent   proto.Message
	BookingEvent proto.Message
}

type RescheduleBookingResponse struct {
	CommsEvent   proto.Message
	BookingEvent proto.Message
}

func (d BookingDomain) GetAvailableSlots(ctx context.Context, params GetAvailableSlotsParams) (GetAvailableSlotsResponse, error) {
	fromAsTime := time.Date(int(params.From.Year), time.Month(params.From.Month), int(params.From.Day), 0, 0, 0, 0, time.UTC)
	toAsTime := time.Date(int(params.To.Year), time.Month(params.To.Month), int(params.To.Day), 0, 0, 0, 0, time.UTC)

	site, occupancyEligibility, err := d.findLowriBeckKeys(ctx, params.AccountID)
	if err != nil {
		return GetAvailableSlotsResponse{}, fmt.Errorf("failed to find postcode and booking reference, %w", err)
	}

	slotsResponse, err := d.lowribeckGw.GetAvailableSlots(ctx, site.Postcode, occupancyEligibility.Reference)
	if err != nil {
		return GetAvailableSlotsResponse{}, fmt.Errorf("failed to get available slots, %w", err)
	}

	targetedSlots := []models.BookingSlot{}

	for _, elem := range slotsResponse.BookingSlots {
		currentSlotTime := time.Date(elem.Date.Year(), elem.Date.Month(), elem.Date.Day(), 0, 0, 0, 0, time.UTC)

		if currentSlotTime.After(fromAsTime) && currentSlotTime.Before(toAsTime) {
			targetedSlots = append(targetedSlots, elem)
		}
	}

	if len(targetedSlots) == 0 {
		return GetAvailableSlotsResponse{
			Slots: targetedSlots,
		}, ErrNoAvailableSlotsForProvidedDates
	}

	return GetAvailableSlotsResponse{
		Slots: targetedSlots,
	}, nil

}

func (d BookingDomain) CreateBooking(ctx context.Context, params CreateBookingParams) (CreateBookingResponse, error) {

	var event *bookingv1.BookingCreatedEvent

	lbVulnerabilities := mapLowribeckVulnerabilities(params.VulnerabilityDetails.GetVulnerabilities())

	site, occupancyEligibility, err := d.findLowriBeckKeys(ctx, params.AccountID)
	if err != nil {
		return CreateBookingResponse{}, err
	}

	response, err := d.lowribeckGw.CreateBooking(ctx, site.Postcode, occupancyEligibility.Reference, params.Slot, params.ContactDetails, lbVulnerabilities, params.VulnerabilityDetails.Other)
	if err != nil {
		return CreateBookingResponse{}, fmt.Errorf("failed to create booking, %w", err)
	}

	if !response.Success {
		return CreateBookingResponse{}, ErrUnsuccessfulBooking
	}

	bookingID := uuid.New().String()

	event = &bookingv1.BookingCreatedEvent{
		BookingId: bookingID,
		Details: &bookingv1.Booking{
			Id:        bookingID,
			AccountId: params.AccountID,
			SiteAddress: &addressv1.Address{
				Uprn: site.UPRN,
				Paf: &addressv1.Address_PAF{
					Organisation:            site.Organisation,
					Department:              site.Department,
					SubBuilding:             site.SubBuildingNameNumber,
					BuildingName:            site.BuildingNameNumber,
					BuildingNumber:          site.BuildingNameNumber,
					DependentThoroughfare:   site.DependentThoroughfare,
					Thoroughfare:            site.Thoroughfare,
					DoubleDependentLocality: site.DoubleDependentLocality,
					DependentLocality:       site.DependentLocality,
					PostTown:                site.Town,
					Postcode:                site.Postcode,
				},
			},
			ContactDetails: &bookingv1.ContactDetails{
				Title:     params.ContactDetails.Title,
				FirstName: params.ContactDetails.FirstName,
				LastName:  params.ContactDetails.LastName,
				Phone:     params.ContactDetails.Mobile,
				Email:     params.ContactDetails.Email,
			},
			Slot: &bookingv1.BookingSlot{
				Date: &date.Date{
					Year:  int32(params.Slot.Date.Year()),
					Month: int32(params.Slot.Date.Month()),
					Day:   int32(params.Slot.Date.Day()),
				},
				StartTime: int32(params.Slot.StartTime),
				EndTime:   int32(params.Slot.EndTime),
			},
			VulnerabilityDetails: params.VulnerabilityDetails,
			Status:               bookingv1.BookingStatus_BOOKING_STATUS_SCHEDULED,
			ExternalReference:    occupancyEligibility.Reference,
			BookingType:          bookingv1.BookingType_BOOKING_TYPE_SMART_BOOKING_JOURNEY,
		},
		OccupancyId:   occupancyEligibility.OccupancyID,
		BookingSource: params.Source,
	}

	return CreateBookingResponse{
		Event: event,
	}, nil
}

func (d BookingDomain) RescheduleBooking(ctx context.Context, params RescheduleBookingParams) (RescheduleBookingResponse, error) {

	var event *bookingv1.BookingRescheduledEvent
	var commsEvent *commsv1.BookingRescheduledCommsEvent

	site, occupancyEligibility, err := d.findLowriBeckKeys(ctx, params.AccountID)
	if err != nil {
		return RescheduleBookingResponse{}, err
	}

	booking, err := d.bookingStore.GetBookingByBookingID(ctx, params.BookingID)
	if err != nil {
		return RescheduleBookingResponse{}, fmt.Errorf("failed to reschedule booking, %w", err)
	}

	lbVulnerabilities := mapLowribeckVulnerabilities(params.VulnerabilityDetails.Vulnerabilities)

	response, err := d.lowribeckGw.CreateBooking(ctx, site.Postcode, occupancyEligibility.Reference, params.Slot, params.ContactDetails, lbVulnerabilities, params.VulnerabilityDetails.Other)
	if err != nil {
		return RescheduleBookingResponse{}, fmt.Errorf("failed to reschedule booking, %w", err)
	}

	if !response.Success {
		return RescheduleBookingResponse{}, ErrUnsuccessfulReschedule
	}

	event = &bookingv1.BookingRescheduledEvent{
		BookingId:            params.BookingID,
		VulnerabilityDetails: params.VulnerabilityDetails,
		ContactDetails: &bookingv1.ContactDetails{
			Title:     params.ContactDetails.Title,
			FirstName: params.ContactDetails.FirstName,
			LastName:  params.ContactDetails.LastName,
			Phone:     params.ContactDetails.Mobile,
			Email:     params.ContactDetails.Email,
		},
		Slot: &bookingv1.BookingSlot{
			Date: &date.Date{
				Year:  int32(params.Slot.Date.Year()),
				Month: int32(params.Slot.Date.Month()),
				Day:   int32(params.Slot.Date.Day()),
			},
			StartTime: int32(params.Slot.StartTime),
			EndTime:   int32(params.Slot.EndTime),
		},
		Status:        bookingv1.BookingStatus_BOOKING_STATUS_SCHEDULED,
		BookingSource: params.Source,
	}

	// this is a temporary change, at this moment we only want to send comms for point of sale journey bookings
	if booking.BookingType == bookingv1.BookingType_BOOKING_TYPE_POINT_OF_SALE_JOURNEY {

		accountNumber, err := d.accountNumber.Get(ctx, params.AccountID)
		if err != nil {
			return RescheduleBookingResponse{}, fmt.Errorf("failed to reschedule booking, %w", err)
		}

		accAddress := models.AccountAddress{
			UPRN: site.UPRN,
			PAF: models.PAF{
				BuildingName:            site.BuildingNameNumber,
				BuildingNumber:          site.BuildingNameNumber,
				Department:              site.Department,
				DependentLocality:       site.DependentLocality,
				DependentThoroughfare:   site.DependentThoroughfare,
				DoubleDependentLocality: site.DoubleDependentLocality,
				Organisation:            site.Organisation,
				PostTown:                site.Town,
				Postcode:                site.Postcode,
				SubBuilding:             site.SubBuildingNameNumber,
				Thoroughfare:            site.Thoroughfare,
			},
		}

		commsEvent = buildRescheduleCommsEvent(params, booking.Contact, accAddress, accountNumber)
		return RescheduleBookingResponse{
			BookingEvent: event,
			CommsEvent:   commsEvent,
		}, nil
	}

	return RescheduleBookingResponse{
		BookingEvent: event,
		CommsEvent:   nil,
	}, nil
}

func (d BookingDomain) GetAvailableSlotsPointOfSale(ctx context.Context, params GetPOSAvailableSlotsParams) (GetAvailableSlotsResponse, error) {
	fromAsTime := time.Date(int(params.From.Year), time.Month(params.From.Month), int(params.From.Day), 0, 0, 0, 0, time.UTC)
	toAsTime := time.Date(int(params.To.Year), time.Month(params.To.Month), int(params.To.Day), 0, 0, 0, 0, time.UTC)

	customerAccountDetails, err := d.getCustomerDetailsPointOfSale(ctx, params.AccountNumber)
	if err != nil {
		return GetAvailableSlotsResponse{}, fmt.Errorf("failed getting available slots, %w", err)
	}

	slotsResponse, err := d.lowribeckGw.GetAvailableSlotsPointOfSale(
		ctx,
		customerAccountDetails.Address.PAF.Postcode,
		customerAccountDetails.ElecOrderSupplies.MPXN,
		customerAccountDetails.GasOrderSupplies.MPXN,
		models.BookingTariffTypeToLowribeckTariffType(customerAccountDetails.ElecOrderSupplies.TariffType),
		models.BookingTariffTypeToLowribeckTariffType(customerAccountDetails.GasOrderSupplies.TariffType),
	)
	if err != nil {
		return GetAvailableSlotsResponse{}, fmt.Errorf("failed to get POS available slots, %w", err)
	}

	targetedSlots := []models.BookingSlot{}

	for _, elem := range slotsResponse.BookingSlots {
		currentSlotTime := time.Date(elem.Date.Year(), elem.Date.Month(), elem.Date.Day(), 0, 0, 0, 0, time.UTC)

		if currentSlotTime.After(fromAsTime) && currentSlotTime.Before(toAsTime) {
			targetedSlots = append(targetedSlots, elem)
		}
	}

	if len(targetedSlots) == 0 {
		return GetAvailableSlotsResponse{
			Slots: targetedSlots,
		}, ErrNoAvailableSlotsForProvidedDates
	}

	return GetAvailableSlotsResponse{
		Slots: targetedSlots,
	}, nil

}

func (d BookingDomain) CreateBookingPointOfSale(ctx context.Context, params CreatePOSBookingParams) (CreateBookingPointOfSaleResponse, error) {

	bookingID := uuid.New().String()
	var bookingEvent *bookingv1.BookingCreatedEvent
	var commsEvent *commsv1.PointOfSaleBookingConfirmationCommsEvent

	lbVulnerabilities := mapLowribeckVulnerabilities(params.VulnerabilityDetails.GetVulnerabilities())

	accountHolderDetails, err := d.getCustomerDetailsPointOfSale(ctx, params.AccountNumber)
	if err != nil {
		return CreateBookingPointOfSaleResponse{}, fmt.Errorf("failed to create booking point of sale, %w", err)
	}

	response, err := d.lowribeckGw.CreateBookingPointOfSale(
		ctx,
		accountHolderDetails.Address.PAF.Postcode,
		accountHolderDetails.ElecOrderSupplies.MPXN,
		accountHolderDetails.GasOrderSupplies.MPXN,
		models.BookingTariffTypeToLowribeckTariffType(accountHolderDetails.ElecOrderSupplies.TariffType),
		models.BookingTariffTypeToLowribeckTariffType(accountHolderDetails.GasOrderSupplies.TariffType),
		params.Slot,
		params.ContactDetails,
		lbVulnerabilities,
		params.VulnerabilityDetails.Other,
	)
	if err != nil {
		return CreateBookingPointOfSaleResponse{}, fmt.Errorf("failed to create POS booking, %w", err)
	}
	if !response.Success {
		return CreateBookingPointOfSaleResponse{}, ErrUnsuccessfulBooking
	}

	commsEvent = buildPointOfSaleCommsEvent(params, *accountHolderDetails)

	bookingEvent = buildBookingEvent(params, *accountHolderDetails, response.ReferenceID, bookingID)

	occupancy, err := d.occupancyStore.GetOccupancyByAccountID(ctx, params.AccountID)
	if err != nil {
		if errors.Is(err, store.ErrOccupancyNotFound) {
			err := d.partialBookingStore.Upsert(ctx, bookingID, bookingEvent)
			if err != nil {
				return CreateBookingPointOfSaleResponse{}, fmt.Errorf("failed to insert partial booking store, %w", err)
			}

			return CreateBookingPointOfSaleResponse{
				BookingEvent: bookingEvent,
				CommsEvent:   commsEvent,
			}, ErrMissingOccupancyInBooking
		}
		return CreateBookingPointOfSaleResponse{}, fmt.Errorf("failed to get occupancy by id: %s, %w", params.AccountID, err)
	}

	bookingEvent.OccupancyId = occupancy.OccupancyID

	return CreateBookingPointOfSaleResponse{
		BookingEvent: bookingEvent,
		CommsEvent:   commsEvent,
	}, nil
}

// this method takes in an accountID and returns the postcode and the booking reference
func (d *BookingDomain) findLowriBeckKeys(ctx context.Context, accountID string) (_ models.Site, _ models.OccupancyEligibility, err error) {
	var span trace.Span
	if d.useTracing {
		ctx, span = tracing.Tracer().Start(ctx, "BookingAPI.BookingDomain.GetSiteExternalReferenceByAccountID")
		defer func() {
			tracing.RecordSpanError(span, err)
			span.End()
		}()
		span.AddEvent("request", trace.WithAttributes(attribute.String("account.id", accountID)))
	}

	site, occupancyEligible, err := d.occupancyStore.GetSiteExternalReferenceByAccountID(ctx, accountID)
	if err != nil {
		if errors.Is(err, store.ErrNoEligibleOccupancyFound) {
			return models.Site{}, models.OccupancyEligibility{}, ErrNoEligibleOccupanciesFound
		}
		return models.Site{}, models.OccupancyEligibility{}, fmt.Errorf("failed to get live occupancies by accountID, %w", err)
	}

	if d.useTracing {
		siteAttr := helpers.CreateSpanAttribute(site, "site", span)
		occupancyEligibleAttr := helpers.CreateSpanAttribute(occupancyEligible, "occupancyEligible", span)
		span.AddEvent("response", trace.WithAttributes(siteAttr, occupancyEligibleAttr))
	}

	return *site, *occupancyEligible, nil
}

func mapLowribeckVulnerabilities(vulnerabilities []bookingv1.Vulnerability) (lbVulnerabilities []lowribeckv1.Vulnerability) {
	for _, vulnerability := range vulnerabilities {
		lbVulnerabilities = append(lbVulnerabilities, models.BookingVulnerabilityToLowribeckVulnerability(vulnerability))
	}

	return
}

func buildPointOfSaleCommsEvent(params CreatePOSBookingParams, accountHolderDetails models.PointOfSaleCustomerDetails) *commsv1.PointOfSaleBookingConfirmationCommsEvent {
	event := &commsv1.PointOfSaleBookingConfirmationCommsEvent{
		AccountId:     params.AccountID,
		AccountNumber: params.AccountNumber,
		AccountHolderContactDetails: &bookingv1.ContactDetails{
			Title:     accountHolderDetails.Details.Title,
			FirstName: accountHolderDetails.Details.FirstName,
			LastName:  accountHolderDetails.Details.LastName,
			Phone:     accountHolderDetails.Details.Mobile,
			Email:     accountHolderDetails.Details.Email,
		},
		BookingDate: &date.Date{
			Year:  int32(params.Slot.Date.Year()),
			Month: int32(params.Slot.Date.Month()),
			Day:   int32(params.Slot.Date.Day()),
		},
		StartTime:      int32(params.Slot.StartTime),
		EndTime:        int32(params.Slot.EndTime),
		BookingType:    bookingv1.BookingType_BOOKING_TYPE_POINT_OF_SALE_JOURNEY,
		SupplyAddress:  toAddress(accountHolderDetails.Address),
		Mpan:           accountHolderDetails.ElecOrderSupplies.MPXN,
		Mprn:           accountHolderDetails.GasOrderSupplies.MPXN,
		ElecTariffType: accountHolderDetails.ElecOrderSupplies.TariffType,
		GasTariffType:  accountHolderDetails.GasOrderSupplies.TariffType,
	}

	if !params.ContactDetails.Empty() {
		event.OnSiteContactDetails = &bookingv1.ContactDetails{
			Title:     params.ContactDetails.Title,
			FirstName: params.ContactDetails.FirstName,
			LastName:  params.ContactDetails.LastName,
			Phone:     params.ContactDetails.Mobile,
			Email:     params.ContactDetails.Email,
		}
	}

	return event
}

func buildRescheduleCommsEvent(params RescheduleBookingParams, accountHolderDetails models.AccountDetails, siteAddress models.AccountAddress, accountNumber string) *commsv1.BookingRescheduledCommsEvent {
	event := &commsv1.BookingRescheduledCommsEvent{
		AccountId:     params.AccountID,
		AccountNumber: accountNumber,
		AccountHolderContactDetails: &bookingv1.ContactDetails{
			Title:     accountHolderDetails.Title,
			FirstName: accountHolderDetails.FirstName,
			LastName:  accountHolderDetails.LastName,
			Phone:     accountHolderDetails.Mobile,
			Email:     accountHolderDetails.Email,
		},
		BookingDate: &date.Date{
			Year:  int32(params.Slot.Date.Year()),
			Month: int32(params.Slot.Date.Month()),
			Day:   int32(params.Slot.Date.Day()),
		},
		StartTime:     int32(params.Slot.StartTime),
		EndTime:       int32(params.Slot.EndTime),
		SupplyAddress: toAddress(siteAddress),
	}

	if !params.ContactDetails.Equals(accountHolderDetails) {
		event.OnSiteContactDetails = &bookingv1.ContactDetails{
			Title:     params.ContactDetails.Title,
			FirstName: params.ContactDetails.FirstName,
			LastName:  params.ContactDetails.LastName,
			Phone:     params.ContactDetails.Mobile,
			Email:     params.ContactDetails.Email,
		}
	}

	return event
}

func buildBookingEvent(params CreatePOSBookingParams, accountHolderDetails models.PointOfSaleCustomerDetails, referenceID, bookingID string) *bookingv1.BookingCreatedEvent {
	return &bookingv1.BookingCreatedEvent{
		BookingId: bookingID,
		Details: &bookingv1.Booking{
			Id:        bookingID,
			AccountId: params.AccountID,
			ContactDetails: &bookingv1.ContactDetails{
				Title:     params.ContactDetails.Title,
				FirstName: params.ContactDetails.FirstName,
				LastName:  params.ContactDetails.LastName,
				Phone:     params.ContactDetails.Mobile,
				Email:     params.ContactDetails.Email,
			},
			Slot: &bookingv1.BookingSlot{
				Date: &date.Date{
					Year:  int32(params.Slot.Date.Year()),
					Month: int32(params.Slot.Date.Month()),
					Day:   int32(params.Slot.Date.Day()),
				},
				StartTime: int32(params.Slot.StartTime),
				EndTime:   int32(params.Slot.EndTime),
			},
			SiteAddress:          toAddress(accountHolderDetails.Address),
			VulnerabilityDetails: params.VulnerabilityDetails,
			Status:               bookingv1.BookingStatus_BOOKING_STATUS_SCHEDULED,
			ExternalReference:    referenceID,
			BookingType:          bookingv1.BookingType_BOOKING_TYPE_POINT_OF_SALE_JOURNEY,
		},
		BookingSource: params.Source,
	}
}
