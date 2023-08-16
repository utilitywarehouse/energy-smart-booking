package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var QueryElapsedHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "booking_api_query_elapsed",
	Help: "the time spent in milliseconds for the query execution, including the network latency",
}, []string{"query_name"})
