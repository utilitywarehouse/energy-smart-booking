package mapper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
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
			expected: &contract.GetAvailableSlotsResponse{
				ErrorCodes: contract.AvailabilityErrorCodes_AVAILABILITY_NO_AVAILABLE_SLOTS.Enum(),
			},
		},
		{
			desc: "Failed - invalid request response 1",
			lb: &lowribeck.GetCalendarAvailabilityResponse{
				ResponseCode:    "EA02",
				ResponseMessage: "Unable to identify postcode",
			},
			expected: &contract.GetAvailableSlotsResponse{
				ErrorCodes: contract.AvailabilityErrorCodes_AVAILABILITY_INVALID_REQUEST.Enum(),
			},
		},
		{
			desc: "Failed - invalid request response 2",
			lb: &lowribeck.GetCalendarAvailabilityResponse{
				ResponseCode:    "EA03",
				ResponseMessage: "Postcode and Reference ID mismatch",
			},
			expected: &contract.GetAvailableSlotsResponse{
				ErrorCodes: contract.AvailabilityErrorCodes_AVAILABILITY_INVALID_REQUEST.Enum(),
			},
		},
		{
			desc: "Failed - generic response",
			lb: &lowribeck.GetCalendarAvailabilityResponse{
				// According to the docs this code doesn't exist, but I've seen it (a lot)
				ResponseCode:    "EC04",
				ResponseMessage: "Generic LB error for failed to process request",
			},
			expected: &contract.GetAvailableSlotsResponse{
				ErrorCodes: contract.AvailabilityErrorCodes_AVAILABILITY_INTERNAL_ERROR.Enum(),
			},
		},
	}

	assert := assert.New(t)
	lbMapper := mapper.NewLowriBeckMapper("sendingSystem", "receivingSystem")

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
	lbMapper := mapper.NewLowriBeckMapper("sendingSystem", "receivingSystem")

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

//	{
//	    "ReferenceId": "UTIL/4568973",
//	    "Mpan": "",
//	    "Mprn": "",
//	    "ResponseMessage": "Booking Confirmed",
//	    "ResponseCode": "B01",
//	    "RequestId": "1234234",
//	    "SendingSystem": "LB",
//	    "ReceivingSystem": "UTIL",
//	    "CreatedDate": "26/07/2023 14:24:34"
//	}
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
	}

	assert := assert.New(t)
	lbMapper := mapper.NewLowriBeckMapper("sendingSystem", "receivingSystem")

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
