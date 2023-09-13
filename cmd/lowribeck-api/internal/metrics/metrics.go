package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var AppointmentNotFoundCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "lb_appointment_not_found_errors_total",
	Help: "the total number of appointment not found errors from Lowri Beck for smart booking",
})

var AppointmentAlreadyExistsCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "lb_appointment_already_exists_errors_total",
	Help: "the total number of appointment already exist errors from Lowri Beck for smart booking",
})

var AppointmentOutOfRangeCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "lb_appointment_out_of_range_errors_total",
	Help: "the total number of appointment out of range errors from Lowri Beck for smart booking",
})

var UnknownErrorCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "lb_unknown_errors_total",
	Help: "the total number of appointment out of range errors from Lowri Beck for smart booking",
})

var InvalidPostcodeCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "lb_invalid_postcode_errors_total",
	Help: "the total number of invalid postcode errors from Lowri Beck for smart booking",
})

var InvalidReferenceCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "lb_invalid_reference_errors_total",
	Help: "the total number of invalid reference errors from Lowri Beck for smart booking",
})

var InvalidSiteCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "lb_invalid_site_errors_total",
	Help: "the total number of invalid site errors from Lowri Beck for smart booking",
})

var InvalidAppointmentDateCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "lb_invalid_appointment_date_errors_total",
	Help: "the total number of invalid appointment date errors from Lowri Beck for smart booking",
})

var InvalidAppointmentTimeCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "lb_invalid_appointment_time_errors_total",
	Help: "the total number of invalid appointment time errors from Lowri Beck for smart booking",
})

var InvalidUnknownParameterCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "lb_invalid_unknown_parameter_errors_total",
	Help: "the total number of invalid request with unknown parameter errors from Lowri Beck for smart booking",
})
