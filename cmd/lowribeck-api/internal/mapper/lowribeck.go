package mapper

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	contract "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/lowribeck"
	"google.golang.org/genproto/googleapis/type/date"
)

const (
	sendingSystem      = "UTIL"
	receivingSystem    = "LB"
	requestTimeFormat  = "02/01/2006 15:04:05"
	responseDateFormat = "02/01/2006"
	responseTimeFormat = "%d:00-%d:00"
)

// TODO
func MapAvailabilityRequest(req *contract.GetAvailableSlotsRequest) *lowribeck.GetCalendarAvailabilityRequest {
	return &lowribeck.GetCalendarAvailabilityRequest{
		PostCode:        req.GetPostcode(),
		ReferenceID:     req.GetReference(),
		SendingSystem:   sendingSystem,
		ReceivingSystem: receivingSystem,
		CreatedDate:     time.Now().UTC().Format(requestTimeFormat),
	}
}

// TODO
func MapAvailableSlotsResponse(resp *lowribeck.GetCalendarAvailabilityResponse) (*contract.GetAvailableSlotsResponse, error) {
	slots, err := MapAvailabilitySlots(resp.CalendarAvailabilityResult)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Slots: %v", slots)
	// var code contract.AvailabilityErrorCodes
	// if resp.ResponseCode != "" {
	// 	code = MapErrorCodes(resp.ResponseCode, resp.ResponseMessage)
	// }
	return &contract.GetAvailableSlotsResponse{
		Slots:      slots,
		ErrorCodes: MapErrorCodes(resp.ResponseCode, resp.ResponseMessage),
	}, nil
}

// TODO
func MapBookingRequest(req *contract.CreateBookingRequest) *lowribeck.CreateBookingRequest {
	return &lowribeck.CreateBookingRequest{
		PostCode:        req.GetPostcode(),
		ReferenceID:     req.GetReference(),
		SendingSystem:   sendingSystem,
		ReceivingSystem: receivingSystem,
		CreatedDate:     time.Now().UTC().Format(requestTimeFormat),
	}
}

// TODO
func MapBookingResponse(_ *lowribeck.CreateBookingResponse) *contract.CreateBookingResponse {
	return &contract.CreateBookingResponse{}
}

func MapAvailabilitySlots(availabilityResults []*lowribeck.AvailabilitySlot) ([]*contract.BookingSlot, error) {
	var err error
	logrus.Debugf("Results: %d", len(availabilityResults))
	slots := make([]*contract.BookingSlot, len(availabilityResults))
	for i, res := range availabilityResults {
		logrus.Debugf("Counter: %d, App Date: %s. App Time: %s", i, res.AppointmentDate, res.AppointmentTime)
		slots[i] = &contract.BookingSlot{}
		slots[i].Date, err = MapAppointmentDate(res.AppointmentDate)
		if err != nil {
			return nil, fmt.Errorf("error converting appointment date: %v", err)
		}
		logrus.Debugf("Date: %s", slots[i].Date)

		slots[i].StartTime, slots[i].EndTime, err = MapAppointmentTime(res.AppointmentTime)
		if err != nil {
			return nil, fmt.Errorf("error converting appointment time: %v", err)
		}
		logrus.Debugf("Start: %d, end: %d", slots[i].StartTime, slots[i].EndTime)
	}
	return slots, nil
}

func MapAppointmentDate(appointmentDate string) (*date.Date, error) {
	appDate, err := time.Parse(responseDateFormat, appointmentDate)
	if err != nil {
		return nil, err
	}

	return &date.Date{
		Year:  int32(appDate.Year()),
		Month: int32(appDate.Month()),
		Day:   int32(appDate.Day()),
	}, nil

}

func MapAppointmentTime(appointmentTime string) (int32, int32, error) {
	var start, end int32
	read, err := fmt.Sscanf(appointmentTime, responseTimeFormat, &start, &end)
	if err != nil {
		return -1, -1, err
	}
	if read != 2 {
		return -1, -1, fmt.Errorf("could not find start and end time: %q", appointmentTime)
	}

	return start, end, nil
}

// TODO
func MapErrorCodes(responseCode, responseMessage string) contract.AvailabilityErrorCodes {
	switch responseMessage {
	case "EA03":
		if responseCode != "" {
			return contract.AvailabilityErrorCodes_AVAILABILITY_INVALID_REQUEST
		}
	}
	return contract.AvailabilityErrorCodes_AVAILABILITY_INTERNAL_ERROR
}
