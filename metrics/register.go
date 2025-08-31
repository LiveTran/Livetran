package metrics

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)



func RegisterStatusGauge(
	ctx context.Context,
	meter metric.Meter,
	name string,
	description string,
	callbackFn func() (active, idle, stopped int64),
) {
	gauge, err := meter.Int64ObservableGauge(
		name,
		metric.WithDescription(description),
	)
	if err != nil {
		slog.Error("Error creating gauge", "error", err)
		return
	}

	_, err = meter.RegisterCallback(func(ctx context.Context, obs metric.Observer) error {
		active, idle, stopped := callbackFn()

		// Export each status with a "status" label
		obs.ObserveInt64(gauge, idle, metric.WithAttributes(attribute.String("status", "idle")))
		obs.ObserveInt64(gauge, active, metric.WithAttributes(attribute.String("status", "active")))
		obs.ObserveInt64(gauge, stopped, metric.WithAttributes(attribute.String("status", "stopped")))

		return nil
	}, gauge)

	if err != nil {
		slog.Error("Error registering callback", "error", err)
	}
}



func RegisterUpDownCounter(ctx context.Context, meter metric.Meter, name string, description string, callbackFn func() (int64, []attribute.KeyValue)) {
	
	counter, err := meter.Int64ObservableUpDownCounter(
		name,
		metric.WithDescription(description),
	)
	if err != nil {
		slog.Error("Error registering up/down counter", "error", err)
		return
	}

	_, err = meter.RegisterCallback(func(ctx context.Context, obs metric.Observer) error {
		value, attrs := callbackFn()
        obs.ObserveInt64(
            counter,
            value,
            metric.WithAttributes(attrs...), // dynamic attrs
        )
		return nil
	}, counter)
	if err != nil {
		slog.Error("Error registering counter callback", "error", err)
	}
}


func RegisterHistogram(meter metric.Meter, name string, description string) (metric.Int64Histogram, error) {
    histogram, err := meter.Int64Histogram(
        name,
        metric.WithDescription(description),
    )
    if err != nil {
        slog.Error("Error creating histogram", "error", err)
        return nil, err
    }

    return histogram, nil
}
