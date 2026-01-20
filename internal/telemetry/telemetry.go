package telemetry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Provider manages OpenTelemetry tracing and metrics for the CLI.
type Provider struct {
	noop           bool
	config         Config
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	tracer         trace.Tracer
	meter          metric.Meter
	instruments    *instruments
}

// Init initializes the telemetry provider. If telemetry is disabled,
// returns a no-op provider that has zero overhead.
func Init(ctx context.Context) (*Provider, error) {
	cfg := LoadConfig()

	// Master switch check - return no-op immediately
	if !cfg.Enabled {
		return &Provider{noop: true, config: cfg}, nil
	}

	// Create resource (service name, version, OS info)
	res, err := newResource(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: telemetry resource creation failed: %v\n", err)
		res = resource.Default()
	}

	// Create trace provider with exporter
	tp, err := newTracerProvider(ctx, cfg, res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: telemetry tracer setup failed: %v\n", err)
		return &Provider{noop: true, config: cfg}, nil
	}

	// Set as global tracer provider
	otel.SetTracerProvider(tp)

	// Create meter provider with exporter
	mp, err := newMeterProvider(ctx, cfg, res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: telemetry meter setup failed: %v\n", err)
		// Continue with traces only, no metrics
		mp = nil
	}

	if mp != nil {
		otel.SetMeterProvider(mp)
	}

	// Get tracer and meter instances
	tracer := tp.Tracer("xbe-cli")
	var mtr metric.Meter
	var inst *instruments

	if mp != nil {
		mtr = mp.Meter("xbe-cli")
		inst, err = newInstruments(mtr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: telemetry instruments creation failed: %v\n", err)
		}
	}

	return &Provider{
		noop:           false,
		config:         cfg,
		tracerProvider: tp,
		meterProvider:  mp,
		tracer:         tracer,
		meter:          mtr,
		instruments:    inst,
	}, nil
}

// Shutdown gracefully shuts down the telemetry providers.
// Respects the context deadline for flushing.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.noop {
		return nil
	}

	var errs []error

	if p.tracerProvider != nil {
		if err := p.tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("tracer shutdown: %w", err))
		}
	}

	if p.meterProvider != nil {
		if err := p.meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("meter shutdown: %w", err))
		}
	}

	return errors.Join(errs...)
}

// Enabled returns true if telemetry is active.
func (p *Provider) Enabled() bool {
	return !p.noop
}

// Tracer returns the tracer for creating spans.
// Returns a no-op tracer if telemetry is disabled.
func (p *Provider) Tracer() trace.Tracer {
	if p.noop || p.tracer == nil {
		return trace.NewNoopTracerProvider().Tracer("")
	}
	return p.tracer
}

// HTTPTransport wraps the given transport with OpenTelemetry instrumentation.
// Returns the original transport unchanged if telemetry is disabled.
func (p *Provider) HTTPTransport(base http.RoundTripper) http.RoundTripper {
	if p.noop {
		return base
	}
	return otelhttp.NewTransport(base)
}

// CommandInfo holds parsed command hierarchy information.
type CommandInfo struct {
	Name   string // The leaf command (e.g., "list")
	Group  string // The resource group (e.g., "action-items")
	Action string // The top-level action (e.g., "view")
	Path   string // Full command path (e.g., "xbe view action-items list")
}

// ParseCommandPath extracts command hierarchy from a command path.
// For "xbe view action-items list":
//   - Action = "view"
//   - Group = "action-items"
//   - Name = "list"
func ParseCommandPath(commandPath string) CommandInfo {
	parts := strings.Fields(commandPath)
	info := CommandInfo{Path: commandPath}

	switch len(parts) {
	case 0:
		// Empty path
	case 1:
		// Just "xbe"
		info.Name = parts[0]
	case 2:
		// "xbe <action>" (e.g., "xbe version")
		info.Name = parts[1]
	case 3:
		// "xbe <action> <group>" (e.g., "xbe view action-items")
		info.Action = parts[1]
		info.Name = parts[2]
	default:
		// "xbe <action> <group> <name>" or deeper
		info.Action = parts[1]
		info.Group = parts[len(parts)-2]
		info.Name = parts[len(parts)-1]
	}

	return info
}

// RecordCommand records metrics for a completed command execution.
func (p *Provider) RecordCommand(ctx context.Context, cmdInfo CommandInfo, success bool, duration time.Duration) {
	if p.noop || p.instruments == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("command.name", cmdInfo.Name),
		attribute.String("command.path", cmdInfo.Path),
		attribute.Bool("success", success),
	}

	// Add action and group if present
	if cmdInfo.Action != "" {
		attrs = append(attrs, attribute.String("command.action", cmdInfo.Action))
	}
	if cmdInfo.Group != "" {
		attrs = append(attrs, attribute.String("command.group", cmdInfo.Group))
	}

	p.instruments.commandCount.Add(ctx, 1, metric.WithAttributes(attrs...))
	p.instruments.commandDuration.Record(ctx, float64(duration.Milliseconds()), metric.WithAttributes(attrs...))
}
