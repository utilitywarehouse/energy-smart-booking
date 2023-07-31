package mapper

import (
	contract "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/lowribeck"
)

// TODO - SMT-176/get-calendar-availability
func MapAvailabilityRequest(req *contract.GetAvailableSlotsRequest) *lowribeck.GetCalendarAvailabilityRequest {
	return &lowribeck.GetCalendarAvailabilityRequest{}
}

// TODO - SMT-176/get-calendar-availability
func MapAvailableSlotsResponse(resp *lowribeck.GetCalendarAvailabilityResponse) (*contract.GetAvailableSlotsResponse, error) {

	return &contract.GetAvailableSlotsResponse{}, nil
}

// TODO - SMT-177/create-booking
func MapBookingRequest(req *contract.CreateBookingRequest) *lowribeck.CreateBookingRequest {
	return &lowribeck.CreateBookingRequest{}
}

// TODO - SMT-177/create-booking
func MapBookingResponse(_ *lowribeck.CreateBookingResponse) *contract.CreateBookingResponse {
	return &contract.CreateBookingResponse{}
}
