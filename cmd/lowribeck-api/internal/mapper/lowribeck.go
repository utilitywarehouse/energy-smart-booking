package mapper

import (
	"fmt"
	"strings"
	"time"

	contract "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/lowribeck"
	"google.golang.org/genproto/googleapis/type/date"
)

const (
	requestTimeFormat     = "02/01/2006 15:04:05"
	appointmentDateFormat = "02/01/2006"
	appointmentTimeFormat = "%d:00-%d:00"
)

type LowriBeck struct {
	sendingSystem   string
	receivingSystem string
}

func NewLowriBeckMapper(sendingSystem, receivingSystem string) *LowriBeck {
	return &LowriBeck{
		sendingSystem:   sendingSystem,
		receivingSystem: receivingSystem,
	}
}

func (lb LowriBeck) AvailabilityRequest(id uint32, req *contract.GetAvailableSlotsRequest) *lowribeck.GetCalendarAvailabilityRequest {
	return &lowribeck.GetCalendarAvailabilityRequest{
		PostCode:        req.GetPostcode(),
		ReferenceID:     req.GetReference(),
		SendingSystem:   lb.sendingSystem,
		ReceivingSystem: lb.receivingSystem,
		CreatedDate:     time.Now().UTC().Format(requestTimeFormat),
		// An ID sent to LB which they return in the response and can be used for debugging issues with them
		RequestID: fmt.Sprintf("%d", id),
	}
}

func (lb LowriBeck) AvailableSlotsResponse(resp *lowribeck.GetCalendarAvailabilityResponse) (*contract.GetAvailableSlotsResponse, error) {
	slots, err := mapAvailabilitySlots(resp.CalendarAvailabilityResult)
	if err != nil {
		return nil, err
	}

	var code *contract.AvailabilityErrorCodes
	if resp.ResponseCode != "" {
		code = mapAvailabilityErrorCodes(resp.ResponseCode)
	}
	return &contract.GetAvailableSlotsResponse{
		Slots:      slots,
		ErrorCodes: code,
	}, nil
}

func (lb LowriBeck) BookingRequest(id uint32, req *contract.CreateBookingRequest) (*lowribeck.CreateBookingRequest, error) {
	appDate, appTime, err := mapBookingSlot(req.GetSlot())
	if err != nil {
		return nil, err
	}
	return &lowribeck.CreateBookingRequest{
		PostCode:          req.GetPostcode(),
		ReferenceID:       req.GetReference(),
		AppointmentDate:   appDate,
		AppointmentTime:   appTime,
		Vulnerabilities:   mapVulnerabilities(req.GetVulnerabilityDetails()),
		SiteContactName:   mapContactName(req.GetContactDetails()),
		SiteContactNumber: req.GetContactDetails().GetPhone(),
		SendingSystem:     lb.sendingSystem,
		ReceivingSystem:   lb.receivingSystem,
		CreatedDate:       time.Now().UTC().Format(requestTimeFormat),
		// An ID sent to LB which they return in the response and can be used for debugging issues with them
		RequestID: fmt.Sprintf("%d", id),
	}, nil
}

func (lb LowriBeck) BookingResponse(resp *lowribeck.CreateBookingResponse) (*contract.CreateBookingResponse, error) {
	success, code := mapBookingResponseCodes(resp.ResponseCode, resp.ResponseMessage)
	return &contract.CreateBookingResponse{
		Success:    success,
		ErrorCodes: code,
	}, nil
}

// TODO REMOVE
// {
// 	"RequestId": "1234234",
// 	"SendingSystem": "UTIL",
// 	"ReceivingSystem": "LB",
// 	"CreatedDate": "25/07/2020 12:47:41",
// 	 "ReferenceId": "UTIL/4568973",
// 	"AppointmentDate": "15/08/2023",
// 	"AppointmentTime": "10:00-14:00",
// 	 "PostCode": "LS8 4EX",
// 	"Mpan": "",
// 	"Mprn": ""
// 	}

// {
//     "ReferenceId": "UTIL/4568973",
//     "Mpan": "",
//     "Mprn": "",
//     "ResponseMessage": "Booking Confirmed",
//     "ResponseCode": "B01",
//     "RequestId": "1234234",
//     "SendingSystem": "LB",
//     "ReceivingSystem": "UTIL",
//     "CreatedDate": "26/07/2023 14:24:34"
// }

func mapAvailabilitySlots(availabilityResults []lowribeck.AvailabilitySlot) ([]*contract.BookingSlot, error) {
	var err error
	slots := make([]*contract.BookingSlot, len(availabilityResults))
	for i, res := range availabilityResults {
		slot := &contract.BookingSlot{}
		slot.Date, err = mapAvailabilityAppointmentDate(res.AppointmentDate)
		if err != nil {
			return nil, fmt.Errorf("error converting appointment date: %v", err)
		}

		slot.StartTime, slot.EndTime, err = mapAvailabilityAppointmentTime(res.AppointmentTime)
		if err != nil {
			return nil, fmt.Errorf("error converting appointment time: %v", err)
		}

		slots[i] = slot
	}
	return slots, nil
}

