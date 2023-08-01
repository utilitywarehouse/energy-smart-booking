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
	From      date.Date
	To        date.Date
}

type CreateBookingParams struct {
	AccountID            string
	ContactDetails       models.AccountDetails
	Slot                 models.Slot
	VulnerabilityDetails *bookingv1.VulnerabilityDetails
}

type RescheduleBookingParams struct {
	AccountID string
	BookingID string
	Slot      models.Slot
}

type GetAvailableSlotsResponse struct {
	Slots []models.Slot
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

	targetedSlots := make([]models.Slot, 0)

	for _, elem := range slots {
		currentSlotTime := time.Date(int(elem.Date.Year), time.Month(elem.Date.Month), int(elem.Date.Day), 0, 0, 0, 0, time.UTC)

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

	lbVulnerabilities := []lowribeckv1.Vulnerability{}

	for _, elem := range params.VulnerabilityDetails.GetVulnerabilities() {
		lbVulnerabilities = append(lbVulnerabilities, models.BookingVulnerabilityToLowribeckVulnerability(elem))
	}

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
					Date:      &params.Slot.Date,
					StartTime: params.Slot.StartTime,
					EndTime:   params.Slot.EndTime,
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

	_, err = d.bookingStore.GetBookingByBookingID(ctx, params.BookingID)
	if err != nil {
		return nil, err
	}

	// MISSING DB CALL TO GET PREVIOUS CREATED_BOOKING

	response, err := d.lowribeckGw.CreateBooking(ctx, site.Postcode, bookingReference, params.Slot, models.AccountDetails{}, []lowribeckv1.Vulnerability{}, "vulnerabilities.other")
	if err != nil {
		return nil, fmt.Errorf("failed to create booking, %w", err)
	}

	if response {
		event = &bookingv1.BookingRescheduledEvent{
			BookingId: params.BookingID,
			Slot: &bookingv1.BookingSlot{
				Date:      &params.Slot.Date,
				StartTime: params.Slot.StartTime,
				EndTime:   params.Slot.EndTime,
			},
		}
	}

	return event, nil
}

// this method takes in an accountID and returns the postcode and the booking reference
func (d BookingDomain) findLowriBeckKeys(ctx context.Context, accountID string) (models.Site, string, error) {

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

	site, err := d.siteStore.GetSiteBySiteID(ctx, targetOccupancy.SiteID)
	if err != nil {
		return models.Site{}, "", fmt.Errorf("failed to get site with site_id :%s, %w", targetOccupancy.SiteID, err)
	}

	mpxn, err := d.serviceStore.GetServiceMPXNByOccupancyID(ctx, targetOccupancy.OccupancyID)
	if err != nil {
		return models.Site{}, "", nil
	}

	reference, err := d.bookingReferenceStore.GetReferenceByMPXN(ctx, mpxn)
	if err != nil {
		return models.Site{}, "", nil
	}

	return *site, reference, nil
}
