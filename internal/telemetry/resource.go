package telemetry

import (
	"context"
	"runtime"

	"github.com/xbe-inc/xbe-cli/internal/version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// newResource creates an OTEL resource with service and host information.
func newResource(ctx context.Context) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("xbe-cli"),
			semconv.ServiceVersion(version.String()),
			attribute.String("host.arch", runtime.GOARCH),
		),
		resource.WithOS(),
		resource.WithHost(),
	)
}
