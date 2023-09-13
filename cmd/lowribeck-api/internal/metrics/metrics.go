package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var LBErrorsCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "errors returned from Lowri Beck API",
	Help: "the count of each type of error",
}, []string{"type"})

const (
	AppointmentNotFound      = "appointment_not_found"
	AppointmentAlreadyExists = "appointment_already_exists"
	AppointmentOutOfRange    = "appointment_out_of_range"
	Unknown                  = "unknown"
	InvalidPostcode          = "invalid_postcode"
	InvalidReference         = "invalid_reference"
	InvalidSite              = "invalid_site"
	InvalidAppointmentDate   = "invalid_appointment_date"
	InvalidAppointmentTime   = "invalid_appointment_time"
	InvalidUnknownParameter  = "invalid_unknown_parameter"
)
