package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	Success = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ducktor_total_successes",
			Help: "The total number of health check successes.",
		},
		[]string{"service_name"},
	)
	Fail = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ducktor_total_failures",
			Help: "The total number of health check failures.",
		},
		[]string{"service_name"},
	)
	Health = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ducktor_is_healthy",
		},
		[]string{"service_name"},
	)
)
