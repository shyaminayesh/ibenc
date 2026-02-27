package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the entire application configuration
type Config struct {
	Prometheus PrometheusConfig `yaml:"prometheus"`
	Iperf3     Iperf3Config     `yaml:"iperf3"`
	Metrics    MetricsConfig    `yaml:"metrics"`
}

// PrometheusConfig holds Grafana Cloud authentication and endpoint details
type PrometheusConfig struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Iperf3Config holds iperf3 test configuration
type Iperf3Config struct {
	Server   string `yaml:"server"`
	Port     int    `yaml:"port"`
	Duration int    `yaml:"duration"`
}

// MetricsConfig holds metric labels configuration
type MetricsConfig struct {
	Location    string `yaml:"location"`
	ISPName     string `yaml:"isp_name"`
	PackageName string `yaml:"package_name"`
}

// LoadConfig loads configuration from YAML file
func LoadConfig(configPath string) (*Config, error) {
	// Expand home directory if needed
	if configPath == "" {
		configPath = "ibenc.yaml"
	}

	if configPath[0:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(home, configPath[2:])
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadConfigWithDefaults loads config and applies environment variable overrides
func LoadConfigWithDefaults(configPath string) (*Config, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// Override with environment variables if present
	applyEnvOverrides(cfg)

	return cfg, nil
}

// Validate checks that all required fields are set
func (c *Config) Validate() error {
	// Prometheus validation
	if c.Prometheus.URL == "" {
		return fmt.Errorf("prometheus.url is required")
	}
	if c.Prometheus.Username == "" {
		return fmt.Errorf("prometheus.username is required")
	}
	if c.Prometheus.Password == "" {
		return fmt.Errorf("prometheus.password is required")
	}

	// Iperf3 validation
	if c.Iperf3.Server == "" {
		return fmt.Errorf("iperf3.server is required")
	}
	if c.Iperf3.Port <= 0 || c.Iperf3.Port > 65535 {
		return fmt.Errorf("iperf3.port must be between 1 and 65535")
	}
	if c.Iperf3.Duration <= 0 {
		return fmt.Errorf("iperf3.duration must be greater than 0")
	}

	// Metrics validation (optional, but should have at least location)
	if c.Metrics.Location == "" {
		return fmt.Errorf("metrics.location is required")
	}

	return nil
}

// applyEnvOverrides applies environment variable overrides to the config
func applyEnvOverrides(cfg *Config) {
	// Prometheus overrides
	if url := os.Getenv("IBENC_PROMETHEUS_URL"); url != "" {
		cfg.Prometheus.URL = url
	}
	if username := os.Getenv("IBENC_PROMETHEUS_USER"); username != "" {
		cfg.Prometheus.Username = username
	}
	if password := os.Getenv("IBENC_PROMETHEUS_PASS"); password != "" {
		cfg.Prometheus.Password = password
	}

	// Iperf3 overrides
	if server := os.Getenv("IBENC_SERVER"); server != "" {
		cfg.Iperf3.Server = server
	}

	// Metrics overrides
	if location := os.Getenv("IBENC_LOCATION"); location != "" {
		cfg.Metrics.Location = location
	}
	if ispName := os.Getenv("IBENC_ISP_NAME"); ispName != "" {
		cfg.Metrics.ISPName = ispName
	}
	if packageName := os.Getenv("IBENC_PACKAGE_NAME"); packageName != "" {
		cfg.Metrics.PackageName = packageName
	}
}

// GetMetricsLabels returns MetricsConfig as metrics.MetricLabels
func (c *Config) GetMetricsLabels() map[string]string {
	return map[string]string{
		"location":     c.Metrics.Location,
		"isp_name":     c.Metrics.ISPName,
		"package_name": c.Metrics.PackageName,
	}
}
