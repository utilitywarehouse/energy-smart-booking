package mapper

import (
	"errors"
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

var (
	ErrInvalidRequest           = errors.New("invalid request")
	ErrAppointmentNotFound      = errors.New("no appointments found")
	ErrAppointmentOutOfRange    = errors.New("appointment out of range")
	ErrAppointmentAlreadyExists = errors.New("appointment already exists")
	ErrInternalError            = errors.New("internal server error")

	invalidPostcode        = "invalid postcode"
	invalidReference       = "invalid reference"
	invalidSite            = "invalid site"
	invalidAppointmentDate = "invalid appointment date"
	invalidAppointmentTime = "invalid appointment time"
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

	if err := mapAvailabilityErrorCodes(resp.ResponseCode, resp.ResponseMessage); err != nil {
		return nil, err
	}

	return &contract.GetAvailableSlotsResponse{
		Slots: slots,
	}, nil
}

func (lb LowriBeck) BookingRequest(id uint32, req *contract.CreateBookingRequest) (*lowribeck.CreateBookingRequest, error) {
	appDate, appTime, err := mapBookingSlot(req.GetSlot())
	if err != nil {
		return nil, err
	}

	return &lowribeck.CreateBookingRequest{
		PostCode:             req.GetPostcode(),
		ReferenceID:          req.GetReference(),
		AppointmentDate:      appDate,
		AppointmentTime:      appTime,
		Vulnerabilities:      mapVulnerabilities(req.GetVulnerabilityDetails()),
		VulnerabilitiesOther: req.GetVulnerabilityDetails().GetOther(),
		SiteContactName:      mapContactName(req.GetContactDetails()),
		SiteContactNumber:    req.GetContactDetails().GetPhone(),
		SendingSystem:        lb.sendingSystem,
		ReceivingSystem:      lb.receivingSystem,
		CreatedDate:          time.Now().UTC().Format(requestTimeFormat),
		// An ID sent to LB which they return in the response and can be used for debugging issues with them
		RequestID: fmt.Sprintf("%d", id),
	}, nil
}

func (lb LowriBeck) BookingResponse(resp *lowribeck.CreateBookingResponse) (*contract.CreateBookingResponse, error) {
	err := mapBookingResponseCodes(resp.ResponseCode, resp.ResponseMessage)
	if err != nil {
		return nil, err
	}
	return &contract.CreateBookingResponse{
		Success: true,
	}, nil
}

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

func mapAvailabilityErrorCodes(responseCode, responseMessage string) error {
	switch responseCode {
	case "":
		return nil
	//EA01 - No available slots for requested postcode
	case "EA01":
		return ErrAppointmentNotFound
		// EA02 - Unable to identify postcode
	case "EA02":
		return fmt.Errorf("%w [%s]", ErrInvalidRequest, invalidPostcode)
	case "EA03":
		// EA03 - Rearranging request sent outside agreed time parameter
		// EA03 - Booking request sent outside agreed time parameter
		switch responseMessage {
		case "Rearranging request sent outside agreed time parameter",
			"Booking request sent outside agreed time parameter":
			return ErrAppointmentOutOfRange
		// EA03 - Work Reference Invalid
		case "Work Reference Invalid":
			return fmt.Errorf("%w [%s]", ErrInvalidRequest, invalidReference)
		}
	}
	return fmt.Errorf("%w [%s]", ErrInternalError, responseMessage)
}

func mapBookingResponseCodes(responseCode, responseMessage string) error {
	switch responseCode {
	// B01 - Booking Confirmed
	// R01 - Reschedule Confirmed
	case "B01", "R01":
		return nil
	// B02 - Appointment not available
	// R02 - Appointment not available
	case "B02", "R02":
		return ErrAppointmentNotFound
	// B03 - Invalid Elec Job Type Code
	// B03 - Invalid Gas Job Type Code
	// B04 - Invalid MPAN
	// B05 - Invalid MPRN
	// R03 - Invalid Elec Job Type Code
	// R03 - Invalid Gas Job Type Code
	// R04 - Invalid MPAN
	// R05 - Invalid MPRN
	// We never send these fields
	case "B03", "B04", "B05", "R03", "R04", "R05":
		return fmt.Errorf("%w [%s]", ErrInvalidRequest, responseMessage)
	// B06 - Invalid Appt Date
	// B06 – Invalid Date Format
	// R06 - Invalid Appt Date
	// R06 – Invalid Date Format
	case "B06", "R06":
		return fmt.Errorf("%w [%s]", ErrInvalidRequest, invalidAppointmentDate)
	// R07 - Invalid Appt Time
	// B07 - Invalid Appt Time
	case "B07", "R07":
		return fmt.Errorf("%w [%s]", ErrInvalidRequest, invalidAppointmentTime)
	// B13 - Invalid Reference ID
	// R12 - Invalid Reference ID
	case "B13", "R12":
		return fmt.Errorf("%w [%s]", ErrInvalidRequest, invalidReference)
	// B08 - Duplicate Elec job exists
	// B08 - Duplicate Gas job exists
	// R08 - Duplicate Elec job exists
	// R08 - Duplicate Gas job exists
	case "B08", "R08":
		return ErrAppointmentAlreadyExists
	case "B09", "R09":
		switch responseMessage {
		// B09 - No available slots for requested postcode
		// R09 - No available slots for requested postcode
		case "No available slots for requested postcode":
			return ErrAppointmentNotFound
		// B09 - Rearranging request sent outside agreed time parameter
		// B09 - Booking request sent outside agreed time parameter
		case "Rearranging request sent outside agreed time parameter",
			"Booking request sent outside agreed time parameter":
			return ErrAppointmentOutOfRange
		// B09 - Site status not suitable for request
		// B09 - Not available as site is complete
		// B09 - The site is currently on hold
		// R09 – Site status not suitable for request
		// R09 - Not available as site is complete
		// R09 - The site is currently on hold
		case "Site status not suitable for request",
			"Not available as site is complete",
			"The site is currently on hold":
			return fmt.Errorf("%w [%s]", ErrInvalidRequest, invalidSite)
		// B09 - Post Code is missing or invalid
		// B09 - Postcode and Reference ID mismatch
		// R09 - Post Code is missing or invalid
		// R09 - Postcode and Reference ID mismatch
		case "Post Code is missing or invalid",
			"Postcode and Reference ID mismatch":
			return fmt.Errorf("%w [%s]", ErrInvalidRequest, invalidPostcode)
		// B09 - No Jobs found for Reference ID
		// R09 - No Jobs found for Reference ID
		case "No Jobs found for Reference ID":
			return fmt.Errorf("%w [%s]", ErrInvalidRequest, invalidReference)
		}
		// R11 – Rearranging request sent outside agreed time parameter
	case "R11":
		return ErrAppointmentOutOfRange
	}
	return fmt.Errorf("%w [%s]", ErrInternalError, responseMessage)
}

func mapBookingSlot(slot *contract.BookingSlot) (string, string, error) {
	if slot == nil {
		return "", "", fmt.Errorf("invalid booking slot")
	}
	slotDate := slot.GetDate()
	if slotDate == nil {
		return "", "", fmt.Errorf("invalid booking slot date")
	}
	appDate := fmt.Sprintf("%02d/%02d/%4d", slotDate.Day, slotDate.Month, slotDate.Year)
	appTime := fmt.Sprintf(appointmentTimeFormat, slot.StartTime, slot.EndTime)

	return appDate, appTime, nil
}

func mapVulnerabilities(vulnerabilities *contract.VulnerabilityDetails) string {
	vulnCodes := make([]string, len(vulnerabilities.GetVulnerabilities()))
	for i, vul := range vulnerabilities.GetVulnerabilities() {
		switch vul {
		// 01 - Hearing Impaired
		case contract.Vulnerability_VULNERABILITY_HEARING:
			vulnCodes[i] = "1"
		// 02 - Visually Impaired
		case contract.Vulnerability_VULNERABILITY_SIGHT:
			vulnCodes[i] = "2"
		// 03 - Elderly
		case contract.Vulnerability_VULNERABILITY_PENSIONABLE_AGE:
			vulnCodes[i] = "3"
		// 04 - Disabled
		case contract.Vulnerability_VULNERABILITY_LEARNING_DIFFICULTIES:
			vulnCodes[i] = "4"
		// 06 - Foreign Language Speaker
		case contract.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY:
			vulnCodes[i] = "6"
		// 07 - Restricted Movement
		case contract.Vulnerability_VULNERABILITY_PHYSICAL_OR_RESTRICTED_MOVEMENT:
			vulnCodes[i] = "7"
		// 08 - Serious Illness
		case contract.Vulnerability_VULNERABILITY_ILLNESS:
			vulnCodes[i] = "8"
		// 09 - Other
		case contract.Vulnerability_VULNERABILITY_OTHER,
			contract.Vulnerability_VULNERABILITY_UNKNOWN:
			vulnCodes[i] = "9"
		}
		// Unused LB codes
		// 05 - Electrical Medical Equipment
	}
	return strings.Join(vulnCodes, ",")
}

func mapContactName(contact *contract.ContactDetails) string {
	contactName := strings.TrimSpace(contact.GetTitle() + " " + contact.GetFirstName())
	return strings.TrimSpace(contactName + " " + contact.GetLastName())

}
