package mapper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	contract "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/lowribeck"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/mapper"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/protobuf/testing/protocmp"
)

const requestTimeFormat = "02/01/2006 15:04:05"

func TestMapAvailableSlotsResponse(t *testing.T) {
	testCases := []struct {
		desc          string
		lb            *lowribeck.GetCalendarAvailabilityResponse
		expected      *contract.GetAvailableSlotsResponse
		expectedError error
	}{
		{
			desc: "Valid",
			lb: &lowribeck.GetCalendarAvailabilityResponse{
				CalendarAvailabilityResult: []lowribeck.AvailabilitySlot{
					{
						AppointmentDate: "01/12/2023",
						AppointmentTime: "10:00-12:00",
					},
					{
						AppointmentDate: "10/08/2023",
						AppointmentTime: "14:00-16:00",
					},
				},
			},
			expected: &contract.GetAvailableSlotsResponse{
				Slots: []*contract.BookingSlot{
					{
						Date: &date.Date{
							Day:   1,
							Month: 12,
							Year:  2023,
						},
						StartTime: 10,
						EndTime:   12,
					},
					{
						Date: &date.Date{
							Day:   10,
							Month: 8,
							Year:  2023,
						},
						StartTime: 14,
						EndTime:   16,
					},
				},
			},
		},
		{
			desc: "Invalid appointment date",
			lb: &lowribeck.GetCalendarAvailabilityResponse{
				CalendarAvailabilityResult: []lowribeck.AvailabilitySlot{
					{
						AppointmentDate: "01/13/2023",
						AppointmentTime: "10:00-12:00",
					},
				},
			},
			expectedError: fmt.Errorf("error converting appointment date: parsing time \"01/13/2023\": month out of range"),
		},
		{
			desc: "Invalid appointment time",
			lb: &lowribeck.GetCalendarAvailabilityResponse{
				CalendarAvailabilityResult: []lowribeck.AvailabilitySlot{
					{
						AppointmentDate: "01/12/2023",
						AppointmentTime: "blah",
					},
				},
			},
			expectedError: fmt.Errorf("error converting appointment time: expected integer"),
		},
		{
			desc: "Invalid appointment end time",
			lb: &lowribeck.GetCalendarAvailabilityResponse{
				CalendarAvailabilityResult: []lowribeck.AvailabilitySlot{
					{
						AppointmentDate: "01/12/2023",
						AppointmentTime: "23:00-24:00",
					},
				},
			},
			expectedError: fmt.Errorf("error converting appointment time: invalid end time: \"23:00-24:00\""),
		},
		{
			desc: "Invalid appointment start and end times",
			lb: &lowribeck.GetCalendarAvailabilityResponse{
				CalendarAvailabilityResult: []lowribeck.AvailabilitySlot{
					{
						AppointmentDate: "01/12/2023",
						AppointmentTime: "22:00-21:00",
					},
				},
			},
			expectedError: fmt.Errorf("error converting appointment time: invalid appointment time: \"22:00-21:00\""),
		},
		{
			desc: "Failed - no slots response",
			lb: &lowribeck.GetCalendarAvailabilityResponse{
				ResponseCode:    "EA01",
				ResponseMessage: "No available slots for requested postcode",
			},
			expectedError: fmt.Errorf("no appointments found"),
		},
		{
			desc: "Failed - invalid request response 1",
			lb: &lowribeck.GetCalendarAvailabilityResponse{
				ResponseCode:    "EA02",
				ResponseMessage: "Unable to identify postcode",
			},
			expectedError: fmt.Errorf("invalid request [postcode]"),
		},
		{
			desc: "Failed - invalid request response 2",
			lb: &lowribeck.GetCalendarAvailabilityResponse{
				ResponseCode:    "EA03",
				ResponseMessage: "Postcode mismatch",
			},
			expectedError: fmt.Errorf("invalid request [postcode]"),
		},
		{
			desc: "Failed - invalid request response 3",
			lb: &lowribeck.GetCalendarAvailabilityResponse{
				ResponseCode:    "EA03",
				ResponseMessage: "Insufficient notice to rearrange this appointment.",
			},
			expectedError: fmt.Errorf("internal server error [Insufficient notice to rearrange this appointment.]"),
		},
		{
			desc: "Failed - generic response",
			lb: &lowribeck.GetCalendarAvailabilityResponse{
				// According to the docs this code doesn't exist, but I've seen it (a lot)
				ResponseCode:    "EC04",
				ResponseMessage: "Generic LB error for failed to process request",
			},
			expectedError: fmt.Errorf("unknown error [Generic LB error for failed to process request]"),
		},
	}

	assert := assert.New(t)
	lbMapper := mapper.NewLowriBeckMapper("sendingSystem", "receivingSystem", "", "", "", "")

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			res, err := lbMapper.AvailableSlotsResponse(tc.lb)
			if tc.expectedError == nil {
				assert.NoError(err, tc.desc)
				diff := cmp.Diff(tc.expected, res, protocmp.Transform(), cmpopts.IgnoreUnexported())
				assert.Empty(diff, tc.desc)
			} else {
				assert.EqualError(err, tc.expectedError.Error(), tc.desc)
			}
		})
	}
}

