package mapper

import (
	"errors"
	"fmt"
	"strings"
	"time"

	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/lowribeck"
	"google.golang.org/genproto/googleapis/type/date"
)

var (
	ErrInvalidElectricityTariffType = errors.New("invalid electricity tariff type")
	ErrInvalidGasTariffType         = errors.New("invalid gas tariff type")
)

const (
	requestTimeFormat     = "02/01/2006 15:04:05"
	appointmentDateFormat = "02/01/2006"
	appointmentTimeFormat = "%d:00-%d:00"
)

type LowriBeck struct {
	sendingSystem   string
	receivingSystem string

	CreditElectricityJobType     string
	PrepaymentElectricityJobType string
	CreditGasJobType             string
	PrepaymentGasJobType         string
}

func NewLowriBeckMapper(sendingSystem, receivingSystem, creditElecJob, prepaymentElecJob, creditGasJob, prepaymentGasJob string) *LowriBeck {
	return &LowriBeck{
		sendingSystem:   sendingSystem,
		receivingSystem: receivingSystem,

		CreditElectricityJobType:     creditElecJob,
		PrepaymentElectricityJobType: prepaymentElecJob,
		CreditGasJobType:             creditGasJob,
		PrepaymentGasJobType:         prepaymentGasJob,
	}
}

