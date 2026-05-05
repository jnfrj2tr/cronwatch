package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Job defines a single cron job to monitor.
type Job struct {
	Name     string        `yaml:"name"`
	Schedule string        `yaml:"schedule"`
	Timeout  time.Duration `yaml:"timeout"`
	Command  string        `yaml:"command"`
}

// AlertConfig holds alerting destination settings.
type AlertConfig struct {
	Email   string `yaml:"email"`
	Webhook string `yaml:"webhook"`
}

// Config is the top-level configuration structure.
type Config struct {
	CheckInterval time.Duration `yaml:"check_interval"`
	LogFile       string        `yaml:"log_file"`
	Alerts        AlertConfig   `yaml:"alerts"`
	Jobs          []Job         `yaml:"jobs"`
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

	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = time.Minute
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.Jobs) == 0 {
		return fmt.Errorf("no jobs defined")
	}
	names := make(map[string]struct{}, len(c.Jobs))
	for i, job := range c.Jobs {
		if job.Name == "" {
			return fmt.Errorf("job[%d]: name is required", i)
		}
		if job.Schedule == "" {
			return fmt.Errorf("job %q: schedule is required", job.Name)
		}
		if _, dup := names[job.Name]; dup {
			return fmt.Errorf("duplicate job name: %q", job.Name)
		}
		names[job.Name] = struct{}{}
	}
	return nil
}
