package main

import (
	"ducktor/pkg/config"
	"ducktor/pkg/monitor"
	"flag"
	"log"
)

func main() {
	configFile := flag.String("config", "config.toml", "Path to the configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configFile)

	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	log.Println("Loaded configuration:")

	for _, service := range cfg.HealthChecks {
		log.Printf("%+v", service)
	}

	m, err := monitor.NewMonitor(cfg.HealthChecks)

	if err != nil {
		log.Fatalf("Error while creating Monitor: %s", err)
	}

	m.Run(cfg.Port)
}
