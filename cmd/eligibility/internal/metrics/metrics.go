package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var SmartBookingEvaluationCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "smart_booking_evaluation_total",
	Help: "the total number of evaluations for smart booking",
}, []string{"criteria"})
