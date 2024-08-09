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
	// url := fmt.Sprintf("%s://%s:%d/%s", h.Protocol, h.Host, h.Port, h.Endpoint)

	builder := strings.Builder{}

	builder.WriteString(h.Protocol)
	builder.WriteString("://")
	builder.WriteString(h.Host)
	if h.Port != 0 {
		builder.WriteString(":")
		builder.WriteString(strconv.Itoa(h.Port))
	}
	if h.Endpoint != "" {
		builder.WriteString("/")
		builder.WriteString(h.Endpoint)
	}

	url := builder.String()

	start := time.Now()

	resp, err := http.Get(url)
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
