package service

import (
	"ducktor/pkg/healthcheck"
	"log"
	"time"
)

type Service struct {
	Name           string
	Checker        healthcheck.HealthChecker
	Interval       time.Duration
	Threshold      int
	UnhealthyCount int
	IsHealthy      bool
}

func (s *Service) Check() healthcheck.HealthCheckResult {
	status := s.Checker.CheckHealth()

	log.Printf("Service %s: %+v", s.Name, status)

	return status
}
