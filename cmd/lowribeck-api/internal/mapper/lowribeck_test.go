package mapper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	addressv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/energy_entities/address/v1"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
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
		expected      *lowribeckv1.GetAvailableSlotsResponse
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
			expected: &lowribeckv1.GetAvailableSlotsResponse{
				Slots: []*lowribeckv1.BookingSlot{
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
		t.Run(tc.desc, func(_ *testing.T) {
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
		lb            *lowribeckv1.CreateBookingRequest
		expected      *lowribeck.CreateBookingRequest
		expectedError error
	}{
		{
			desc: "Valid",
			lb: &lowribeckv1.CreateBookingRequest{
				Postcode:  "postcode",
				Reference: "reference",
				Slot: &lowribeckv1.BookingSlot{
					Date: &date.Date{
						Day:   1,
						Month: 12,
						Year:  2023,
					},
					StartTime: 10,
					EndTime:   12,
				},
				VulnerabilityDetails: &lowribeckv1.VulnerabilityDetails{
					Vulnerabilities: []lowribeckv1.Vulnerability{
						lowribeckv1.Vulnerability_VULNERABILITY_HEARING,
						lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
					},
					Other: "other",
				},
				ContactDetails: &lowribeckv1.ContactDetails{
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
			lb:            &lowribeckv1.CreateBookingRequest{},
			expectedError: fmt.Errorf("invalid booking slot"),
		},
		{
			desc: "Empty appointment date",
			lb: &lowribeckv1.CreateBookingRequest{
				Slot: &lowribeckv1.BookingSlot{
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
		t.Run(tc.desc, func(_ *testing.T) {
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
		expected      *lowribeckv1.CreateBookingResponse
		expectedError error
	}{
		{
			desc: "Success",
			lb: &lowribeck.CreateBookingResponse{
				ResponseCode:    "B01",
				ResponseMessage: "Booking Confirmed",
			},
			expected: &lowribeckv1.CreateBookingResponse{
				Success: true,
			},
		},
		{
			desc: "Appointment not available",
			lb: &lowribeck.CreateBookingResponse{
				ResponseCode:    "B02",
				ResponseMessage: "Appointment not available",
			},
			expected: &lowribeckv1.CreateBookingResponse{
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
			expected: &lowribeckv1.CreateBookingResponse{
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
			expected: &lowribeckv1.CreateBookingResponse{
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
			expected: &lowribeckv1.CreateBookingResponse{
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
			expected: &lowribeckv1.CreateBookingResponse{
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
			expected: &lowribeckv1.CreateBookingResponse{
				Success: false,
			},
			expectedError: fmt.Errorf("invalid request [postcode]"),
		},
	}

	assert := assert.New(t)
	lbMapper := mapper.NewLowriBeckMapper("sendingSystem", "receivingSystem", "", "", "", "")

	for _, tc := range testCases {
		t.Run(tc.desc, func(_ *testing.T) {
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
		req *lowribeckv1.GetAvailableSlotsPointOfSaleRequest
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
				req: &lowribeckv1.GetAvailableSlotsPointOfSaleRequest{
					Postcode:              "ZE 11",
					Mpan:                  "mpan-1",
					Mprn:                  "",
					ElectricityTariffType: lowribeckv1.TariffType_TARIFF_TYPE_CREDIT,
					GasTariffType:         lowribeckv1.TariffType_TARIFF_TYPE_UNKNOWN,
				},
			},
			expectedError: nil,
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
				req: &lowribeckv1.GetAvailableSlotsPointOfSaleRequest{
					Postcode:              "ZE 11",
					Mpan:                  "mpan-1",
					Mprn:                  "",
					ElectricityTariffType: lowribeckv1.TariffType_TARIFF_TYPE_PREPAYMENT,
					GasTariffType:         lowribeckv1.TariffType_TARIFF_TYPE_UNKNOWN,
				},
			},
			expectedError: nil,
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
				req: &lowribeckv1.GetAvailableSlotsPointOfSaleRequest{
					Postcode:              "ZE 11",
					Mpan:                  "mpan-1",
					Mprn:                  "mprn-1",
					ElectricityTariffType: lowribeckv1.TariffType_TARIFF_TYPE_CREDIT,
					GasTariffType:         lowribeckv1.TariffType_TARIFF_TYPE_CREDIT,
				},
			},
			expectedError: nil,
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
				req: &lowribeckv1.GetAvailableSlotsPointOfSaleRequest{
					Postcode:              "ZE 11",
					Mpan:                  "mpan-1",
					Mprn:                  "mprn-1",
					ElectricityTariffType: lowribeckv1.TariffType_TARIFF_TYPE_CREDIT,
					GasTariffType:         lowribeckv1.TariffType_TARIFF_TYPE_PREPAYMENT,
				},
			},
			expectedError: nil,
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
		{
			desc: "should error because invalid tariff type was sent",
			input: inputParams{
				id: 1,
				req: &lowribeckv1.GetAvailableSlotsPointOfSaleRequest{
					Postcode:              "ZE 11",
					Mpan:                  "mpan-1",
					Mprn:                  "mprn-1",
					ElectricityTariffType: lowribeckv1.TariffType_TARIFF_TYPE_UNKNOWN,
					GasTariffType:         lowribeckv1.TariffType_TARIFF_TYPE_PREPAYMENT,
				},
			},
			expectedError: mapper.ErrInvalidElectricityTariffType,
			expected:      nil,
		},
	}

	assert := assert.New(t)
	lbMapper := mapper.NewLowriBeckMapper("sendingSystem", "receivingSystem", "crElec", "ppmElec", "crGas", "ppmGas")

	for _, tc := range testCases {
		t.Run(tc.desc, func(_ *testing.T) {
			res, err := lbMapper.AvailabilityRequestPointOfSale(tc.input.id, tc.input.req)
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

func TestMapBookingPointOfSaleRequest(t *testing.T) {

	type inputParams struct {
		id  uint32
		req *lowribeckv1.CreateBookingPointOfSaleRequest
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
				req: &lowribeckv1.CreateBookingPointOfSaleRequest{
					Mpan:                  "mpan-1",
					Mprn:                  "",
					ElectricityTariffType: lowribeckv1.TariffType_TARIFF_TYPE_CREDIT,
					GasTariffType:         lowribeckv1.TariffType_TARIFF_TYPE_UNKNOWN,
					Slot: &lowribeckv1.BookingSlot{
						Date: &date.Date{
							Year:  2020,
							Month: 12,
							Day:   13,
						},
						StartTime: 10,
						EndTime:   12,
					},
					VulnerabilityDetails: &lowribeckv1.VulnerabilityDetails{
						Vulnerabilities: []lowribeckv1.Vulnerability{
							lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
						},
						Other: "Other Vuln",
					},
					ContactDetails: &lowribeckv1.ContactDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Doe",
						Phone:     "2002-2001",
					},
					SiteAddress: &addressv1.Address{
						Uprn: "uprn-1",
						Paf: &addressv1.Address_PAF{
							Organisation:            "org",
							Department:              "department-1",
							SubBuilding:             "sub-1",
							BuildingName:            "bn-1",
							BuildingNumber:          "bnum-1",
							DependentThoroughfare:   "dt-1",
							Thoroughfare:            "tf-1",
							DoubleDependentLocality: "ddl-1",
							DependentLocality:       "dl-1",
							PostTown:                "pt",
							Postcode:                "ZE 11",
						},
					},
				},
			},
			expected: &lowribeck.CreateBookingRequest{
				RequestID:               "1",
				SendingSystem:           "sendingSystem",
				ReceivingSystem:         "receivingSystem",
				AppointmentDate:         "13/12/2020",
				AppointmentTime:         "10:00-12:00",
				SubBuildName:            "sub-1",
				BuildingName:            "bnum-1 bn-1",
				DependThroughfare:       "dt-1",
				Throughfare:             "tf-1",
				DoubleDependantLocality: "ddl-1",
				DependantLocality:       "dl-1",
				PostTown:                "pt",
				County:                  "", // There is no County in the PAF format
				PostCode:                "ZE 11",
				Mpan:                    "mpan-1",
				Mprn:                    "",
				ElecJobTypeCode:         "crElec",
				GasJobTypeCode:          "",
				Vulnerabilities:         "6",
				VulnerabilitiesOther:    "Other Vuln",
				SiteContactName:         "Mr John Doe",
				SiteContactNumber:       "2002-2001",
				CreatedDate:             time.Now().UTC().Format(requestTimeFormat),
			},
		},
		{
			desc: "Success - Electricity Prepayment - No MPRN",
			input: inputParams{
				id: 1,
				req: &lowribeckv1.CreateBookingPointOfSaleRequest{
					Mpan:                  "mpan-1",
					Mprn:                  "",
					ElectricityTariffType: lowribeckv1.TariffType_TARIFF_TYPE_PREPAYMENT,
					GasTariffType:         lowribeckv1.TariffType_TARIFF_TYPE_UNKNOWN,
					Slot: &lowribeckv1.BookingSlot{
						Date: &date.Date{
							Year:  2020,
							Month: 12,
							Day:   13,
						},
						StartTime: 10,
						EndTime:   12,
					},
					VulnerabilityDetails: &lowribeckv1.VulnerabilityDetails{
						Vulnerabilities: []lowribeckv1.Vulnerability{
							lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
						},
						Other: "Other Vuln",
					},
					ContactDetails: &lowribeckv1.ContactDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Doe",
						Phone:     "2002-2001",
					},
					SiteAddress: &addressv1.Address{
						Uprn: "uprn-1",
						Paf: &addressv1.Address_PAF{
							Organisation:            "org",
							Department:              "department-1",
							SubBuilding:             "sub-1",
							BuildingName:            "bn-1",
							BuildingNumber:          "bnum-1",
							DependentThoroughfare:   "dt-1",
							Thoroughfare:            "tf-1",
							DoubleDependentLocality: "ddl-1",
							DependentLocality:       "dl-1",
							PostTown:                "pt",
							Postcode:                "ZE 11",
						},
					},
				},
			},
			expected: &lowribeck.CreateBookingRequest{
				RequestID:               "1",
				SendingSystem:           "sendingSystem",
				ReceivingSystem:         "receivingSystem",
				AppointmentDate:         "13/12/2020",
				AppointmentTime:         "10:00-12:00",
				SubBuildName:            "sub-1",
				BuildingName:            "bnum-1 bn-1",
				DependThroughfare:       "dt-1",
				Throughfare:             "tf-1",
				DoubleDependantLocality: "ddl-1",
				DependantLocality:       "dl-1",
				PostTown:                "pt",
				County:                  "", // There is no County in the PAF format
				PostCode:                "ZE 11",
				Mpan:                    "mpan-1",
				Mprn:                    "",
				ElecJobTypeCode:         "ppmElec",
				GasJobTypeCode:          "",
				Vulnerabilities:         "6",
				VulnerabilitiesOther:    "Other Vuln",
				SiteContactName:         "Mr John Doe",
				SiteContactNumber:       "2002-2001",
				CreatedDate:             time.Now().UTC().Format(requestTimeFormat),
			},
		},
		{
			desc: "Success - Electricity Credit - Gas Credit",
			input: inputParams{
				id: 1,
				req: &lowribeckv1.CreateBookingPointOfSaleRequest{
					Mpan:                  "mpan-1",
					Mprn:                  "mprn-1",
					ElectricityTariffType: lowribeckv1.TariffType_TARIFF_TYPE_CREDIT,
					GasTariffType:         lowribeckv1.TariffType_TARIFF_TYPE_CREDIT,
					Slot: &lowribeckv1.BookingSlot{
						Date: &date.Date{
							Year:  2020,
							Month: 12,
							Day:   13,
						},
						StartTime: 10,
						EndTime:   12,
					},
					VulnerabilityDetails: &lowribeckv1.VulnerabilityDetails{
						Vulnerabilities: []lowribeckv1.Vulnerability{
							lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
						},
						Other: "Other Vuln",
					},
					ContactDetails: &lowribeckv1.ContactDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Doe",
						Phone:     "2002-2001",
					},
					SiteAddress: &addressv1.Address{
						Uprn: "uprn-1",
						Paf: &addressv1.Address_PAF{
							Organisation:            "org",
							Department:              "department-1",
							SubBuilding:             "sub-1",
							BuildingName:            "bn-1",
							BuildingNumber:          "bnum-1",
							DependentThoroughfare:   "dt-1",
							Thoroughfare:            "tf-1",
							DoubleDependentLocality: "ddl-1",
							DependentLocality:       "dl-1",
							PostTown:                "pt",
							Postcode:                "ZE 11",
						},
					},
				},
			},
			expected: &lowribeck.CreateBookingRequest{
				RequestID:               "1",
				SendingSystem:           "sendingSystem",
				ReceivingSystem:         "receivingSystem",
				AppointmentDate:         "13/12/2020",
				AppointmentTime:         "10:00-12:00",
				SubBuildName:            "sub-1",
				BuildingName:            "bnum-1 bn-1",
				DependThroughfare:       "dt-1",
				Throughfare:             "tf-1",
				DoubleDependantLocality: "ddl-1",
				DependantLocality:       "dl-1",
				PostTown:                "pt",
				County:                  "", // There is no County in the PAF format
				PostCode:                "ZE 11",
				Mpan:                    "mpan-1",
				Mprn:                    "mprn-1",
				ElecJobTypeCode:         "crElec",
				GasJobTypeCode:          "crGas",
				Vulnerabilities:         "6",
				VulnerabilitiesOther:    "Other Vuln",
				SiteContactName:         "Mr John Doe",
				SiteContactNumber:       "2002-2001",
				CreatedDate:             time.Now().UTC().Format(requestTimeFormat),
			},
		},
		{
			desc: "Success - Electricity Credit - Prepayment Credit",
			input: inputParams{
				id: 1,
				req: &lowribeckv1.CreateBookingPointOfSaleRequest{
					Mpan:                  "mpan-1",
					Mprn:                  "mprn-1",
					ElectricityTariffType: lowribeckv1.TariffType_TARIFF_TYPE_CREDIT,
					GasTariffType:         lowribeckv1.TariffType_TARIFF_TYPE_PREPAYMENT,
					Slot: &lowribeckv1.BookingSlot{
						Date: &date.Date{
							Year:  2020,
							Month: 12,
							Day:   13,
						},
						StartTime: 10,
						EndTime:   12,
					},
					VulnerabilityDetails: &lowribeckv1.VulnerabilityDetails{
						Vulnerabilities: []lowribeckv1.Vulnerability{
							lowribeckv1.Vulnerability_VULNERABILITY_FOREIGN_LANGUAGE_ONLY,
						},
						Other: "Other Vuln",
					},
					ContactDetails: &lowribeckv1.ContactDetails{
						Title:     "Mr",
						FirstName: "John",
						LastName:  "Doe",
						Phone:     "2002-2001",
					},
					SiteAddress: &addressv1.Address{
						Uprn: "uprn-1",
						Paf: &addressv1.Address_PAF{
							Organisation:            "org",
							Department:              "department-1",
							SubBuilding:             "sub-1",
							BuildingName:            "bn-1",
							BuildingNumber:          "bnum-1",
							DependentThoroughfare:   "dt-1",
							Thoroughfare:            "tf-1",
							DoubleDependentLocality: "ddl-1",
							DependentLocality:       "dl-1",
							PostTown:                "pt",
							Postcode:                "ZE 11",
						},
					},
				},
			},
			expected: &lowribeck.CreateBookingRequest{
				RequestID:               "1",
				SendingSystem:           "sendingSystem",
				ReceivingSystem:         "receivingSystem",
				AppointmentDate:         "13/12/2020",
				AppointmentTime:         "10:00-12:00",
				SubBuildName:            "sub-1",
				BuildingName:            "bnum-1 bn-1",
				DependThroughfare:       "dt-1",
				Throughfare:             "tf-1",
				DoubleDependantLocality: "ddl-1",
				DependantLocality:       "dl-1",
				PostTown:                "pt",
				County:                  "", // There is no County in the PAF format
				PostCode:                "ZE 11",
				Mpan:                    "mpan-1",
				Mprn:                    "mprn-1",
				ElecJobTypeCode:         "crElec",
				GasJobTypeCode:          "ppmGas",
				Vulnerabilities:         "6",
				VulnerabilitiesOther:    "Other Vuln",
				SiteContactName:         "Mr John Doe",
				SiteContactNumber:       "2002-2001",
				CreatedDate:             time.Now().UTC().Format(requestTimeFormat),
			},
		},
	}

	assert := assert.New(t)
	lbMapper := mapper.NewLowriBeckMapper("sendingSystem", "receivingSystem", "crElec", "ppmElec", "crGas", "ppmGas")

	for _, tc := range testCases {
		t.Run(tc.desc, func(_ *testing.T) {
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

func Test_MapBookingPointOfSaleResponse(t *testing.T) {

	type inputParams struct {
		req *lowribeck.CreateBookingResponse
	}

	testCases := []struct {
		desc          string
		input         inputParams
		expected      *lowribeckv1.CreateBookingPointOfSaleResponse
		expectedError error
	}{
		{
			desc: "Success - Has Reference ID",
			input: inputParams{
				req: &lowribeck.CreateBookingResponse{
					ReferenceID:  "reference-id-1",
					ResponseCode: "B01",
				},
			},
			expected: &lowribeckv1.CreateBookingPointOfSaleResponse{
				Success:   true,
				Reference: "reference-id-1",
			},
		},
		{
			desc: "Success - Has Reference ID",
			input: inputParams{
				req: &lowribeck.CreateBookingResponse{
					ReferenceID:  "reference-id-1",
					ResponseCode: "R01",
				},
			},
			expected: &lowribeckv1.CreateBookingPointOfSaleResponse{
				Success:   true,
				Reference: "reference-id-1",
			},
		},
	}

	assert := assert.New(t)
	lbMapper := mapper.NewLowriBeckMapper("sendingSystem", "receivingSystem", "crElec", "ppmElec", "crGas", "ppmGas")

	for _, tc := range testCases {
		t.Run(tc.desc, func(_ *testing.T) {
			res, err := lbMapper.BookingResponsePointOfSale(tc.input.req)
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
