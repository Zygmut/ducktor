package monitor

import (
	"ducktor/pkg/healthcheck"
	"ducktor/pkg/metrics"
	"ducktor/pkg/service"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
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
)

func NewMonitor(configs []healthcheck.HealthCheck) (*Monitor, error) {
	services := make([]service.Service, len(configs))
	lastChecks := make([]healthcheck.HealthCheckResult, len(configs))

	for i, config := range configs {
		checker, err := healthcheck.HealthCheckerFrom(config)

		if err != nil {
			return nil, fmt.Errorf("error while creating HealthChecker: %s", err)
		}

		services[i] = service.Service{
			Name:               config.Name,
			Checker:            checker,
			Interval:           time.Duration(config.Interval) * time.Second,
			UnhealthyThreshold: config.UnHealthyThreshold,
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

					s.HealthyCount = min(s.HealthyThreshold, s.HealthyCount+1)

					if s.HealthyCount >= s.HealthyThreshold {
						s.IsHealthy = true
						metrics.Success.With(prometheus.Labels{"service_name": s.Name}).Inc()
						metrics.Health.With(prometheus.Labels{"service_name": s.Name}).Set(1)
					}

				} else {
					s.HealthyCount = 0

					s.UnhealthyCount = min(s.UnhealthyThreshold, s.UnhealthyCount+1)

					if s.UnhealthyCount >= s.UnhealthyThreshold {
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
	routes := map[string]http.HandlerFunc{
		"/health": health(*m),
		"/info":   info(*m),
	}

	for key, value := range routes {
		slog.Info(fmt.Sprintf("Registered http://localhost:%d%s", port, key))
		http.HandleFunc(key, value)
	}

	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func info(m Monitor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method `OPTIONS` not implemented"})
			return
		}

		// Headers and CORS stuff
		headers := map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
			"Content-Type":                 "application/json",
		}

		for key, value := range headers {
			w.Header().Set(key, value)
		}

		services := []map[string]interface{}{}

		healthMu.Lock()
		defer healthMu.Unlock()

		for idx, svc := range m.Services {
			serviceStatus := "OK"
			if !svc.IsHealthy {
				serviceStatus = "KO"
			}

			services = append(services, map[string]interface{}{
				"name":                svc.Name,
				"latency":             m.LastChecks[idx].ResponseTime.String(),
				"type":                strings.ReplaceAll(reflect.TypeOf(svc.Checker).Elem().Name(), "Checker", ""),
				"status":              serviceStatus,
				"healthy_threshold":   svc.HealthyThreshold,
				"healthy_count":       svc.HealthyCount,
				"unhealthy_threshold": svc.UnhealthyThreshold,
				"unhealthy_count":     svc.UnhealthyCount,
				"interface":           svc.Checker,
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
			if svc.IsHealthy {
				response[svc.Name] = "OK"
			} else {
				status = http.StatusServiceUnavailable
				response[svc.Name] = "KO"
			}
		}

		w.WriteHeader(status)

		json.NewEncoder(w).Encode(response)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
