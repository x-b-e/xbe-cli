package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// newMeterProvider creates a MeterProvider with the appropriate exporter
// based on configuration.
func newMeterProvider(ctx context.Context, cfg Config, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	var reader sdkmetric.Reader
	var err error

	switch cfg.MetricsExporter {
	case "console":
		exporter, exporterErr := stdoutmetric.New(stdoutmetric.WithWriter(os.Stderr))
		if exporterErr != nil {
			return nil, fmt.Errorf("failed to create console metric exporter: %w", exporterErr)
		}
		// For console, use periodic reader with short interval for debugging
		reader = sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(time.Second))

	case "none", "":
		// No exporter, return provider that won't export
		return sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
		), nil

	case "otlp":
		fallthrough
	default:
		exporter, exporterErr := newOTLPMetricExporter(ctx, cfg)
		if exporterErr != nil {
			return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", exporterErr)
		}
		err = exporterErr
		reader = sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(10*time.Second))
	}

	if err != nil {
		return nil, err
	}

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	), nil
}

func newOTLPMetricExporter(ctx context.Context, cfg Config) (sdkmetric.Exporter, error) {
	// Use HTTP/protobuf if specified, otherwise default to gRPC
	if cfg.OTLPProtocol == "http/protobuf" {
		return newOTLPMetricHTTPExporter(ctx, cfg)
	}
	return newOTLPMetricGRPCExporter(ctx, cfg)
}

func newOTLPMetricGRPCExporter(ctx context.Context, cfg Config) (sdkmetric.Exporter, error) {
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(cfg.OTLPEndpoint),
	}

	// Only disable TLS if explicitly requested (for local development)
	if cfg.OTLPInsecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}

	// Add headers if configured
	if len(cfg.OTLPHeaders) > 0 {
		opts = append(opts, otlpmetricgrpc.WithHeaders(cfg.OTLPHeaders))
	}

	return otlpmetricgrpc.New(ctx, opts...)
}

func newOTLPMetricHTTPExporter(ctx context.Context, cfg Config) (sdkmetric.Exporter, error) {
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(cfg.OTLPEndpoint),
	}

	// Only disable TLS if explicitly requested (for local development)
	if cfg.OTLPInsecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	// Add headers if configured
	if len(cfg.OTLPHeaders) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(cfg.OTLPHeaders))
	}

	return otlpmetrichttp.New(ctx, opts...)
}

// Metric instruments for the CLI
type instruments struct {
	commandCount    metric.Int64Counter
	commandDuration metric.Float64Histogram
}

func newInstruments(meter metric.Meter) (*instruments, error) {
	cmdCount, err := meter.Int64Counter(
		"xbe.cli.command.count",
		metric.WithDescription("Number of CLI commands executed"),
		metric.WithUnit("{command}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create command counter: %w", err)
	}

	cmdDuration, err := meter.Float64Histogram(
		"xbe.cli.command.duration",
		metric.WithDescription("Duration of CLI command execution"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create command duration histogram: %w", err)
	}

	return &instruments{
		commandCount:    cmdCount,
		commandDuration: cmdDuration,
	}, nil
}
