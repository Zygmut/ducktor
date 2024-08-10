package monitor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"ducktor/pkg/healthcheck"
	"ducktor/pkg/service"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Monitor struct {
	Services        []service.Service
	SuccessCounters []prometheus.Counter
	FailedCounters  []prometheus.Counter
}

var (
	healthMu sync.Mutex
)

func NewMonitor(configs []healthcheck.HealthCheck) (*Monitor, error) {
	services := make([]service.Service, len(configs))
	success := make([]prometheus.Counter, len(configs))
	fail := make([]prometheus.Counter, len(configs))

	for i, config := range configs {
		checker, err := healthcheck.NewHealthChecker(config)
		if err != nil {
			return nil, fmt.Errorf("error while creating HealthChecker: %s", err)
		}

		services[i] = service.Service{
			Name:               config.Name,
			Checker:            checker,
			Interval:           time.Duration(config.Interval) * time.Second,
			UnHealthyThreshold: config.UnHealthyThreshold,
			HealthyThreshold:   config.HealthyThreshold,
			IsHealthy:          false,
		}

		success[i] = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("ducktor_%s_total_successes", config.Name),
				Help: fmt.Sprintf("The total number of health check successes for service %s.", config.Name),
			},
		)

		fail[i] = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("ducktor_%s_total_failures", config.Name),
				Help: fmt.Sprintf("The total number of health check failures for service %s.", config.Name),
			},
		)

	}

	return &Monitor{Services: services, SuccessCounters: success, FailedCounters: fail}, nil
}

func health(m Monitor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK

		w.Header().Set("Content-Type", "application/json")

		response := make(map[string]string)

		healthMu.Lock()
		defer healthMu.Unlock()

		for _, svc := range m.Services {
			if !svc.IsHealthy {
				status = http.StatusServiceUnavailable
				response[svc.Name] = "KO"
			} else {
				response[svc.Name] = "OK"
			}
		}

		w.WriteHeader(status)

		json.NewEncoder(w).Encode(response)
	}
}

func serve(m *Monitor, port int) {
	http.HandleFunc("/health", health(*m))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func (m *Monitor) Run(port int) {

	go serve(m, port)

	for idx := range m.Services {
		svc := &m.Services[idx]
		succ := m.SuccessCounters[idx]
		fail := m.FailedCounters[idx]

		go func(s *service.Service, succ prometheus.Counter, fail prometheus.Counter) {
			for {
				result := s.Check()

				healthMu.Lock()

				if result.IsHealthy {
					s.UnhealthyCount = 0
					s.HealthyCount++

					if s.HealthyCount >= s.HealthyThreshold {
						s.IsHealthy = true
						succ.Inc()
					}

				} else {
					s.HealthyCount = 0
					s.UnhealthyCount++

					if s.UnhealthyCount >= s.UnHealthyThreshold {
						s.IsHealthy = false
						fail.Inc()
					}
				}

				healthMu.Unlock()

				time.Sleep(s.Interval)
			}
		}(svc, succ, fail)
	}

	// Keep the main function alive
	select {}
}
