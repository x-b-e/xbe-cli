package telemetry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear all relevant env vars
	clearEnvVars(t)

	cfg := LoadConfig()

	if cfg.Enabled {
		t.Error("expected Enabled to be false by default")
	}
	if cfg.TracesExporter != "otlp" {
		t.Errorf("expected TracesExporter to be 'otlp', got %q", cfg.TracesExporter)
	}
	if cfg.MetricsExporter != "otlp" {
		t.Errorf("expected MetricsExporter to be 'otlp', got %q", cfg.MetricsExporter)
	}
	if cfg.OTLPEndpoint != "localhost:4317" {
		t.Errorf("expected OTLPEndpoint to be 'localhost:4317', got %q", cfg.OTLPEndpoint)
	}
	if cfg.OTLPProtocol != "grpc" {
		t.Errorf("expected OTLPProtocol to be 'grpc', got %q", cfg.OTLPProtocol)
	}
	if cfg.OTLPInsecure {
		t.Error("expected OTLPInsecure to be false by default (secure by default)")
	}
}

func TestLoadConfig_EnvVars(t *testing.T) {
	clearEnvVars(t)

	t.Setenv("XBE_TELEMETRY_ENABLED", "1")
	t.Setenv("OTEL_TRACES_EXPORTER", "console")
	t.Setenv("OTEL_METRICS_EXPORTER", "none")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "collector.example.com:4317")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "true")
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "Authorization=Bearer token,X-Custom=value")

	cfg := LoadConfig()

	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if cfg.TracesExporter != "console" {
		t.Errorf("expected TracesExporter to be 'console', got %q", cfg.TracesExporter)
	}
	if cfg.MetricsExporter != "none" {
		t.Errorf("expected MetricsExporter to be 'none', got %q", cfg.MetricsExporter)
	}
	if cfg.OTLPEndpoint != "collector.example.com:4317" {
		t.Errorf("expected OTLPEndpoint to be 'collector.example.com:4317', got %q", cfg.OTLPEndpoint)
	}
	if cfg.OTLPProtocol != "http/protobuf" {
		t.Errorf("expected OTLPProtocol to be 'http/protobuf', got %q", cfg.OTLPProtocol)
	}
	if !cfg.OTLPInsecure {
		t.Error("expected OTLPInsecure to be true")
	}
	if len(cfg.OTLPHeaders) != 2 {
		t.Errorf("expected 2 headers, got %d", len(cfg.OTLPHeaders))
	}
	if cfg.OTLPHeaders["Authorization"] != "Bearer token" {
		t.Errorf("expected Authorization header to be 'Bearer token', got %q", cfg.OTLPHeaders["Authorization"])
	}
	if cfg.OTLPHeaders["X-Custom"] != "value" {
		t.Errorf("expected X-Custom header to be 'value', got %q", cfg.OTLPHeaders["X-Custom"])
	}
}

func TestLoadConfig_FileConfig(t *testing.T) {
	clearEnvVars(t)

	// Create a temp config file
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "xbe")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configContent := `{
		"telemetry": {
			"enabled": true,
			"traces_exporter": "console",
			"metrics_exporter": "none",
			"otlp_endpoint": "file.example.com:4317",
			"otlp_insecure": true
		}
	}`
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Set XDG_CONFIG_HOME to our temp dir
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg := LoadConfig()

	if !cfg.Enabled {
		t.Error("expected Enabled to be true from file")
	}
	if cfg.TracesExporter != "console" {
		t.Errorf("expected TracesExporter to be 'console', got %q", cfg.TracesExporter)
	}
	if cfg.MetricsExporter != "none" {
		t.Errorf("expected MetricsExporter to be 'none', got %q", cfg.MetricsExporter)
	}
	if cfg.OTLPEndpoint != "file.example.com:4317" {
		t.Errorf("expected OTLPEndpoint to be 'file.example.com:4317', got %q", cfg.OTLPEndpoint)
	}
	if !cfg.OTLPInsecure {
		t.Error("expected OTLPInsecure to be true from file")
	}
}

func TestLoadConfig_EnvOverridesFile(t *testing.T) {
	clearEnvVars(t)

	// Create a temp config file
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "xbe")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configContent := `{
		"telemetry": {
			"enabled": true,
			"otlp_endpoint": "file.example.com:4317"
		}
	}`
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	// Env should override file
	t.Setenv("XBE_TELEMETRY_ENABLED", "0")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "env.example.com:4317")

	cfg := LoadConfig()

	// Env overrides file for enabled
	if cfg.Enabled {
		t.Error("expected Enabled to be false (env override)")
	}
	// Env overrides file for endpoint
	if cfg.OTLPEndpoint != "env.example.com:4317" {
		t.Errorf("expected OTLPEndpoint to be 'env.example.com:4317', got %q", cfg.OTLPEndpoint)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	clearEnvVars(t)

	// Create a temp config file with invalid JSON
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "xbe")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, []byte("{invalid json"), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Should not panic, should use defaults
	cfg := LoadConfig()

	// Defaults should be used
	if cfg.Enabled {
		t.Error("expected Enabled to be false (default after invalid JSON)")
	}
	if cfg.OTLPEndpoint != "localhost:4317" {
		t.Errorf("expected OTLPEndpoint to be 'localhost:4317' (default), got %q", cfg.OTLPEndpoint)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	clearEnvVars(t)

	// Point to a directory that doesn't have a config file
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Should not panic, should use defaults
	cfg := LoadConfig()

	if cfg.Enabled {
		t.Error("expected Enabled to be false (default when file missing)")
	}
	if cfg.OTLPEndpoint != "localhost:4317" {
		t.Errorf("expected OTLPEndpoint to be 'localhost:4317' (default), got %q", cfg.OTLPEndpoint)
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1", true},
		{"true", true},
		{"TRUE", true},
		{"True", true},
		{"yes", true},
		{"YES", true},
		{"0", false},
		{"false", false},
		{"FALSE", false},
		{"no", false},
		{"", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseBool(tt.input)
			if result != tt.expected {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]string
	}{
		{"key=value", map[string]string{"key": "value"}},
		{"key1=value1,key2=value2", map[string]string{"key1": "value1", "key2": "value2"}},
		{"key=value with spaces", map[string]string{"key": "value with spaces"}},
		{" key = value ", map[string]string{"key": "value"}},
		{"", map[string]string{}},
		{"invalid", map[string]string{}},
		{"key=", map[string]string{"key": ""}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseHeaders(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseHeaders(%q) returned %d headers, want %d", tt.input, len(result), len(tt.expected))
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("parseHeaders(%q)[%q] = %q, want %q", tt.input, k, result[k], v)
				}
			}
		})
	}
}

func clearEnvVars(t *testing.T) {
	t.Helper()
	envVars := []string{
		"XBE_TELEMETRY_ENABLED",
		"OTEL_TRACES_EXPORTER",
		"OTEL_METRICS_EXPORTER",
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_PROTOCOL",
		"OTEL_EXPORTER_OTLP_INSECURE",
		"OTEL_EXPORTER_OTLP_HEADERS",
		"XDG_CONFIG_HOME",
	}
	for _, env := range envVars {
		t.Setenv(env, "")
	}
}
