package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level portwatch configuration.
type Config struct {
	Interval time.Duration `yaml:"interval"`
	Ports    []PortConfig  `yaml:"ports"`
}

// PortConfig defines a single port to monitor and its actions.
type PortConfig struct {
	Host    string   `yaml:"host"`
	Port    int      `yaml:"port"`
	Actions []Action `yaml:"actions"`
}

// Action defines what to do when a port changes state.
type Action struct {
	On      string `yaml:"on"`      // "open", "close", or "any"
	Webhook string `yaml:"webhook"` // URL to POST to
	Command string `yaml:"command"` // shell command to run
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if cfg.Interval == 0 {
		cfg.Interval = 30 * time.Second
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	for i, p := range c.Ports {
		if p.Host == "" {
			return fmt.Errorf("port[%d]: host is required", i)
		}
		if p.Port <= 0 || p.Port > 65535 {
			return fmt.Errorf("port[%d]: port must be between 1 and 65535", i)
		}
		for j, a := range p.Actions {
			if a.On != "open" && a.On != "close" && a.On != "any" {
				return fmt.Errorf("port[%d].action[%d]: on must be 'open', 'close', or 'any'", i, j)
			}
			if a.Webhook == "" && a.Command == "" {
				return fmt.Errorf("port[%d].action[%d]: webhook or command is required", i, j)
			}
		}
	}
	return nil
}
