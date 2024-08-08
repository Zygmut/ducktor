package monitor

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"ducktor/pkg/config"
	"ducktor/pkg/healthcheck"
	"ducktor/pkg/service"
)

type Monitor struct {
	Services         []service.Service
	DefaultInterval  int
	DefaultThreshold int
}

var (
	healthMu sync.Mutex
)

func NewMonitor(configs []config.ServiceConfig, defaultInterval, defaultThreshold int) (*Monitor, error) {
	services := make([]service.Service, len(configs))

	for i, config := range configs {
		checker, err := healthcheck.NewHealthChecker(config)
		if err != nil {
			return nil, fmt.Errorf("Error while creating HealthChecker: %s", err)
		}

		interval := config.Interval
		if interval == 0 {
			interval = defaultInterval
		}

		threshold := config.Threshold
		if threshold == 0 {
			threshold = defaultThreshold
		}

		services[i] = service.Service{
			Name:      config.Name,
			Checker:   checker,
			Interval:  time.Duration(interval) * time.Second,
			Threshold: threshold,
			IsHealthy: false,
		}
	}

	return &Monitor{Services: services}, nil
}

func health(m Monitor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allHealthy := true

		healthMu.Lock()

		for _, svc := range m.Services {
			if !svc.IsHealthy {
				allHealthy = false
				break
			}
		}

		healthMu.Unlock()

		if allHealthy {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func status(m Monitor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response := make(map[string]string)

		healthMu.Lock()

		for _, svc := range m.Services {
			if svc.IsHealthy {
				response[svc.Name] = "OK"
			} else {
				response[svc.Name] = "KO"
			}
		}

		healthMu.Unlock()

		json.NewEncoder(w).Encode(response)
	}
}

func serve(m *Monitor, port int) {
	http.HandleFunc("/health", health(*m))
	http.HandleFunc("/status", status(*m))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func (m *Monitor) Run(port int) {

	go serve(m, port)

	for idx := range m.Services {
		svc := &m.Services[idx]

		go func(s *service.Service) {
			for {
				result := s.Check()

				healthMu.Lock()

				if result.IsHealthy {
					s.UnhealthyCount = 0
					s.IsHealthy = true

				} else {
					s.UnhealthyCount++
					if s.UnhealthyCount >= s.Threshold {
						log.Printf("Service %s is unhealthy (%d consecutive failures)\n", s.Name, s.UnhealthyCount)
					}
				}

				healthMu.Unlock()

				time.Sleep(s.Interval)
			}
		}(svc)
	}

	// Keep the main function alive
	select {}
}
