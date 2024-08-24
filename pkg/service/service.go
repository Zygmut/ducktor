package service

import (
	"ducktor/pkg/healthcheck"
	"fmt"
	"log/slog"
	"time"
)

type Service struct {
	Name               string
	Checker            healthcheck.HealthChecker
	Interval           time.Duration
	UnhealthyThreshold int
	HealthyThreshold   int
	UnhealthyCount     int
	HealthyCount       int
	IsHealthy          bool
}

func (s *Service) Check() healthcheck.HealthCheckResult {
	status := s.Checker.CheckHealth()

	slog.Info(fmt.Sprintf("Service %s healthcheck: %+v", s.Name, status))

	return status
}
