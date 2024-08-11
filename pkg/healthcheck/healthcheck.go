package healthcheck

import (
	"fmt"
	"time"
)

type HealthCheck struct {
	Name               string
	Interface          string
	Host               string
	Port               int
	Endpoint           string
	HealthyThreshold   int
	UnHealthyThreshold int
	Interval           int
	Match              int
}

type HealthChecker interface {
	CheckHealth() HealthCheckResult
}

type HealthCheckResult struct {
	IsHealthy    bool
	ResponseTime time.Duration
	Error        error
}

func NewHealthChecker(config HealthCheck) (HealthChecker, error) {
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
		return nil, fmt.Errorf("unsupported interface: %s", config.Interface)
	}
}
