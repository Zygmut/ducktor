package healthcheck

import (
	"fmt"
	"net/http"
	"time"
)

type HTTPChecker struct {
	Endpoint string
	Host     string
	Port     int
	Protocol string
	Match    int
}

func (h *HTTPChecker) CheckHealth() HealthCheckResult {
	url := fmt.Sprintf("%s://%s:%d/%s", h.Protocol, h.Host, h.Port, h.Endpoint)

	start := time.Now()

	resp, err := http.Get(url)
	responseTime := time.Since(start)

	if err == nil && resp.StatusCode != h.Match {
		err = fmt.Errorf("Application returned status code %d but we expected %d", resp.StatusCode, h.Match)
	}

	isHealthy := err == nil

	if resp != nil {
		defer resp.Body.Close()
	}

	return HealthCheckResult{
		IsHealthy:    isHealthy,
		ResponseTime: responseTime,
		Error:        err,
	}
}