func mapAvailabilityAppointmentDate(appointmentDate string) (*date.Date, error) {
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

func mapAvailabilityAppointmentTime(appointmentTime string) (int32, int32, error) {
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

func mapAvailabilityErrorCodes(responseCode string) *contract.AvailabilityErrorCodes {
	switch responseCode {
	case "EA01":
		return contract.AvailabilityErrorCodes_AVAILABILITY_NO_AVAILABLE_SLOTS.Enum()
	case "EA02", "EA03":
		return contract.AvailabilityErrorCodes_AVAILABILITY_INVALID_REQUEST.Enum()
	}
	return contract.AvailabilityErrorCodes_AVAILABILITY_INTERNAL_ERROR.Enum()
}

// TODO
func mapBookingResponseCodes(responseCode, responseMessage string) (bool, *contract.BookingErrorCodes) {
	switch responseCode {
	// B01 - Booking Confirmed
	case "B01":
		return true, nil
		// B02 - Appointment not available
	case "B02":
		return false, contract.BookingErrorCodes_BOOKING_APPOINTMENT_UNAVAILABLE.Enum()
	// B03 - Invalid Elec Job Type Code
	// B03 - Invalid Gas Job Type Code
	// B04 - Invalid MPAN
	// B05 - Invalid MPRN
	// B06 - Invalid Appt Date
	// B06 – Invalid Date Format
	// B07 - Invalid Appt Time
	// B13 - Invalid Reference ID
	case "B03", "B04", "B05", "B06", "B07", "B13":
		return false, contract.BookingErrorCodes_BOOKING_INVALID_REQUEST.Enum()
	// B08 - Duplicate Elec job exists
	// B08 - Duplicate Gas job exists
	case "B08":
		return false, contract.BookingErrorCodes_BOOKING_DUPLICATE_JOB_EXISTS.Enum()
	case "B09":
		switch responseMessage {
		// B09 - No available slots for requested postcode
		// B09 - Rearranging request sent outside agreed time parameter
		// B09 - Booking request sent outside agreed time parameter
		case "No available slots for requested postcode",
			"Rearranging request sent outside agreed time parameter",
			"Booking request sent outside agreed time parameter":
			return false, contract.BookingErrorCodes_BOOKING_NO_AVAILABLE_SLOTS.Enum()
		// B09 - Site status not suitable for request
		// B09 - Not available as site is complete
		case "Site status not suitable for request",
			"Not available as site is complete":
			return false, contract.BookingErrorCodes_BOOKING_INVALID_SITE.Enum()
		// B09 - Post Code is missing or invalid
		// B09 - Postcode and Reference ID mismatch
		case "Post Code is missing or invalid",
			"Postcode and Reference ID mismatch":
			return false, contract.BookingErrorCodes_BOOKING_POSTCODE_REFERENCE_MISMATCH.Enum()
		}

	}
	// BookingErrorCodes_BOOKING_APPOINTMENT_UNAVAILABLE     BookingErrorCodes = 0
	// BookingErrorCodes_BOOKING_NO_AVAILABLE_SLOTS          BookingErrorCodes = 1
	// BookingErrorCodes_BOOKING_INVALID_REQUEST             BookingErrorCodes = 2
	// BookingErrorCodes_BOOKING_DUPLICATE_JOB_EXISTS        BookingErrorCodes = 3
	// BookingErrorCodes_BOOKING_POSTCODE_REFERENCE_MISMATCH BookingErrorCodes = 4
	// BookingErrorCodes_BOOKING_TIMEOUT                     BookingErrorCodes = 5
	// BookingErrorCodes_BOOKING_INVALID_SITE                BookingErrorCodes = 6
	// BookingErrorCodes_BOOKING_INTERNAL_ERROR              BookingErrorCodes = 7

	// B09 - Generic LB error for failed to process request

	// B09 - No available slots for requested postcode
	// B09 - Rearranging request sent outside agreed time parameter
	// B09 - Booking request sent outside agreed time parameter

	// B09 - Emergency jobs cannot be rescheduled
	// B09 - This jobs is currently on hold
	// B09 - No Jobs found for Reference ID
	// B09 - Failed to create Elec job
	// B09 - Failed to create Gas job

	// B09 - Unable able to reschedule the appt
	// B10 - Not appointed as MOP or MAM

	// B11 - No Elec GUID info for SMETS2 maintenance job request
	// B11 - No Gas GUID info for SMETS2 maintenance job request
	// B11 – No Comms Hub GUID info for SMETS2 maintenance job request
	// B12 - No SSC info for trad elec maintenance job request
	// B13 - Invalid Reference ID
	// S01 – Invalid sending system
	// S02 – Invalid receiving system

	return false, contract.BookingErrorCodes_BOOKING_INTERNAL_ERROR.Enum()
}

func mapBookingSlot(slot *contract.BookingSlot) (string, string, error) {
	if slot == nil {
		return "", "", fmt.Errorf("invalid booking slot time")
	}
	slotDate := slot.GetDate()
	if slotDate == nil {
		return "", "", fmt.Errorf("invalid booking slot date")
	}
	appDate := fmt.Sprintf("%02d/%02d/%4d", slotDate.Day, slotDate.Month, slotDate.Year)
	appTime := fmt.Sprintf(appointmentTimeFormat, slot.StartTime, slot.EndTime)

	return appDate, appTime, nil
}

// TODO
func mapVulnerabilities(contact *contract.VulnerabilityDetails) string {
	// 01 - Hearing Impaired
	// 02 - Visually Impaired
	// 03 - Elderly
	// 04 - Disabled
	// 05 - Electrical Medical Equipment
	// 06 - Foreign Language Speaker
	// 07 - Restricted Movement
	// 08 - Serious Illness
	// 09 - Other
	return ""
}

func mapContactName(contact *contract.ContactDetails) string {
	return strings.TrimSpace(contact.GetFirstName() + " " + contact.GetLastName())
}
