package monitor

import (
	"ducktor/pkg/healthcheck"
	"ducktor/pkg/metrics"
	"ducktor/pkg/service"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Monitor struct {
	Services        []service.Service
	LastChecks      []healthcheck.HealthCheckResult
	SuccessCounters []prometheus.Counter
	FailedCounters  []prometheus.Counter
	HealthStatus    []prometheus.Summary
}

var (
	healthMu sync.Mutex
)

func NewMonitor(configs []healthcheck.HealthCheck) (*Monitor, error) {
	services := make([]service.Service, len(configs))
	lastChecks := make([]healthcheck.HealthCheckResult, len(configs))

	for i, config := range configs {
		checker, err := healthcheck.NewHealthChecker(config)
		if err != nil {
			return nil, fmt.Errorf("error while creating HealthChecker: %s", err)
		}

		services[i] = service.Service{
			Name:               config.Name,
			Checker:            checker,
			Interval:           time.Duration(config.Interval) * time.Second,
			UnHealthyThreshold: config.UnHealthyThreshold,
			HealthyThreshold:   config.HealthyThreshold,
			IsHealthy:          false,
		}

		lastChecks[i] = healthcheck.HealthCheckResult{IsHealthy: false}

	}

	return &Monitor{Services: services, LastChecks: lastChecks}, nil
}

func (m *Monitor) Run(port int) {

	go serve(m, port)

	for idx := range m.Services {
		go func(idx int) {
			for {
				s := &m.Services[idx]

				result := s.Check()

				healthMu.Lock()

				m.LastChecks[idx] = result

				if result.IsHealthy {
					s.UnhealthyCount = 0
					s.HealthyCount++

					if s.HealthyCount >= s.HealthyThreshold {
						s.IsHealthy = true
						metrics.Success.With(prometheus.Labels{"service_name": s.Name}).Inc()
						metrics.Health.With(prometheus.Labels{"service_name": s.Name}).Set(1)
					}

				} else {
					s.HealthyCount = 0
					s.UnhealthyCount++

					if s.UnhealthyCount >= s.UnHealthyThreshold {
						s.IsHealthy = false
						metrics.Fail.With(prometheus.Labels{"service_name": s.Name}).Inc()
						metrics.Health.With(prometheus.Labels{"service_name": s.Name}).Set(0)
					}
				}

				healthMu.Unlock()

				time.Sleep(s.Interval)
			}
		}(idx)
	}

	// Keep the main function alive
	select {}
}

func serve(m *Monitor, port int) {

	http.HandleFunc("/health", health(*m))
	http.HandleFunc("/info", info(*m))
	http.HandleFunc("/", webView(*m, port))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func info(m Monitor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// CORS STUFF
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			return
		}

		w.Header().Set("Content-Type", "application/json")

		services := []map[string]string{}

		healthMu.Lock()
		defer healthMu.Unlock()

		for idx, svc := range m.Services {
			serviceStatus := "OK"
			if !svc.IsHealthy {
				serviceStatus = "KO"
			}

			services = append(services, map[string]string{
				"name":    svc.Name,
				"latency": m.LastChecks[idx].ResponseTime.String(),
				"status":  serviceStatus,
			})
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(services)
	}
}

func health(m Monitor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK

		w.Header().Set("Content-Type", "application/json")

		response := make(map[string]string)

		healthMu.Lock()
		defer healthMu.Unlock()

		for _, svc := range m.Services {
			if !svc.IsHealthy {
				status = http.StatusServiceUnavailable
				response[svc.Name] = "KO"
			} else {
				response[svc.Name] = "OK"
			}
		}

		w.WriteHeader(status)

		json.NewEncoder(w).Encode(response)
	}
}

func webView(m Monitor, port int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		webpage := PAGE_TEMPLATE

		healthyServices := 0
		unHealthyServices := 0

		for _, check := range m.Services {
			if check.IsHealthy {
				healthyServices++
			} else {
				unHealthyServices++
			}
		}

		ducktorStatus := "OK"
		if unHealthyServices > 0 {
			ducktorStatus = "KO"
		}

		checks := make([]string, len(m.Services))
		for idx := range m.Services {
			checks[idx] = serviceHtml(m.Services[idx], m.LastChecks[idx])
		}

		services := strings.Join(checks, "\n")

		webpage = strings.ReplaceAll(webpage, "{{total_healthy}}", strconv.Itoa(healthyServices))
		webpage = strings.ReplaceAll(webpage, "{{total_unhealthy}}", strconv.Itoa(unHealthyServices))
		webpage = strings.ReplaceAll(webpage, "{{ducktor_status}}", ducktorStatus)
		webpage = strings.ReplaceAll(webpage, "{{services}}", services)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", webpage)
	}
}

