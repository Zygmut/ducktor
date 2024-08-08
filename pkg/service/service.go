package service

import (
	"ducktor/pkg/healthcheck"
	"time"
)

type Service struct {
	Name           string
	Checker        healthcheck.HealthChecker
	Interval       time.Duration
	Threshold      int
	UnhealthyCount int
}

func (s *Service) Check() healthcheck.HealthCheckResult {
	return s.Checker.CheckHealth()
}
