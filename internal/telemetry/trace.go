package telemetry

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// newTracerProvider creates a TracerProvider with the appropriate exporter
// based on configuration.
func newTracerProvider(ctx context.Context, cfg Config, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	var exporter sdktrace.SpanExporter
	var err error

	switch cfg.TracesExporter {
	case "console":
		exporter, err = stdouttrace.New(stdouttrace.WithWriter(os.Stderr))
		if err != nil {
			return nil, fmt.Errorf("failed to create console trace exporter: %w", err)
		}

	case "none", "":
		// No exporter, return provider with no-op behavior
		return sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
		), nil

	case "otlp":
		fallthrough
	default:
		exporter, err = newOTLPTraceExporter(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
		}
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	), nil
}

func newOTLPTraceExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	// Use HTTP/protobuf if specified, otherwise default to gRPC
	if cfg.OTLPProtocol == "http/protobuf" {
		return newOTLPTraceHTTPExporter(ctx, cfg)
	}
	return newOTLPTraceGRPCExporter(ctx, cfg)
}

func newOTLPTraceGRPCExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
	}

	// Only disable TLS if explicitly requested (for local development)
	if cfg.OTLPInsecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	// Add headers if configured
	if len(cfg.OTLPHeaders) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(cfg.OTLPHeaders))
	}

	return otlptracegrpc.New(ctx, opts...)
}

func newOTLPTraceHTTPExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
	}

	// Only disable TLS if explicitly requested (for local development)
	if cfg.OTLPInsecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	// Add headers if configured
	if len(cfg.OTLPHeaders) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(cfg.OTLPHeaders))
	}

	return otlptracehttp.New(ctx, opts...)
}
