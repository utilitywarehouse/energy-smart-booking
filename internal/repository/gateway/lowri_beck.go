package gateway

import (
	"context"
	"fmt"
	"time"

	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"google.golang.org/genproto/googleapis/type/date"
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
	ErrorCode    *bookingv1.AvailabilityErrorCodes
}

type CreateBookingResponse struct {
	Success   bool
	ErrorCode *bookingv1.BookingErrorCodes
}

func (g LowriBeckGateway) GetAvailableSlots(ctx context.Context, postcode, reference string) (AvailableSlotsResponse, error) {
	var errorCode *bookingv1.AvailabilityErrorCodes = nil

	availableSlots, err := g.client.GetAvailableSlots(g.mai.ToCtx(ctx), &lowribeckv1.GetAvailableSlotsRequest{
		Postcode:  postcode,
		Reference: reference,
	})
	if err != nil {
		return AvailableSlotsResponse{}, fmt.Errorf("failed to get available slots, %w", err)
	}

	slots := []models.BookingSlot{}

	for _, elem := range availableSlots.GetSlots() {
		slots = append(slots, models.BookingSlot{
			Date:      time.Date(int(elem.Date.Year), time.Month(elem.Date.Month), int(elem.Date.Day), 0, 0, 0, 0, time.UTC),
			StartTime: int(elem.GetStartTime()),
			EndTime:   int(elem.GetEndTime()),
		})
	}

	if availableSlots.ErrorCodes != nil {
		eCode, err := models.AvailabilityLowriBeckErrorCodeToBookingErrorCode(*availableSlots.ErrorCodes)
		if err != nil {
			return AvailableSlotsResponse{}, err
		}

		errorCode = &eCode
	}

	return AvailableSlotsResponse{
		BookingSlots: slots,
		ErrorCode:    errorCode,
	}, nil
}

func (g LowriBeckGateway) CreateBooking(ctx context.Context, postcode, reference string, slot models.BookingSlot, accountDetails models.AccountDetails, vulnerabilities []lowribeckv1.Vulnerability, other string) (CreateBookingResponse, error) {
	var errorCode *bookingv1.BookingErrorCodes = nil

	bookingResponse, err := g.client.CreateBooking(g.mai.ToCtx(ctx), &lowribeckv1.CreateBookingRequest{
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
			Title:     accountDetails.Title,
			FirstName: accountDetails.FirstName,
			LastName:  accountDetails.LastName,
			Phone:     accountDetails.Mobile,
		},
	})
	if err != nil {
		return CreateBookingResponse{}, fmt.Errorf("failed to get available slots, %w", err)
	}

	if bookingResponse.ErrorCodes != nil {
		eCode, err := models.BookingLowriBeckErrorCodeToBookingErrorCode(*bookingResponse.ErrorCodes)
		if err != nil {
			return CreateBookingResponse{}, err
		}

		errorCode = &eCode
	}

	return CreateBookingResponse{
		Success:   bookingResponse.Success,
		ErrorCode: errorCode,
	}, nil
}
