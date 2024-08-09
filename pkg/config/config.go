package config

import (
	"ducktor/pkg/healthcheck"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Port         int                       `toml:"port"`
	HealthChecks []healthcheck.HealthCheck `toml:"healthcheck"`
}

func LoadConfig(filePath string) (*Config, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("`%s` is not a valid path", filePath)
	}

	var cfg Config
	if _, err := toml.DecodeFile(filePath, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}

	return &cfg, nil
}