func TestMapBookingRequest(t *testing.T) {
	testCases := []struct {
		desc          string
		lb            *contract.CreateBookingRequest
		expected      *lowribeck.CreateBookingRequest
		expectedError error
	}{
		{
			desc: "Valid",
			lb: &contract.CreateBookingRequest{
				Postcode:  "postcode",
				Reference: "reference",
				Slot: &contract.BookingSlot{
					Date: &date.Date{
						Day:   1,
						Month: 12,
						Year:  2023,
					},
					StartTime: 10,
					EndTime:   12,
				},
				VulnerabilityDetails: &contract.VulnerabilityDetails{
					Vulnerabilities: []contract.Vulnerability{
						contract.Vulnerability_VULNERABILITY_HEARING,
						contract.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
					},
					Other: "other",
				},
				ContactDetails: &contract.ContactDetails{
					FirstName: "Home",
					LastName:  "Alone",
					Phone:     "tel",
				},
			},
			expected: &lowribeck.CreateBookingRequest{
				RequestID:            "0",
				PostCode:             "postcode",
				ReferenceID:          "reference",
				AppointmentDate:      "01/12/2023",
				AppointmentTime:      "10:00-12:00",
				SiteContactName:      "Home Alone",
				SiteContactNumber:    "tel",
				SendingSystem:        "sendingSystem",
				ReceivingSystem:      "receivingSystem",
				Vulnerabilities:      "1,6",
				VulnerabilitiesOther: "other",
				CreatedDate:          time.Now().UTC().Format(requestTimeFormat),
			},
		},
		{
			desc:          "Empty appointment slot",
			lb:            &contract.CreateBookingRequest{},
			expectedError: fmt.Errorf("invalid booking slot"),
		},
		{
			desc: "Empty appointment date",
			lb: &contract.CreateBookingRequest{
				Slot: &contract.BookingSlot{
					StartTime: 10,
					EndTime:   12,
				},
			},
			expectedError: fmt.Errorf("invalid booking slot date"),
		},
	}

	assert := assert.New(t)
	lbMapper := mapper.NewLowriBeckMapper("sendingSystem", "receivingSystem", "", "", "", "")

	for i, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			res, err := lbMapper.BookingRequest(uint32(i), tc.lb)
			if tc.expectedError == nil {
				assert.NoError(err, tc.desc)
				diff := cmp.Diff(tc.expected, res, protocmp.Transform(), cmpopts.IgnoreUnexported(), cmpopts.EquateApproxTime(time.Second))
				assert.Empty(diff, tc.desc)
			} else {
				assert.EqualError(err, tc.expectedError.Error(), tc.desc)
			}
		})
	}
}

