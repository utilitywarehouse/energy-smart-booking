package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var LBErrorsCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "lb_errors_total",
	Help: "the count of each type of error",
}, []string{"type", "endpoint"})

// LBErrorsCount type
const (
	AppointmentNotFound           = "appointment_not_found"
	AppointmentAlreadyExists      = "appointment_already_exists"
	AppointmentOutOfRange         = "appointment_out_of_range"
	Internal                      = "internal"
	LBStatus                      = "lb_status"
	Unknown                       = "unknown"
	InvalidPostcode               = "invalid_postcode"
	InvalidReference              = "invalid_reference"
	InvalidSite                   = "invalid_site"
	InvalidAppointmentDate        = "invalid_appointment_date"
	InvalidAppointmentTime        = "invalid_appointment_time"
	InvalidUnknownParameter       = "invalid_unknown_parameter"
	InvalidMPAN                   = "invalid_mpan"
	InvalidMPRN                   = "invalid_mprn"
	InvalidElectricityJobTypeCode = "invalid_electricity_job_type_code"
	InvalidGasJobTypeCode         = "invalid_gas_job_type_code"
)

// LBErrorsCount endpoint
const (
	GetAvailableSlots = "get_available_slots"
	CreateBooking     = "create_booking"
)

var LBAPIRunning = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "lb_api_running",
	Help: "yes or no (1,0) the LB API is currently running",
})

var LBResponseCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "lb_responses_total",
	Help: "the status code returned from each LB request",
}, []string{"code", "endpoint"})