func (lb LowriBeck) AvailabilityRequest(id uint32, req *lowribeckv1.GetAvailableSlotsRequest) *lowribeck.GetCalendarAvailabilityRequest {
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

func (lb LowriBeck) AvailabilityRequestPointOfSale(id uint32, req *lowribeckv1.GetAvailableSlotsPointOfSaleRequest) (*lowribeck.GetCalendarAvailabilityRequest, error) {

	elecJobTypeCode, gasJobTypeCode, err := lb.mapTariffTypeToJobType(req.GetElectricityTariffType(), req.GetGasTariffType())
	if err != nil {
		return nil, err
	}

	request := &lowribeck.GetCalendarAvailabilityRequest{
		PostCode:        req.GetPostcode(),
		Mpan:            req.GetMpan(),
		ElecJobTypeCode: elecJobTypeCode,
		SendingSystem:   lb.sendingSystem,
		ReceivingSystem: lb.receivingSystem,
		CreatedDate:     time.Now().UTC().Format(requestTimeFormat),
		// An ID sent to LB which they return in the response and can be used for debugging issues with them
		RequestID: fmt.Sprintf("%d", id),
	}

	if req.GetMprn() != "" {
		request.Mprn = req.GetMprn()
	}

	if req.GetGasTariffType() != lowribeckv1.TariffType_TARIFF_TYPE_UNKNOWN {
		request.GasJobTypeCode = gasJobTypeCode
	}

	return request, nil
}

func (lb LowriBeck) AvailableSlotsResponse(resp *lowribeck.GetCalendarAvailabilityResponse) (*lowribeckv1.GetAvailableSlotsResponse, error) {
	if err := mapAvailabilityErrorCodes(resp.ResponseCode, resp.ResponseMessage); err != nil {
		return nil, err
	}

	slots, err := mapAvailabilitySlots(resp.CalendarAvailabilityResult)
	if err != nil {
		return nil, err
	}

	return &lowribeckv1.GetAvailableSlotsResponse{
		Slots: slots,
	}, nil
}

func (lb LowriBeck) AvailableSlotsPointOfSaleResponse(resp *lowribeck.GetCalendarAvailabilityResponse) (*lowribeckv1.GetAvailableSlotsPointOfSaleResponse, error) {
	if err := mapAvailabilityErrorCodes(resp.ResponseCode, resp.ResponseMessage); err != nil {
		return nil, err
	}

	slots, err := mapAvailabilitySlots(resp.CalendarAvailabilityResult)
	if err != nil {
		return nil, err
	}

	return &lowribeckv1.GetAvailableSlotsPointOfSaleResponse{
		Slots: slots,
	}, nil
}

func (lb LowriBeck) BookingRequest(id uint32, req *lowribeckv1.CreateBookingRequest) (*lowribeck.CreateBookingRequest, error) {
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

func (lb LowriBeck) BookingRequestPointOfSale(id uint32, req *lowribeckv1.CreateBookingPointOfSaleRequest) (*lowribeck.CreateBookingRequest, error) {
	appDate, appTime, err := mapBookingSlot(req.GetSlot())
	if err != nil {
		return nil, err
	}

	elecJobTypeCode, gasJobTypeCode, err := lb.mapTariffTypeToJobType(req.GetElectricityTariffType(), req.GetGasTariffType())
	if err != nil {
		return nil, err
	}

	request := &lowribeck.CreateBookingRequest{
		SubBuildName:            req.SiteAddress.Paf.GetSubBuilding(),
		BuildingName:            req.SiteAddress.Paf.GetBuildingName(),
		DependThroughfare:       req.SiteAddress.Paf.GetDependentThoroughfare(),
		Throughfare:             req.SiteAddress.Paf.GetThoroughfare(),
		DoubleDependantLocality: req.SiteAddress.Paf.GetDoubleDependentLocality(),
		DependantLocality:       req.SiteAddress.Paf.GetDependentLocality(),
		PostTown:                req.SiteAddress.Paf.GetPostTown(),
		County:                  "", // There is no County in the PAF format
		PostCode:                req.SiteAddress.Paf.GetPostcode(),
		Mpan:                    req.GetMpan(),
		ElecJobTypeCode:         elecJobTypeCode,
		AppointmentDate:         appDate,
		AppointmentTime:         appTime,
		Vulnerabilities:         mapVulnerabilities(req.GetVulnerabilityDetails()),
		VulnerabilitiesOther:    req.GetVulnerabilityDetails().GetOther(),
		SiteContactName:         mapContactName(req.GetContactDetails()),
		SiteContactNumber:       req.GetContactDetails().GetPhone(),
		SendingSystem:           lb.sendingSystem,
		ReceivingSystem:         lb.receivingSystem,
		CreatedDate:             time.Now().UTC().Format(requestTimeFormat),
		// An ID sent to LB which they return in the response and can be used for debugging issues with them
		RequestID: fmt.Sprintf("%d", id),
	}

	if req.GetMprn() != "" {
		request.Mprn = req.GetMprn()
	}

	if req.GetGasTariffType() != lowribeckv1.TariffType_TARIFF_TYPE_UNKNOWN {
		request.GasJobTypeCode = gasJobTypeCode
	}

	return request, nil
}

func (lb LowriBeck) BookingResponse(resp *lowribeck.CreateBookingResponse) (*lowribeckv1.CreateBookingResponse, error) {
	err := mapBookingResponseCodes(resp.ResponseCode, resp.ResponseMessage)
	if err != nil {
		return nil, err
	}
	return &lowribeckv1.CreateBookingResponse{
		Success: true,
	}, nil
}

func (lb LowriBeck) BookingResponsePointOfSale(resp *lowribeck.CreateBookingResponse) (*lowribeckv1.CreateBookingPointOfSaleResponse, error) {
	err := mapBookingResponseCodes(resp.ResponseCode, resp.ResponseMessage)
	if err != nil {
		return nil, err
	}
	return &lowribeckv1.CreateBookingPointOfSaleResponse{
		Success:   true,
		Reference: resp.ReferenceID,
	}, nil
}

func (lb LowriBeck) UpdateContactDetailsRequest(id uint32, req *lowribeckv1.UpdateContactDetailsRequest) *lowribeck.UpdateContactDetailsRequest {
	return &lowribeck.UpdateContactDetailsRequest{
		ReferenceID:          req.GetReference(),
		Vulnerabilities:      mapVulnerabilities(req.GetVulnerabilityDetails()),
		VulnerabilitiesOther: req.GetVulnerabilityDetails().GetOther(),
		SiteContactName:      mapContactName(req.GetContactDetails()),
		SiteContactNumber:    req.GetContactDetails().GetPhone(),
		SendingSystem:        lb.sendingSystem,
		ReceivingSystem:      lb.receivingSystem,
		CreatedDate:          time.Now().UTC().Format(requestTimeFormat),
		// An ID sent to LB which they return in the response and can be used for debugging issues with them
		RequestID: fmt.Sprintf("%d", id),
	}
}

func (lb LowriBeck) UpdateContactDetailsResponse(resp *lowribeck.UpdateContactDetailsResponse) (*lowribeckv1.UpdateContactDetailsResponse, error) {
	err := mapUpdateContactDetailsResponseCodes(resp.ResponseCode, resp.ResponseMessage)
	if err != nil {
		return nil, err
	}
	return &lowribeckv1.UpdateContactDetailsResponse{
		Success: true,
	}, nil
}

func mapAvailabilitySlots(availabilityResults []lowribeck.AvailabilitySlot) ([]*lowribeckv1.BookingSlot, error) {
	var err error
	slots := make([]*lowribeckv1.BookingSlot, len(availabilityResults))
	for i, res := range availabilityResults {
		slot := &lowribeckv1.BookingSlot{}
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
		return NewInvalidRequestError(InvalidPostcode)
	case "EA03":
		// EA03 - Rearranging request sent outside agreed time parameter
		// EA03 - Booking request sent outside agreed time parameter
		switch responseMessage {
		case "Rearranging request sent outside agreed time parameter",
			"Booking request sent outside agreed time parameter":
			return ErrAppointmentOutOfRange
		// EA03 - Postcode mismatch
		case "Postcode and Reference ID mismatch", // error in spec
			"Postcode mismatch": // error seen
			return NewInvalidRequestError(InvalidPostcode)
		// EA03 - Work Reference Invalid
		case "Work Reference Invalid", // error in spec
			"Invalid Reference ID": // error seen
			return NewInvalidRequestError(InvalidReference)
		// EA03 - Invalid Job/Sub Job Code
		case "Invalid Job/Sub Job Code":
			return ErrInvalidJobTypeCode
		case "Insufficient notice to rearrange this appointment.":
			return fmt.Errorf("%w [%s]", ErrInternalError, responseMessage)
		}
	}
	return fmt.Errorf("%w [%s]", ErrUnknownError, responseMessage)
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
		// R03 - Invalid Elec Job Type Code
		// R03 - Invalid Gas Job Type Code
	case "B03", "R03":
		switch responseMessage {
		case "Invalid Elec Job Type Code":
			return ErrInvalidElectricityJobTypeCode
		case "Invalid Gas Job Type Code":
			return ErrInvalidGasJobTypeCode
		}
		// B04 - Invalid MPAN
		// R04 - Invalid MPAN
	case "B04", "R04":
		return NewInvalidRequestError(InvalidMPAN)
	// B05 - Invalid MPRN
	// R05 - Invalid MPRN
	case "B05", "R05":
		return NewInvalidRequestError(InvalidMPRN)
	// B06 - Invalid Appt Date
	// B06 – Invalid Date Format
	// R06 - Invalid Appt Date
	// R06 – Invalid Date Format
	case "B06", "R06":
		return NewInvalidRequestError(InvalidAppointmentDate)
	// R07 - Invalid Appt Time
	// B07 - Invalid Appt Time
	case "B07", "R07":
		return NewInvalidRequestError(InvalidAppointmentTime)
	// B13 - Invalid Reference ID
	// R12 - Invalid Reference ID
	case "B13", "R12":
		return NewInvalidRequestError(InvalidReference)
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
			return NewInvalidRequestError(InvalidSite)
		// B09 - Post Code is missing or invalid
		// B09 - Postcode and Reference ID mismatch
		// R09 - Post Code is missing or invalid
		// R09 - Postcode and Reference ID mismatch
		case "Post Code is missing or invalid",
			"Postcode and Reference ID mismatch", // error in spec
			"Postcode mismatch":                  // error seen
			return NewInvalidRequestError(InvalidPostcode)
		// B09 - No Jobs found for Reference ID
		// R09 - No Jobs found for Reference ID
		case "No Jobs found for Reference ID":
			return NewInvalidRequestError(InvalidReference)
		}
		// R11 – Rearranging request sent outside agreed time parameter
	case "R11":
		return ErrAppointmentOutOfRange
	}
	return fmt.Errorf("%w [%s]", ErrUnknownError, responseMessage)
}

