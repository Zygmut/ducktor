package healthcheck

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HTTPChecker struct {
	Protocol string
	Host     string
	Port     int
	Endpoint string
	Match    int
}

func (h *HTTPChecker) CheckHealth() HealthCheckResult {
	url := strings.Builder{}

	url.WriteString(h.Protocol)
	url.WriteString("://")
	url.WriteString(h.Host)
	if h.Port != 0 {
		url.WriteString(":")
		url.WriteString(strconv.Itoa(h.Port))
	}
	if h.Endpoint != "" {
		url.WriteString("/")
		url.WriteString(h.Endpoint)
	}

	start := time.Now()
	resp, err := http.Get(url.String())
	responseTime := time.Since(start)

	if err == nil && resp.StatusCode != h.Match {
		err = fmt.Errorf("application returned status code %d but we expected %d", resp.StatusCode, h.Match)
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
