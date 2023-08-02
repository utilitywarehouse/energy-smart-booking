package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	addressv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/address/v1"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/protobuf/proto"
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
	VulnerabilityDetails *bookingv1.VulnerabilityDetails
}

type RescheduleBookingParams struct {
	AccountID string
	BookingID string
	Slot      models.BookingSlot
}

type GetAvailableSlotsResponse struct {
	Slots []models.BookingSlot
}

func (d BookingDomain) GetAvailableSlots(ctx context.Context, params GetAvailableSlotsParams) (GetAvailableSlotsResponse, error) {
	fromAsTime := time.Date(int(params.From.Year), time.Month(params.From.Month), int(params.From.Day), 0, 0, 0, 0, time.UTC)
	toAsTime := time.Date(int(params.To.Year), time.Month(params.To.Month), int(params.To.Day), 0, 0, 0, 0, time.UTC)

	site, bookingReference, err := d.findLowriBeckKeys(ctx, params.AccountID)
	if err != nil {
		return GetAvailableSlotsResponse{}, fmt.Errorf("failed to find postcode and booking reference, %w", err)
	}

	slots, err := d.lowribeckGw.GetAvailableSlots(ctx, site.Postcode, bookingReference)
	if err != nil {
		return GetAvailableSlotsResponse{}, fmt.Errorf("failed to get available slots, %w", err)
	}

	targetedSlots := []models.BookingSlot{}

	for _, elem := range slots {
		currentSlotTime := time.Date(elem.Date.Year(), elem.Date.Month(), elem.Date.Day(), 0, 0, 0, 0, time.UTC)

		if currentSlotTime.After(fromAsTime) && currentSlotTime.Before(toAsTime) {
			targetedSlots = append(targetedSlots, elem)
		}
	}

	return GetAvailableSlotsResponse{
		Slots: targetedSlots,
	}, nil

}

func (d BookingDomain) CreateBooking(ctx context.Context, params CreateBookingParams) (proto.Message, error) {

	var event *bookingv1.BookingCreatedEvent

	lbVulnerabilities := mapLowribeckVulnerabilities(params.VulnerabilityDetails.GetVulnerabilities())

	site, bookingReference, err := d.findLowriBeckKeys(ctx, params.AccountID)
	if err != nil {
		return nil, err
	}

	response, err := d.lowribeckGw.CreateBooking(ctx, site.Postcode, bookingReference, params.Slot, params.ContactDetails, lbVulnerabilities, params.VulnerabilityDetails.Other)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking, %w", err)
	}

	if response {
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
			},
		}
	}

	return event, nil
}

func (d BookingDomain) RescheduleBooking(ctx context.Context, params RescheduleBookingParams) (proto.Message, error) {

	var event *bookingv1.BookingRescheduledEvent = nil

	site, bookingReference, err := d.findLowriBeckKeys(ctx, params.AccountID)
	if err != nil {
		return nil, err
	}

	booking, err := d.bookingStore.GetBookingByBookingID(ctx, params.BookingID)
	if err != nil {
		return nil, err
	}

	lbVulnerabilities := mapLowribeckVulnerabilities(booking.VulnerabilityDetails.Vulnerabilities)

	response, err := d.lowribeckGw.CreateBooking(ctx, site.Postcode, bookingReference, params.Slot, booking.Contact, lbVulnerabilities, booking.VulnerabilityDetails.Other)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking, %w", err)
	}

	if response {
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
		}
	}

	return event, nil
}

// this method takes in an accountID and returns the postcode and the booking reference
func (d *BookingDomain) findLowriBeckKeys(ctx context.Context, accountID string) (models.Site, string, error) {

	var targetOccupancy models.Occupancy = models.Occupancy{}

	liveOccupancies, err := d.occupancyStore.GetLiveOccupanciesByAccountID(ctx, accountID)
	if err != nil {
		return models.Site{}, "", fmt.Errorf("failed to get live occupancies by accountID, %w", err)
	}

	if len(liveOccupancies) == 0 {
		return models.Site{}, "", ErrNoOccupanciesFound
	}

	for _, occupancy := range liveOccupancies {
		isEligible, err := d.eligibilityGw.GetEligibility(ctx, accountID, occupancy.OccupancyID)
		if err != nil {
			return models.Site{}, "", fmt.Errorf("failed to get eligibility for accountId: %s, occupancyId: %s, %w", accountID, occupancy.OccupancyID, err)
		}

		if isEligible {
			targetOccupancy = occupancy
			break
		}
	}

	if targetOccupancy.IsEmpty() {
		return models.Site{}, "", ErrNoEligibleOccupanciesFound
	}

	site, err := d.siteStore.GetSiteByOccupancyID(ctx, targetOccupancy.OccupancyID)
	if err != nil {
		return models.Site{}, "", fmt.Errorf("failed to get site with site_id :%s, %w", targetOccupancy.SiteID, err)
	}

	reference, err := d.serviceStore.GetReferenceByOccupancyID(ctx, targetOccupancy.OccupancyID)
	if err != nil {
		return models.Site{}, "", nil
	}

	return *site, reference, nil
}

func mapLowribeckVulnerabilities(vulnerabilities []bookingv1.Vulnerability) (lbVulnerabilities []lowribeckv1.Vulnerability) {
	for _, vulnerability := range vulnerabilities {
		lbVulnerabilities = append(lbVulnerabilities, models.BookingVulnerabilityToLowribeckVulnerability(vulnerability))
	}

	return
}
