package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	addressv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/address/v1"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
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
	AccountID string
	BookingID string
	Source    bookingv1.BookingSource
	Slot      models.BookingSlot
}

type GetAvailableSlotsResponse struct {
	Slots []models.BookingSlot
}

type CreateBookingResponse struct {
	Event proto.Message
}

type RescheduleBookingResponse struct {
	Event proto.Message
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

	if response.Success {
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
				Status:               bookingv1.BookingStatus_BOOKING_STATUS_COMPLETED,
				ExternalReference:    occupancyEligibility.Reference,
			},
			OccupancyId:   occupancyEligibility.OccupancyID,
			BookingSource: params.Source,
		}
	}

	return CreateBookingResponse{
		Event: event,
	}, nil
}

func (d BookingDomain) RescheduleBooking(ctx context.Context, params RescheduleBookingParams) (RescheduleBookingResponse, error) {

	var event *bookingv1.BookingRescheduledEvent = nil

	site, occupancyEligibility, err := d.findLowriBeckKeys(ctx, params.AccountID)
	if err != nil {
		return RescheduleBookingResponse{}, err
	}

	booking, err := d.bookingStore.GetBookingByBookingID(ctx, params.BookingID)
	if err != nil {
		return RescheduleBookingResponse{}, err
	}

	lbVulnerabilities := mapLowribeckVulnerabilities(booking.VulnerabilityDetails.Vulnerabilities)

	response, err := d.lowribeckGw.CreateBooking(ctx, site.Postcode, occupancyEligibility.Reference, params.Slot, booking.Contact, lbVulnerabilities, booking.VulnerabilityDetails.Other)
	if err != nil {
		return RescheduleBookingResponse{}, fmt.Errorf("failed to reschedule booking, %w", err)
	}

	if response.Success {
		event = &bookingv1.BookingRescheduledEvent{
			BookingId: params.BookingID,
			Slot: &bookingv1.BookingSlot{
				Date: &date.Date{
					Year:  int32(params.Slot.Date.Year()),
					Month: int32(params.Slot.Date.Month()),
					Day:   int32(params.Slot.Date.Day()),
				},
				StartTime: int32(params.Slot.StartTime),
				EndTime:   int32(params.Slot.EndTime),
			},
			BookingSource: params.Source,
		}
	}

	return RescheduleBookingResponse{
		Event: event,
	}, nil
}

// this method takes in an accountID and returns the postcode and the booking reference
func (d *BookingDomain) findLowriBeckKeys(ctx context.Context, accountID string) (_ models.Site, _ models.OccupancyEligibility, err error) {
	var span trace.Span
	if d.useTracing {
		ctx, span = tracing.Tracer().Start(ctx, "BookingAPI.BookingDomain.GetSiteExternalReferenceByAccountID")
		defer func() {
			tracing.RecordSpanError(span, err) // nolint: errcheck
			span.End()
		}()
		span.AddEvent("request", trace.WithAttributes(attribute.String("accountID", accountID)))
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