func TestMapBookingResponse(t *testing.T) {
	testCases := []struct {
		desc          string
		lb            *lowribeck.CreateBookingResponse
		expected      *contract.CreateBookingResponse
		expectedError error
	}{
		{
			desc: "Success",
			lb: &lowribeck.CreateBookingResponse{
				ResponseCode:    "B01",
				ResponseMessage: "Booking Confirmed",
			},
			expected: &contract.CreateBookingResponse{
				Success: true,
			},
		},
		{
			desc: "Appointment not available",
			lb: &lowribeck.CreateBookingResponse{
				ResponseCode:    "B02",
				ResponseMessage: "Appointment not available",
			},
			expected: &contract.CreateBookingResponse{
				Success: false,
			},
			expectedError: fmt.Errorf("no appointments found"),
		},
		{
			desc: "Invalid Appointment Time",
			lb: &lowribeck.CreateBookingResponse{
				ResponseCode:    "B07",
				ResponseMessage: "Invalid Appt Time",
			},
			expected: &contract.CreateBookingResponse{
				Success: false,
			},
			expectedError: fmt.Errorf("invalid request [appointment time]"),
		},
		{
			desc: "Duplicate Elec job exists",
			lb: &lowribeck.CreateBookingResponse{
				ResponseCode:    "B08",
				ResponseMessage: "Duplicate Elec job exists",
			},
			expected: &contract.CreateBookingResponse{
				Success: false,
			},
			expectedError: fmt.Errorf("appointment already exists"),
		},
		{
			desc: "No available slots for requested postcode",
			lb: &lowribeck.CreateBookingResponse{
				ResponseCode:    "B09",
				ResponseMessage: "No available slots for requested postcode",
			},
			expected: &contract.CreateBookingResponse{
				Success: false,
			},
			expectedError: fmt.Errorf("no appointments found"),
		},
		{
			desc: "Site status not suitable for request",
			lb: &lowribeck.CreateBookingResponse{
				ResponseCode:    "B09",
				ResponseMessage: "Site status not suitable for request",
			},
			expected: &contract.CreateBookingResponse{
				Success: false,
			},
			expectedError: fmt.Errorf("invalid request [site]"),
		},
		{
			desc: "Post Code is missing or invalid",
			lb: &lowribeck.CreateBookingResponse{
				ResponseCode:    "B09",
				ResponseMessage: "Post Code is missing or invalid",
			},
			expected: &contract.CreateBookingResponse{
				Success: false,
			},
			expectedError: fmt.Errorf("invalid request [postcode]"),
		},
	}

	assert := assert.New(t)
	lbMapper := mapper.NewLowriBeckMapper("sendingSystem", "receivingSystem", "", "", "", "")

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			res, err := lbMapper.BookingResponse(tc.lb)
			if tc.expectedError == nil {
				assert.NoError(err, tc.desc)
				diff := cmp.Diff(tc.expected, res, protocmp.Transform(), cmpopts.IgnoreUnexported())
				assert.Empty(diff, tc.desc)
			} else {
				assert.EqualError(err, tc.expectedError.Error(), tc.desc)
			}
		})
	}
}

