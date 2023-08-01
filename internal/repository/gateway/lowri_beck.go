package gateway

import (
	"context"
	"fmt"

	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

type LowriBeckGateway struct {
	mai    MachineAuthInjector
	client LowriBeckClient
}

func NewLowriBeckGateway(mai MachineAuthInjector, client LowriBeckClient) LowriBeckGateway {
	return LowriBeckGateway{mai, client}
}

func (g LowriBeckGateway) GetAvailableSlots(ctx context.Context, postcode, reference string) ([]models.Slot, error) {
	availableSlots, err := g.client.GetAvailableSlots(g.mai.ToCtx(ctx), &lowribeckv1.GetAvailableSlotsRequest{
		Postcode:  postcode,
		Reference: reference,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get available slots, %w", err)
	}

	slots := make([]models.Slot, len(availableSlots.GetSlots()))

	for _, elem := range availableSlots.GetSlots() {
		slots = append(slots, models.Slot{
			Date:      *elem.GetDate(),
			StartTime: elem.GetStartTime(),
			EndTime:   elem.GetEndTime(),
		})
	}

	return slots, nil
}

func (g LowriBeckGateway) CreateBooking(ctx context.Context, postcode, reference string, slot models.Slot, accountDetails models.AccountDetails, vulnerabilities []lowribeckv1.Vulnerability, other string) (bool, error) {
	bookingResponse, err := g.client.CreateBooking(g.mai.ToCtx(ctx), &lowribeckv1.CreateBookingRequest{
		Postcode:  postcode,
		Reference: reference,
		Slot: &lowribeckv1.BookingSlot{
			Date:      &slot.Date,
			StartTime: slot.StartTime,
			EndTime:   slot.EndTime,
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
		return bookingResponse.Success, fmt.Errorf("failed to get available slots, reason: %s, %w", bookingResponse.GetErrorCodes().String(), err)
	}

	return bookingResponse.Success, nil
}
