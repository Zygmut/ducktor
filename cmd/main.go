package main

import (
	"ducktor/pkg/config"
	"ducktor/pkg/monitor"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	configFile := flag.String("config", "config.toml", "Path to the configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configFile)

	if err != nil {
		slog.Error(fmt.Sprintf("Error loading configuration: %v", err))
		os.Exit(1)
	}

	slog.Info("Loaded configuration:")

	for _, service := range cfg.HealthChecks {
		slog.Error(fmt.Sprintf("%+v", service))
		os.Exit(1)
	}

	m, err := monitor.NewMonitor(cfg.HealthChecks)

	if err != nil {
		slog.Error(fmt.Sprintf("Error while creating Monitor: %s", err))
		os.Exit(1)
	}

	go m.Run(cfg.Port)

	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(":2112", nil)

	if err != nil {
		slog.Error(fmt.Sprintf("Error while creating the metrics endpoint: %s", err))
		os.Exit(1)
	}
}