func TestMapAvailableSlotsPointOfSaleResponse(t *testing.T) {

	type inputParams struct {
		id  uint32
		req *contract.GetAvailableSlotsPointOfSaleRequest
	}

	testCases := []struct {
		desc          string
		input         inputParams
		expected      *lowribeck.GetCalendarAvailabilityRequest
		expectedError error
	}{
		{
			desc: "Success - Electricity Credit - No MPRN",
			input: inputParams{
				id: 1,
				req: &contract.GetAvailableSlotsPointOfSaleRequest{
					Postcode:              "ZE 11",
					Mpan:                  "mpan-1",
					Mprn:                  nil,
					ElectricityTariffType: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
					GasTariffType:         nil,
				},
			},
			expected: &lowribeck.GetCalendarAvailabilityRequest{
				RequestID:       "1",
				SendingSystem:   "sendingSystem",
				ReceivingSystem: "receivingSystem",
				PostCode:        "ZE 11",
				Mpan:            "mpan-1",
				Mprn:            "",
				ElecJobTypeCode: "crElec",
				GasJobTypeCode:  "",
				CreatedDate:     time.Now().UTC().Format(requestTimeFormat),
			},
		},
		{
			desc: "Success - Electricity Prepayment - No MPRN",
			input: inputParams{
				id: 1,
				req: &contract.GetAvailableSlotsPointOfSaleRequest{
					Postcode:              "ZE 11",
					Mpan:                  "mpan-1",
					Mprn:                  nil,
					ElectricityTariffType: bookingv1.TariffType_TARIFF_TYPE_PREPAYMENT,
					GasTariffType:         nil,
				},
			},
			expected: &lowribeck.GetCalendarAvailabilityRequest{
				RequestID:       "1",
				SendingSystem:   "sendingSystem",
				ReceivingSystem: "receivingSystem",
				PostCode:        "ZE 11",
				Mpan:            "mpan-1",
				Mprn:            "",
				ElecJobTypeCode: "ppmElec",
				GasJobTypeCode:  "",
				CreatedDate:     time.Now().UTC().Format(requestTimeFormat),
			},
		},
		{
			desc: "Success - Electricity Credit - Credit Gas",
			input: inputParams{
				id: 1,
				req: &contract.GetAvailableSlotsPointOfSaleRequest{
					Postcode:              "ZE 11",
					Mpan:                  "mpan-1",
					Mprn:                  strToPtr("mprn-1"),
					ElectricityTariffType: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
					GasTariffType:         tariffTypeToPtr(bookingv1.TariffType_TARIFF_TYPE_CREDIT),
				},
			},
			expected: &lowribeck.GetCalendarAvailabilityRequest{
				RequestID:       "1",
				SendingSystem:   "sendingSystem",
				ReceivingSystem: "receivingSystem",
				PostCode:        "ZE 11",
				Mpan:            "mpan-1",
				Mprn:            "mprn-1",
				ElecJobTypeCode: "crElec",
				GasJobTypeCode:  "crGas",
				CreatedDate:     time.Now().UTC().Format(requestTimeFormat),
			},
		},
		{
			desc: "Success - Electricity Credit - Credit Gas",
			input: inputParams{
				id: 1,
				req: &contract.GetAvailableSlotsPointOfSaleRequest{
					Postcode:              "ZE 11",
					Mpan:                  "mpan-1",
					Mprn:                  strToPtr("mprn-1"),
					ElectricityTariffType: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
					GasTariffType:         tariffTypeToPtr(bookingv1.TariffType_TARIFF_TYPE_PREPAYMENT),
				},
			},
			expected: &lowribeck.GetCalendarAvailabilityRequest{
				RequestID:       "1",
				SendingSystem:   "sendingSystem",
				ReceivingSystem: "receivingSystem",
				PostCode:        "ZE 11",
				Mpan:            "mpan-1",
				Mprn:            "mprn-1",
				ElecJobTypeCode: "crElec",
				GasJobTypeCode:  "ppmGas",
				CreatedDate:     time.Now().UTC().Format(requestTimeFormat),
			},
		},
	}

	assert := assert.New(t)
	lbMapper := mapper.NewLowriBeckMapper("sendingSystem", "receivingSystem", "crElec", "ppmElec", "crGas", "ppmGas")

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			res := lbMapper.AvailabilityRequestPointOfSale(tc.input.id, tc.input.req)
			diff := cmp.Diff(tc.expected, res, protocmp.Transform(), cmpopts.IgnoreUnexported(), cmpopts.EquateApproxTime(time.Second))
			assert.Empty(diff, tc.desc)
		})
	}
}

