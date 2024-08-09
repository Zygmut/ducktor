package service

import (
	"ducktor/pkg/healthcheck"
	"log"
	"time"
)

type Service struct {
	Name               string
	Checker            healthcheck.HealthChecker
	Interval           time.Duration
	UnHealthyThreshold int
	HealthyThreshold   int
	UnhealthyCount     int
	HealthyCount       int
	IsHealthy          bool
}

func (s *Service) Check() healthcheck.HealthCheckResult {
	status := s.Checker.CheckHealth()

	log.Printf("Service %s healthcheck: %+v", s.Name, status)

	return status
}
