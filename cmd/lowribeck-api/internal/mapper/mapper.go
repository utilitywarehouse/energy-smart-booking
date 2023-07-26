package mapper

import (
	"time"

	contract "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/lowribeck"
)

const (
	sendingSystem   = "UTIL"
	receivingSystem = "LB"
)

// TODO
func MapAvailabilityRequest(req *contract.GetAvailableSlotsRequest) *lowribeck.GetCalendarAvailabilityRequest {
	return &lowribeck.GetCalendarAvailabilityRequest{
		PostCode:        req.GetPostcode(),
		ReferenceID:     req.GetReference(),
		SendingSystem:   sendingSystem,
		ReceivingSystem: receivingSystem,
		CreatedDate:     time.Now().UTC().String(),
	}
}

// TODO
func MapAvailableSlotsResponse(resp *lowribeck.GetCalendarAvailabilityResponse) *contract.GetAvailableSlotsResponse {
	return &contract.GetAvailableSlotsResponse{}
}

// TODO
func MapBookingRequest(req *contract.CreateBookingRequest) *lowribeck.CreateBookingRequest {
	return &lowribeck.CreateBookingRequest{
		PostCode:        req.GetPostcode(),
		ReferenceID:     req.GetReference(),
		SendingSystem:   sendingSystem,
		ReceivingSystem: receivingSystem,
		CreatedDate:     time.Now().UTC().String(),
	}
}

// TODO
func MapBookingResponse(resp *lowribeck.CreateBookingResponse) *contract.CreateBookingResponse {
	return &contract.CreateBookingResponse{}
}
