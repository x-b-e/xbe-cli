package telemetry

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestInit_Disabled(t *testing.T) {
	// Clear env vars to ensure telemetry is disabled
	t.Setenv("XBE_TELEMETRY_ENABLED", "0")

	ctx := context.Background()
	provider, err := Init(ctx)
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	if provider == nil {
		t.Fatal("Init() returned nil provider")
	}

	if provider.Enabled() {
		t.Error("expected Enabled() to return false when disabled")
	}

	if !provider.noop {
		t.Error("expected provider to be in no-op mode")
	}

	// Shutdown should be a no-op and not error
	if err := provider.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() returned error: %v", err)
	}
}

func TestInit_DisabledIgnoresOTELVars(t *testing.T) {
	// Disable telemetry but set OTEL vars
	t.Setenv("XBE_TELEMETRY_ENABLED", "0")
	t.Setenv("OTEL_TRACES_EXPORTER", "console")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "example.com:4317")

	ctx := context.Background()
	provider, err := Init(ctx)
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	// Should still be disabled
	if provider.Enabled() {
		t.Error("expected Enabled() to return false even with OTEL vars set")
	}

	if !provider.noop {
		t.Error("expected provider to be in no-op mode")
	}
}

func TestInit_ConsoleExporter(t *testing.T) {
	t.Setenv("XBE_TELEMETRY_ENABLED", "1")
	t.Setenv("OTEL_TRACES_EXPORTER", "console")
	t.Setenv("OTEL_METRICS_EXPORTER", "console")

	ctx := context.Background()
	provider, err := Init(ctx)
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	if provider == nil {
		t.Fatal("Init() returned nil provider")
	}

	if !provider.Enabled() {
		t.Error("expected Enabled() to return true")
	}

	if provider.noop {
		t.Error("expected provider to NOT be in no-op mode")
	}

	// Tracer should not be nil
	tracer := provider.Tracer()
	if tracer == nil {
		t.Error("Tracer() returned nil")
	}

	// Shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := provider.Shutdown(shutdownCtx); err != nil {
		t.Errorf("Shutdown() returned error: %v", err)
	}
}

func TestShutdown_Noop(t *testing.T) {
	t.Setenv("XBE_TELEMETRY_ENABLED", "0")

	ctx := context.Background()
	provider, err := Init(ctx)
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	// No-op shutdown should be instant and not error
	start := time.Now()
	if err := provider.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() returned error: %v", err)
	}
	elapsed := time.Since(start)

	// Should be very fast (< 100ms)
	if elapsed > 100*time.Millisecond {
		t.Errorf("No-op shutdown took too long: %v", elapsed)
	}
}

func TestShutdown_WithTimeout(t *testing.T) {
	t.Setenv("XBE_TELEMETRY_ENABLED", "1")
	t.Setenv("OTEL_TRACES_EXPORTER", "console")
	t.Setenv("OTEL_METRICS_EXPORTER", "console")

	ctx := context.Background()
	provider, err := Init(ctx)
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	// Shutdown with a short timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Should complete without error (console exporter is fast)
	if err := provider.Shutdown(shutdownCtx); err != nil {
		t.Errorf("Shutdown() returned error: %v", err)
	}
}

func TestHTTPTransport_Wrapping(t *testing.T) {
	t.Setenv("XBE_TELEMETRY_ENABLED", "1")
	t.Setenv("OTEL_TRACES_EXPORTER", "console")

	ctx := context.Background()
	provider, err := Init(ctx)
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	defer provider.Shutdown(ctx)

	// Get wrapped transport
	transport := provider.HTTPTransport(http.DefaultTransport)

	// Should be different from the original (wrapped)
	if transport == http.DefaultTransport {
		t.Error("expected HTTPTransport to return a wrapped transport")
	}
}

func TestHTTPTransport_NoopWhenDisabled(t *testing.T) {
	t.Setenv("XBE_TELEMETRY_ENABLED", "0")

	ctx := context.Background()
	provider, err := Init(ctx)
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	// Get transport when disabled
	transport := provider.HTTPTransport(http.DefaultTransport)

	// Should be the same as original (no wrapping)
	if transport != http.DefaultTransport {
		t.Error("expected HTTPTransport to return original transport when disabled")
	}
}

func TestRecordCommand_Noop(t *testing.T) {
	t.Setenv("XBE_TELEMETRY_ENABLED", "0")

	ctx := context.Background()
	provider, err := Init(ctx)
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	// Should not panic when called on no-op provider
	provider.RecordCommand(ctx, "test", "xbe test", true, time.Second)
}

func TestRecordCommand_Enabled(t *testing.T) {
	t.Setenv("XBE_TELEMETRY_ENABLED", "1")
	t.Setenv("OTEL_TRACES_EXPORTER", "console")
	t.Setenv("OTEL_METRICS_EXPORTER", "console")

	ctx := context.Background()
	provider, err := Init(ctx)
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	defer provider.Shutdown(ctx)

	// Should not panic when called on enabled provider
	provider.RecordCommand(ctx, "test", "xbe test", true, time.Second)
	provider.RecordCommand(ctx, "test", "xbe test", false, 500*time.Millisecond)
}

func TestTracer_NoopWhenDisabled(t *testing.T) {
	t.Setenv("XBE_TELEMETRY_ENABLED", "0")

	ctx := context.Background()
	provider, err := Init(ctx)
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	// Get tracer when disabled
	tracer := provider.Tracer()

	// Should not be nil (returns no-op tracer)
	if tracer == nil {
		t.Error("Tracer() should not return nil even when disabled")
	}

	// Should be able to create spans (no-op)
	_, span := tracer.Start(ctx, "test")
	span.End() // Should not panic
}
