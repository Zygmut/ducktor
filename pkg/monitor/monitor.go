package monitor

import (
	"ducktor/pkg/config"
	"ducktor/pkg/healthcheck"
	"ducktor/pkg/service"
	"fmt"
	"time"
)

type Monitor struct {
	Services         []service.Service
	DefaultInterval  int
	DefaultThreshold int
}

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
		}
	}

	return &Monitor{Services: services}, nil
}

func (m *Monitor) Run() {
	for _, svc := range m.Services {
		go func(s service.Service) {
			for {
				result := s.Check()

				if result.Status == "healthy" {
					s.UnhealthyCount = 0
				} else {
					s.UnhealthyCount++
					if s.UnhealthyCount >= s.Threshold {
						fmt.Printf("Service %s is unhealthy (after %d consecutive failures)\n", s.Name, s.Threshold)
					}
				}

				time.Sleep(s.Interval)
			}
		}(svc)
	}

	// Keep the main function alive
	select {}
}
