package monitor

import (
	"ducktor/pkg/healthcheck"
	"ducktor/pkg/metrics"
	"ducktor/pkg/service"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Monitor struct {
	Services        []service.Service
	LastChecks      []healthcheck.HealthCheckResult
	SuccessCounters []prometheus.Counter
	FailedCounters  []prometheus.Counter
	HealthStatus    []prometheus.Summary
}

var (
	healthMu sync.Mutex

	//go:embed index.html
	HTML_CONTENT string
)

func NewMonitor(configs []healthcheck.HealthCheck) (*Monitor, error) {
	services := make([]service.Service, len(configs))
	lastChecks := make([]healthcheck.HealthCheckResult, len(configs))

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

		lastChecks[i] = healthcheck.HealthCheckResult{IsHealthy: false}

	}

	return &Monitor{Services: services, LastChecks: lastChecks}, nil
}

func (m *Monitor) Run(port int) {

	go serve(m, port)

	for idx := range m.Services {
		go func(idx int) {
			for {
				s := &m.Services[idx]

				result := s.Check()

				healthMu.Lock()

				m.LastChecks[idx] = result

				if result.IsHealthy {
					s.UnhealthyCount = 0
					s.HealthyCount++

					if s.HealthyCount >= s.HealthyThreshold {
						s.IsHealthy = true
						metrics.Success.With(prometheus.Labels{"service_name": s.Name}).Inc()
						metrics.Health.With(prometheus.Labels{"service_name": s.Name}).Set(1)
					}

				} else {
					s.HealthyCount = 0
					s.UnhealthyCount++

					if s.UnhealthyCount >= s.UnHealthyThreshold {
						s.IsHealthy = false
						metrics.Fail.With(prometheus.Labels{"service_name": s.Name}).Inc()
						metrics.Health.With(prometheus.Labels{"service_name": s.Name}).Set(0)
					}
				}

				healthMu.Unlock()

				time.Sleep(s.Interval)
			}
		}(idx)
	}

	// Keep the main function alive
	select {}
}

func serve(m *Monitor, port int) {

	http.HandleFunc("/health", health(*m))
	http.HandleFunc("/info", info(*m))
	http.HandleFunc("/", webView(*m, port))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func info(m Monitor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// CORS STUFF
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			return
		}

		w.Header().Set("Content-Type", "application/json")

		services := []map[string]string{}

		healthMu.Lock()
		defer healthMu.Unlock()

		for idx, svc := range m.Services {
			serviceStatus := "OK"
			if !svc.IsHealthy {
				serviceStatus = "KO"
			}

			services = append(services, map[string]string{
				"name":    svc.Name,
				"latency": m.LastChecks[idx].ResponseTime.String(),
				"status":  serviceStatus,
			})
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(services)
	}
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

func webView(m Monitor, port int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", string(HTML_CONTENT))
	}
}
