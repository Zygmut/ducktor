package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	DefaultInterval  int             `toml:"default_interval"`
	DefaultThreshold int             `toml:"default_threshold"`
	Port             int             `toml:"port"`
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

type ServiceConfig struct {
	Name      string
	Interface string
	Endpoint  string
	Match     int
	Host      string
	Port      int
	Interval  int
	Threshold int
}
