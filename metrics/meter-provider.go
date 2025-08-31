package metrics

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv/v1.26.0"
)

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// isMetricsEnabled checks if metrics are enabled via environment variable
func IsMetricsEnabled() bool {
	return os.Getenv("ENABLE_METRICS") == "true"
}

// InitMeterProvider configures OTLP metrics export over HTTP.
// Returns nil if metrics are disabled via ENABLE_METRICS=false
func InitMeterProvider(ctx context.Context) (*metric.MeterProvider, error) {
	// Check if metrics are enabled
	if !IsMetricsEnabled() { 
		return nil,nil
	}

	endpoint := getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318")
	insecure := getenv("OTEL_EXPORTER_OTLP_INSECURE", "true") == "true"

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(endpoint),
	}
	if insecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	exp, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	// Attach service/resource attributes
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("live-tran-api"),
			semconv.ServiceVersionKey.String(getenv("SERVICE_VERSION", "0.1.0")),
			attribute.String("env", getenv("ENV", "dev")),
		),
	)
	if err != nil {
		return nil, err
	}

	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(
			metric.NewPeriodicReader(exp, metric.WithInterval(5*time.Second)),
		),
	)

	otel.SetMeterProvider(mp)
	return mp, nil
}

// ShutdownMeterProvider gracefully shuts down the meter provider.
func ShutdownMeterProvider(ctx context.Context, mp *metric.MeterProvider) error {
	if mp != nil {
		return mp.Shutdown(ctx)
	}
	return nil
}