func mapUpdateContactDetailsResponseCodes(responseCode, responseMessage string) error {
	switch responseCode {
	// U01 - Update confirmed
	case "U01":
		return nil
	// U02 - Invalid Reference ID
	case "U02":
		return NewInvalidRequestError(InvalidReference)
	// U03 - Invalid MPAN
	case "U03":
		return NewInvalidRequestError(InvalidMPAN)
	// U04 - Invalid MPRN
	case "U04":
		return NewInvalidRequestError(InvalidMPRN)
	case "U05":
		switch responseMessage {
		// U05 - No jobs found
		case "No jobs found":
			return ErrAppointmentNotFound
		// U05 - Update sent for an appointment in the past
		case "Update sent for an appointment in the past":
			return fmt.Errorf("%w [%s]", ErrInternalError, responseMessage)
		}
	// U06 - Request sent after update deadline
	case "B06", "R06":
		return NewInvalidRequestError(InvalidAppointmentDate)

	}
	return fmt.Errorf("%w [%s]", ErrUnknownError, responseMessage)
}

func mapBookingSlot(slot *lowribeckv1.BookingSlot) (string, string, error) {
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

func mapVulnerabilities(vulnerabilities *lowribeckv1.VulnerabilityDetails) string {
	vulnCodes := make([]string, len(vulnerabilities.GetVulnerabilities()))
	for i, vul := range vulnerabilities.GetVulnerabilities() {
		switch vul {
		// 01 - Hearing Impaired
		case lowribeckv1.Vulnerability_VULNERABILITY_HEARING:
			vulnCodes[i] = "1"
		// 02 - Visually Impaired
		case lowribeckv1.Vulnerability_VULNERABILITY_SIGHT:
			vulnCodes[i] = "2"
		// 03 - Elderly
		case lowribeckv1.Vulnerability_VULNERABILITY_PENSIONABLE_AGE:
			vulnCodes[i] = "3"
		// 04 - Disabled
		case lowribeckv1.Vulnerability_VULNERABILITY_LEARNING_DIFFICULTIES:
			vulnCodes[i] = "4"
		// 06 - Foreign Language Speaker
		case lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY:
			vulnCodes[i] = "6"
		// 07 - Restricted Movement
		case lowribeckv1.Vulnerability_VULNERABILITY_PHYSICAL_OR_RESTRICTED_MOVEMENT:
			vulnCodes[i] = "7"
		// 08 - Serious Illness
		case lowribeckv1.Vulnerability_VULNERABILITY_ILLNESS:
			vulnCodes[i] = "8"
		// 09 - Other
		case lowribeckv1.Vulnerability_VULNERABILITY_OTHER,
			lowribeckv1.Vulnerability_VULNERABILITY_UNKNOWN:
			vulnCodes[i] = "9"
		}
		// Unused LB codes
		// 05 - Electrical Medical Equipment
	}
	return strings.Join(vulnCodes, ",")
}

func mapContactName(contact *lowribeckv1.ContactDetails) string {
	contactName := strings.TrimSpace(contact.GetTitle() + " " + contact.GetFirstName())
	return strings.TrimSpace(contactName + " " + contact.GetLastName())

}

func (lb LowriBeck) mapTariffTypeToJobType(elecTariffType, gasTariffType lowribeckv1.TariffType) (elecJobTypeCode string, gasJobTypeCode string, err error) {

	switch elecTariffType {
	case lowribeckv1.TariffType_TARIFF_TYPE_CREDIT:
		elecJobTypeCode = lb.CreditElectricityJobType
	case lowribeckv1.TariffType_TARIFF_TYPE_PREPAYMENT:
		elecJobTypeCode = lb.PrepaymentElectricityJobType
	default:
		return "", "", ErrInvalidElectricityTariffType
	}

	switch gasTariffType {
	case lowribeckv1.TariffType_TARIFF_TYPE_CREDIT:
		gasJobTypeCode = lb.CreditGasJobType
	case lowribeckv1.TariffType_TARIFF_TYPE_PREPAYMENT:
		gasJobTypeCode = lb.PrepaymentGasJobType
	case lowribeckv1.TariffType_TARIFF_TYPE_UNKNOWN:
		gasJobTypeCode = ""
	default:
		return "", "", ErrInvalidGasTariffType
	}

	return elecJobTypeCode, gasJobTypeCode, nil
}
