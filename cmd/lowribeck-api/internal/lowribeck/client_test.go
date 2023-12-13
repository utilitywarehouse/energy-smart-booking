package lowribeck_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/lowribeck"
	"google.golang.org/protobuf/testing/protocmp"
)

func Test_GetCalendarAvailability_PointOfSale(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/appointmentManagement/getCalendarAvailability" {
			t.Errorf("Expected to request '/appointmentManagement/getCalendarAvailability', got: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ResponseCode": "B01", "Mpan": "mpan-1", "Mprn": "mprn-1", "ElecJobTypeCode": "210", "GasJobTypeCode": "410", "CalendarAvailabilityResult": [{"AppointmentDate": "12/12/2012", "AppointmentTime": "12:00-14:00"}] }`))
	}))
	defer server.Close()

	client := lowribeck.New(server.Client(), "", "", server.URL+"/")

	assert := assert.New(t)

	expectedResult := &lowribeck.GetCalendarAvailabilityResponse{
		Mpan:            "mpan-1",
		Mprn:            "mprn-1",
		ElecJobTypeCode: "210",
		GasJobTypeCode:  "410",
		CalendarAvailabilityResult: []lowribeck.AvailabilitySlot{
			{
				AppointmentDate: "12/12/2012",
				AppointmentTime: "12:00-14:00",
			},
		},
		ResponseCode: "B01",
	}

	resp, err := client.GetCalendarAvailabilityPointOfSale(context.Background(), &lowribeck.GetCalendarAvailabilityRequest{
		RequestID:       "req-1",
		SendingSystem:   "uw",
		ReceivingSystem: "lb",
		PostCode:        "2EZ",
		Mpan:            "mpan-1",
		Mprn:            "mprn-1",
		ElecJobTypeCode: "210",
		GasJobTypeCode:  "410",
	})
	if err != nil {
		t.Fatal(err)
	}

	diff := cmp.Diff(expectedResult, resp, protocmp.Transform(), cmpopts.IgnoreUnexported())
	if !assert.Empty(diff) {
		t.Fatal(diff)
	}
}

func Test_GetCalendarAvailability(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/appointmentManagement/getCalendarAvailability" {
			t.Errorf("Expected to request '/appointmentManagement/getCalendarAvailability', got: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ResponseCode": "B01", "CalendarAvailabilityResult": [{"AppointmentDate": "12/12/2012", "AppointmentTime": "12:00-14:00"}] }`))
	}))
	defer server.Close()

	client := lowribeck.New(server.Client(), "", "", server.URL+"/")

	assert := assert.New(t)

	expectedResult := &lowribeck.GetCalendarAvailabilityResponse{
		CalendarAvailabilityResult: []lowribeck.AvailabilitySlot{
			{
				AppointmentDate: "12/12/2012",
				AppointmentTime: "12:00-14:00",
			},
		},
		ResponseCode: "B01",
	}

	resp, err := client.GetCalendarAvailability(context.Background(), &lowribeck.GetCalendarAvailabilityRequest{
		RequestID:       "req-1",
		SendingSystem:   "uw",
		ReceivingSystem: "lb",
		PostCode:        "2EZ",
		ReferenceID:     "ref-id-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	diff := cmp.Diff(expectedResult, resp, protocmp.Transform(), cmpopts.IgnoreUnexported())
	if !assert.Empty(diff) {
		t.Fatal(diff)
	}
}

func Test_CreateBooking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/appointmentManagement/book" {
			t.Errorf("Expected to request '/appointmentManagement/book', got: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ResponseCode": "B01"}`))
	}))
	defer server.Close()

	client := lowribeck.New(server.Client(), "", "", server.URL+"/")

	assert := assert.New(t)

	expectedResult := &lowribeck.CreateBookingResponse{
		ResponseCode: "B01",
	}

	resp, err := client.CreateBooking(context.Background(), &lowribeck.CreateBookingRequest{
		RequestID:       "req-1",
		SendingSystem:   "uw",
		ReceivingSystem: "lb",
		PostCode:        "2EZ",
		ReferenceID:     "ref-id-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	diff := cmp.Diff(expectedResult, resp, protocmp.Transform(), cmpopts.IgnoreUnexported())
	if !assert.Empty(diff) {
		t.Fatal(diff)
	}
}

func Test_CreateBooking_PointOfSale(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/appointmentManagement/book" {
			t.Errorf("Expected to request '/appointmentManagement/book', got: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ResponseCode": "B01"}`))
	}))
	defer server.Close()

	client := lowribeck.New(server.Client(), "", "", server.URL+"/")

	assert := assert.New(t)

	expectedResult := &lowribeck.CreateBookingResponse{
		ResponseCode: "B01",
	}

	resp, err := client.CreateBooking(context.Background(), &lowribeck.CreateBookingRequest{
		RequestID:       "req-1",
		SendingSystem:   "uw",
		ReceivingSystem: "lb",
		PostCode:        "2EZ",
		Mpan:            "mpan-1",
		Mprn:            "mprn-1",
		ElecJobTypeCode: "210",
		GasJobTypeCode:  "410",
	})
	if err != nil {
		t.Fatal(err)
	}

	diff := cmp.Diff(expectedResult, resp, protocmp.Transform(), cmpopts.IgnoreUnexported())
	if !assert.Empty(diff) {
		t.Fatal(diff)
	}
}

func Test_UpdateContactDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/appointmentManagement/updateContact" {
			t.Errorf("Expected to request '/appointmentManagement/updateContact', got: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ResponseCode": "U01","ResponseMessage": "Update Confirmed"}`))
	}))
	defer server.Close()

	client := lowribeck.New(server.Client(), "", "", server.URL+"/")

	assert := assert.New(t)

	expectedResult := &lowribeck.UpdateContactDetailsResponse{
		ResponseCode:    "U01",
		ResponseMessage: "Update Confirmed",
	}

	resp, err := client.UpdateContactDetails(context.Background(), &lowribeck.UpdateContactDetailsRequest{
		RequestID:         "req-1",
		SendingSystem:     "uw",
		ReceivingSystem:   "lb",
		ReferenceID:       "ref-id-1",
		Vulnerabilities:   "01,02,03",
		SiteContactName:   "Test User",
		SiteContactNumber: "01234567890",
	})
	if err != nil {
		t.Fatal(err)
	}

	diff := cmp.Diff(expectedResult, resp, protocmp.Transform(), cmpopts.IgnoreUnexported())
	if !assert.Empty(diff) {
		t.Fatal(diff)
	}
}
