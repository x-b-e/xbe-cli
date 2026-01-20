package telemetry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds telemetry configuration loaded from environment variables
// and/or the config file.
type Config struct {
	Enabled         bool
	TracesExporter  string // "otlp", "console", "none"
	MetricsExporter string // "otlp", "console", "none"
	OTLPEndpoint    string
	OTLPProtocol    string // "grpc" or "http/protobuf"
	OTLPInsecure    bool   // Disable TLS (for local development only)
	OTLPHeaders     map[string]string
}

// fileConfig mirrors the JSON structure in ~/.config/xbe/config.json
type fileConfig struct {
	Telemetry *fileTelemetryConfig `json:"telemetry,omitempty"`
}

type fileTelemetryConfig struct {
	Enabled         *bool             `json:"enabled,omitempty"`
	TracesExporter  string            `json:"traces_exporter,omitempty"`
	MetricsExporter string            `json:"metrics_exporter,omitempty"`
	OTLPEndpoint    string            `json:"otlp_endpoint,omitempty"`
	OTLPProtocol    string            `json:"otlp_protocol,omitempty"`
	OTLPInsecure    *bool             `json:"otlp_insecure,omitempty"`
	OTLPHeaders     map[string]string `json:"otlp_headers,omitempty"`
}

// LoadConfig loads telemetry configuration with precedence:
// 1. Environment variables (highest)
// 2. Config file (~/.config/xbe/config.json)
// 3. Defaults (lowest)
func LoadConfig() Config {
	// Start with defaults (secure by default)
	cfg := Config{
		Enabled:         false,
		TracesExporter:  "otlp",
		MetricsExporter: "otlp",
		OTLPEndpoint:    "localhost:4317",
		OTLPProtocol:    "grpc",
		OTLPInsecure:    false, // TLS required by default
		OTLPHeaders:     make(map[string]string),
	}

	// Load from config file (if exists)
	if fileCfg, err := loadConfigFile(); err == nil && fileCfg.Telemetry != nil {
		applyFileConfig(&cfg, fileCfg.Telemetry)
	}

	// Override with environment variables (highest precedence)
	applyEnvVars(&cfg)

	return cfg
}

func loadConfigFile() (fileConfig, error) {
	path := configFilePath()
	content, err := os.ReadFile(path)
	if err != nil {
		return fileConfig{}, err
	}

	var cfg fileConfig
	if err := json.Unmarshal(content, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to parse config file %s: %v\n", path, err)
		return fileConfig{}, err
	}

	return cfg, nil
}

func configFilePath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "xbe", "config.json")
}

func applyFileConfig(cfg *Config, fileCfg *fileTelemetryConfig) {
	if fileCfg.Enabled != nil {
		cfg.Enabled = *fileCfg.Enabled
	}
	if fileCfg.TracesExporter != "" {
		cfg.TracesExporter = fileCfg.TracesExporter
	}
	if fileCfg.MetricsExporter != "" {
		cfg.MetricsExporter = fileCfg.MetricsExporter
	}
	if fileCfg.OTLPEndpoint != "" {
		cfg.OTLPEndpoint = fileCfg.OTLPEndpoint
	}
	if fileCfg.OTLPProtocol != "" {
		cfg.OTLPProtocol = fileCfg.OTLPProtocol
	}
	if fileCfg.OTLPInsecure != nil {
		cfg.OTLPInsecure = *fileCfg.OTLPInsecure
	}
	if len(fileCfg.OTLPHeaders) > 0 {
		cfg.OTLPHeaders = fileCfg.OTLPHeaders
	}
}

func applyEnvVars(cfg *Config) {
	// XBE_TELEMETRY_ENABLED is the master switch
	if val := os.Getenv("XBE_TELEMETRY_ENABLED"); val != "" {
		cfg.Enabled = parseBool(val)
	}

	// Standard OTEL environment variables
	if val := os.Getenv("OTEL_TRACES_EXPORTER"); val != "" {
		cfg.TracesExporter = val
	}
	if val := os.Getenv("OTEL_METRICS_EXPORTER"); val != "" {
		cfg.MetricsExporter = val
	}
	if val := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); val != "" {
		cfg.OTLPEndpoint = val
	}
	if val := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"); val != "" {
		cfg.OTLPProtocol = val
	}
	if val := os.Getenv("OTEL_EXPORTER_OTLP_INSECURE"); val != "" {
		cfg.OTLPInsecure = parseBool(val)
	}
	if val := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"); val != "" {
		cfg.OTLPHeaders = parseHeaders(val)
	}
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "1" || s == "true" || s == "yes"
}

func parseHeaders(s string) map[string]string {
	headers := make(map[string]string)
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return headers
}
