package healthcheck

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// HTTPChecker checks the health of an HTTP/HTTPS service.
type HTTPChecker struct {
	Endpoint string
	Host     string
	Port     int
	Protocol string
	Match    int // Expected HTTP status code
}

// CheckHealth checks the health of an HTTP/HTTPS service.
func (h *HTTPChecker) CheckHealth() HealthCheckResult {
	url := fmt.Sprintf("%s://%s:%s/%s", h.Protocol, h.Host, strconv.Itoa(h.Port), h.Endpoint)

	log.Printf("Url: %s", url)
	start := time.Now()

	resp, err := http.Get(url)
	responseTime := time.Since(start)

	log.Printf("Response: %+v", resp)

	status := "healthy"
	if err != nil || resp.StatusCode != h.Match {
		status = "unhealthy"
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	return HealthCheckResult{
		Status:       status,
		ResponseTime: responseTime,
		Error:        err,
	}
}
