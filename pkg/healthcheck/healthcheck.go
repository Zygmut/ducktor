package healthcheck

import (
	"ducktor/pkg/config"
	"fmt"
	"time"
)

type HealthChecker interface {
	CheckHealth() HealthCheckResult
}

type HealthCheckResult struct {
	IsHealthy    bool
	ResponseTime time.Duration
	Error        error
}

func NewHealthChecker(config config.ServiceConfig) (HealthChecker, error) {
	switch config.Interface {
	case "http", "https":
		return &HTTPChecker{
			Endpoint: config.Endpoint,
			Host:     config.Host,
			Port:     config.Port,
			Protocol: config.Interface,
			Match:    config.Match,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", config.Interface)
	}
}
