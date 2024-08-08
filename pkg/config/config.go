package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml" // Importing the TOML library
)

type Config struct {
	DefaultInterval  int             `toml:"default_interval"`
	DefaultThreshold int             `toml:"default_threshold"`
	Services         []ServiceConfig `toml:"services"`
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

// ServiceConfig represents the configuration for a single service.
type ServiceConfig struct {
	Name      string
	Interface string
	Endpoint  string
	Match     int // Expected HTTP status code or string for stream
	Host      string
	Port      int
	Interval  int
	Threshold int
}
