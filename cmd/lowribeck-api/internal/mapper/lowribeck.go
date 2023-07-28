package mapper

import (
	"fmt"
	"time"

	contract "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/lowribeck"
	"google.golang.org/genproto/googleapis/type/date"
)

const (
	sendingSystem         = "UTIL"
	receivingSystem       = "LB"
	requestTimeFormat     = "02/01/2006 15:04:05"
	appointmentDateFormat = "02/01/2006"
	appointmentTimeFormat = "%d:00-%d:00"
)

func MapAvailabilityRequest(id string, req *contract.GetAvailableSlotsRequest) *lowribeck.GetCalendarAvailabilityRequest {
	return &lowribeck.GetCalendarAvailabilityRequest{
		PostCode:        req.GetPostcode(),
		ReferenceID:     req.GetReference(),
		SendingSystem:   sendingSystem,
		ReceivingSystem: receivingSystem,
		CreatedDate:     time.Now().UTC().Format(requestTimeFormat),
		// An ID sent to LB which they return in the response and can be used for debugging issues with them
		RequestID: id,
	}
}

func MapAvailableSlotsResponse(resp *lowribeck.GetCalendarAvailabilityResponse) (*contract.GetAvailableSlotsResponse, error) {
	slots, err := mapAvailabilitySlots(resp.CalendarAvailabilityResult)
	if err != nil {
		return nil, err
	}

	var code *contract.AvailabilityErrorCodes
	if resp.ResponseCode != "" {
		errorCode := mapAvailabilityErrorCodes(resp.ResponseMessage)
		code = &errorCode
	}
	return &contract.GetAvailableSlotsResponse{
		Slots:      slots,
		ErrorCodes: code,
	}, nil
}

// TODO - SMT-177/create-booking
func MapBookingRequest(req *contract.CreateBookingRequest) *lowribeck.CreateBookingRequest {
	return &lowribeck.CreateBookingRequest{}
}

// TODO - SMT-177/create-booking
func MapBookingResponse(_ *lowribeck.CreateBookingResponse) *contract.CreateBookingResponse {
	return &contract.CreateBookingResponse{}
}

func mapAvailabilitySlots(availabilityResults []*lowribeck.AvailabilitySlot) ([]*contract.BookingSlot, error) {
	var err error
	slots := make([]*contract.BookingSlot, len(availabilityResults))
	for i, res := range availabilityResults {
		slots[i] = &contract.BookingSlot{}
		slots[i].Date, err = mapAppointmentDate(res.AppointmentDate)
		if err != nil {
			return nil, fmt.Errorf("error converting appointment date: %v", err)
		}

		slots[i].StartTime, slots[i].EndTime, err = mapAppointmentTime(res.AppointmentTime)
		if err != nil {
			return nil, fmt.Errorf("error converting appointment time: %v", err)
		}
	}
	return slots, nil
}

func mapAppointmentDate(appointmentDate string) (*date.Date, error) {
	appDate, err := time.Parse(appointmentDateFormat, appointmentDate)
	if err != nil {
		return nil, err
	}

	return &date.Date{
		Year:  int32(appDate.Year()),
		Month: int32(appDate.Month()),
		Day:   int32(appDate.Day()),
	}, nil

}

func mapAppointmentTime(appointmentTime string) (int32, int32, error) {
	var start, end int32
	read, err := fmt.Sscanf(appointmentTime, appointmentTimeFormat, &start, &end)
	if err != nil {
		return -1, -1, err
	}
	if read != 2 {
		return -1, -1, fmt.Errorf("could not find start and end time: %q", appointmentTime)
	}
	if start < 0 || start > 23 {
		return -1, -1, fmt.Errorf("invalid start time: %q", appointmentTime)
	}
	if end < 0 || end > 23 {
		return -1, -1, fmt.Errorf("invalid end time: %q", appointmentTime)
	}
	if start > end {
		return -1, -1, fmt.Errorf("invalid appointment time: %q", appointmentTime)
	}

	return start, end, nil
}

// TODO
func mapAvailabilityErrorCodes(responseMessage string) contract.AvailabilityErrorCodes {
	// Should be responseCode, not responseMessage - lowribeck API error
	switch responseMessage {
	case "EA01":
		return contract.AvailabilityErrorCodes_AVAILABILITY_NO_AVAILABLE_SLOTS
	case "EA02", "EA03":
		return contract.AvailabilityErrorCodes_AVAILABILITY_INVALID_REQUEST
	}
	return contract.AvailabilityErrorCodes_AVAILABILITY_INTERNAL_ERROR
}