func TestMapBookingPointOfSaleResponse(t *testing.T) {

	type inputParams struct {
		id  uint32
		req *contract.CreateBookingPointOfSaleRequest
	}

	testCases := []struct {
		desc          string
		input         inputParams
		expected      *lowribeck.CreateBookingRequest
		expectedError error
	}{
		{
			desc: "Success - Electricity Credit - No MPRN",
			input: inputParams{
				id: 1,
				req: &contract.CreateBookingPointOfSaleRequest{
					Postcode:              "ZE 11",
					Mpan:                  "mpan-1",
					Mprn:                  nil,
					ElectricityTariffType: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
					GasTariffType:         nil,
					Slot: &contract.BookingSlot{
						Date: &date.Date{
							Year:  2020,
							Month: 12,
							Day:   13,
						},
						StartTime: 10,
						EndTime:   12,
					},
					VulnerabilityDetails: &contract.VulnerabilityDetails{
						Vulnerabilities: []contract.Vulnerability{
							contract.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
						},
						Other: "Other Vuln",
					},
					ContactDetails: &contract.ContactDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Doe",
						Phone:     "2002-2001",
					},
				},
			},
			expected: &lowribeck.CreateBookingRequest{
				RequestID:            "1",
				SendingSystem:        "sendingSystem",
				ReceivingSystem:      "receivingSystem",
				AppointmentDate:      "13/12/2020",
				AppointmentTime:      "10:00-12:00",
				PostCode:             "ZE 11",
				Mpan:                 "mpan-1",
				Mprn:                 "",
				ElecJobTypeCode:      "crElec",
				GasJobTypeCode:       "",
				Vulnerabilities:      "6",
				VulnerabilitiesOther: "Other Vuln",
				SiteContactName:      "Mr John Doe",
				SiteContactNumber:    "2002-2001",
				CreatedDate:          time.Now().UTC().Format(requestTimeFormat),
			},
		},
		{
			desc: "Success - Electricity Prepayment - No MPRN",
			input: inputParams{
				id: 1,
				req: &contract.CreateBookingPointOfSaleRequest{
					Postcode:              "ZE 11",
					Mpan:                  "mpan-1",
					Mprn:                  nil,
					ElectricityTariffType: bookingv1.TariffType_TARIFF_TYPE_PREPAYMENT,
					GasTariffType:         nil,
					Slot: &contract.BookingSlot{
						Date: &date.Date{
							Year:  2020,
							Month: 12,
							Day:   13,
						},
						StartTime: 10,
						EndTime:   12,
					},
					VulnerabilityDetails: &contract.VulnerabilityDetails{
						Vulnerabilities: []contract.Vulnerability{
							contract.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
						},
						Other: "Other Vuln",
					},
					ContactDetails: &contract.ContactDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Doe",
						Phone:     "2002-2001",
					},
				},
			},
			expected: &lowribeck.CreateBookingRequest{
				RequestID:            "1",
				SendingSystem:        "sendingSystem",
				ReceivingSystem:      "receivingSystem",
				AppointmentDate:      "13/12/2020",
				AppointmentTime:      "10:00-12:00",
				PostCode:             "ZE 11",
				Mpan:                 "mpan-1",
				Mprn:                 "",
				ElecJobTypeCode:      "ppmElec",
				GasJobTypeCode:       "",
				Vulnerabilities:      "6",
				VulnerabilitiesOther: "Other Vuln",
				SiteContactName:      "Mr John Doe",
				SiteContactNumber:    "2002-2001",
				CreatedDate:          time.Now().UTC().Format(requestTimeFormat),
			},
		},
		{
			desc: "Success - Electricity Credit - Gas Credit",
			input: inputParams{
				id: 1,
				req: &contract.CreateBookingPointOfSaleRequest{
					Postcode:              "ZE 11",
					Mpan:                  "mpan-1",
					Mprn:                  strToPtr("mprn-1"),
					ElectricityTariffType: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
					GasTariffType:         tariffTypeToPtr(bookingv1.TariffType_TARIFF_TYPE_CREDIT),
					Slot: &contract.BookingSlot{
						Date: &date.Date{
							Year:  2020,
							Month: 12,
							Day:   13,
						},
						StartTime: 10,
						EndTime:   12,
					},
					VulnerabilityDetails: &contract.VulnerabilityDetails{
						Vulnerabilities: []contract.Vulnerability{
							contract.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
						},
						Other: "Other Vuln",
					},
					ContactDetails: &contract.ContactDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Doe",
						Phone:     "2002-2001",
					},
				},
			},
			expected: &lowribeck.CreateBookingRequest{
				RequestID:            "1",
				SendingSystem:        "sendingSystem",
				ReceivingSystem:      "receivingSystem",
				AppointmentDate:      "13/12/2020",
				AppointmentTime:      "10:00-12:00",
				PostCode:             "ZE 11",
				Mpan:                 "mpan-1",
				Mprn:                 "mprn-1",
				ElecJobTypeCode:      "crElec",
				GasJobTypeCode:       "crGas",
				Vulnerabilities:      "6",
				VulnerabilitiesOther: "Other Vuln",
				SiteContactName:      "Mr John Doe",
				SiteContactNumber:    "2002-2001",
				CreatedDate:          time.Now().UTC().Format(requestTimeFormat),
			},
		},
		{
			desc: "Success - Electricity Credit - Prepayment Credit",
			input: inputParams{
				id: 1,
				req: &contract.CreateBookingPointOfSaleRequest{
					Postcode:              "ZE 11",
					Mpan:                  "mpan-1",
					Mprn:                  strToPtr("mprn-1"),
					ElectricityTariffType: bookingv1.TariffType_TARIFF_TYPE_CREDIT,
					GasTariffType:         tariffTypeToPtr(bookingv1.TariffType_TARIFF_TYPE_PREPAYMENT),
					Slot: &contract.BookingSlot{
						Date: &date.Date{
							Year:  2020,
							Month: 12,
							Day:   13,
						},
						StartTime: 10,
						EndTime:   12,
					},
					VulnerabilityDetails: &contract.VulnerabilityDetails{
						Vulnerabilities: []contract.Vulnerability{
							contract.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
						},
						Other: "Other Vuln",
					},
					ContactDetails: &contract.ContactDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Doe",
						Phone:     "2002-2001",
					},
				},
			},
			expected: &lowribeck.CreateBookingRequest{
				RequestID:            "1",
				SendingSystem:        "sendingSystem",
				ReceivingSystem:      "receivingSystem",
				AppointmentDate:      "13/12/2020",
				AppointmentTime:      "10:00-12:00",
				PostCode:             "ZE 11",
				Mpan:                 "mpan-1",
				Mprn:                 "mprn-1",
				ElecJobTypeCode:      "crElec",
				GasJobTypeCode:       "ppmGas",
				Vulnerabilities:      "6",
				VulnerabilitiesOther: "Other Vuln",
				SiteContactName:      "Mr John Doe",
				SiteContactNumber:    "2002-2001",
				CreatedDate:          time.Now().UTC().Format(requestTimeFormat),
			},
		},
	}

	assert := assert.New(t)
	lbMapper := mapper.NewLowriBeckMapper("sendingSystem", "receivingSystem", "crElec", "ppmElec", "crGas", "ppmGas")

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			res, err := lbMapper.BookingRequestPointOfSale(tc.input.id, tc.input.req)
			if tc.expectedError == nil {
				assert.NoError(err, tc.desc)
				diff := cmp.Diff(tc.expected, res, protocmp.Transform(), cmpopts.IgnoreUnexported(), cmpopts.EquateApproxTime(time.Second))
				assert.Empty(diff, tc.desc)
			} else {
				assert.EqualError(err, tc.expectedError.Error(), tc.desc)
			}
		})
	}
}

func strToPtr(s string) *string {
	return &s
}

func tariffTypeToPtr(t bookingv1.TariffType) *bookingv1.TariffType {
	return &t
}