func serviceHtml(s service.Service, h healthcheck.HealthCheckResult) string {
	service := SERVICE_TEMPLATE
	color := "error"
	status := "down"
	if s.IsHealthy {
		color = "success"
		status = "up"
	}
	service = strings.ReplaceAll(service, "{{name}}", s.Name)
	service = strings.ReplaceAll(service, "{{latency}}", h.ResponseTime.String())
	service = strings.ReplaceAll(service, "{{color}}", color)
	service = strings.ReplaceAll(service, "{{status}}", status)
	service = strings.ReplaceAll(service, "{{interface}}", "HTTP")

	return service
}

const (
	SERVICE_TEMPLATE = `
            <div class="card-bordered bg-base-300 w-full rounded-lg overflow-hidden border-{{color}} p-4">
                <div class="flex items-center justify-between mb-2">
                    <h2 class="text-xl font-semibold">{{name}}</h2>

                    <i class="text-xl text-{{color}} fa-solid fa-thumbs-{{status}}"></i>
                </div>
                <!-- Extra Info Depending on Interface -->
                <p>
                    Endpoint: <span class="font-medium">http://localhost:8080/health</span>
                </p>

                <!-- Latency Info -->
                <p class="text-sm">
                    Latency: <span class="font-medium">{{latency}}</span>
                </p>
            </div>`
	PAGE_TEMPLATE = `<!DOCTYPE html>
<html>

<head>
    <meta charset='utf-8'>
    <meta http-equiv='X-UA-Compatible' content='IE=edge'>
    <title>Ducktor</title>
    <meta name='viewport' content='width=device-width, initial-scale=1'>
    <link href="https://cdn.jsdelivr.net/npm/daisyui@4.12.10/dist/full.min.css" rel="stylesheet" type="text/css" />
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.6.0/css/all.min.css"
        integrity="sha512-Kc323vGBEqzTmouAECnVceyQqyqdsSiqLQISBL29aUW4U/M7pSPA/gEUZQqv1cwx4OnYxTxve5UMg5GT6L4JJg=="
        crossorigin="anonymous" referrerpolicy="no-referrer" />
    <script src="https://cdn.tailwindcss.com"></script>
</head>

<body data-theme="dracula">
    <div class="container mx-auto max-w-6xl p-8">
        <nav class="mb-16 p-4 navbar shadow-lg bg-neutral text-neutral-content rounded-box">
            <div class="flex-none">
                <span class="text-4xl pb-1">🦆</span>
            </div>
            <div class="flex-1 px-2 mx-2">
                <span class="text-2xl font-bold">Ducktor</span>
            </div>
        </nav>

        <summary class="mb-8 min-v-screen flex flex-col items-center">
            <div class="stats shadow w-full bg-base-300 stats-vertical md:stats-horizontal">
                <div class="stat">
                    <div class="stat-title">Total Healthy Checks</div>
                    <div class="stat-value text-success">{{total_healthy}}</div>
                </div>

                <div class="stat">
                    <div class="stat-title">Total Unhealthy Checks</div>
                    <div id="total_unhealthy" class="stat-value text-error">{{total_unhealthy}}</div>
                </div>

                <div class="stat">
                    <div class="stat-title">Current Status</div>
                    <div id="current_status" class="stat-value text-accent">{{ducktor_status}}</div>
                </div>
            </div>
        </summary>

        <main class="min-h-screen flex flex-col items-center gap-4">
            {{services}}
        </main>
    </div>
<script>
  function fetchServiceStatus() {
    fetch('localhost:{{port}}/info')
      .then(response => response.json())
      .then(data => {
        // Update the page content with the new status
        document.getElementById('service-name').innerText = data.name;
        document.getElementById('service-latency').innerText = data.latency;
        document.getElementById('service-status').innerText = data.status;
        document.getElementById('service-interface').innerText = data.interface;

        // Update color based on status
        if (data.status === 'up') {
          document.getElementById('service-status').className = 'text-green-500';
        } else {
          document.getElementById('service-status').className = 'text-red-500';
        }
      })
      .catch(error => console.error('Error fetching service status:', error));
  }

  // Refresh every second (1000 ms)
  setInterval(fetchServiceStatus, 1000);
</script>
</body>

</html>
`
)
