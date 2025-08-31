package metrics

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)



func RegisterGauge(ctx context.Context, meter metric.Meter, name string, description string, callbackFn func() int64) {

	gauge, err := meter.Int64ObservableGauge(
		name,
		metric.WithDescription(description),
	)
	if err != nil {
		slog.Error("Error registering gauge", "Error", err);
		return
	}

	_, err = meter.RegisterCallback(func(ctx context.Context, obs metric.Observer) error {
		obs.ObserveInt64(gauge, callbackFn())
		return nil
	}, gauge)
	if err != nil {
		slog.Error("Error registering callback", "Error", err);
	}

}


func RegisterUpDownCounter(ctx context.Context, meter metric.Meter, name string, description string, callbackFn func() (int64, []attribute.KeyValue)) {
	
	counter, err := meter.Int64ObservableUpDownCounter(
		name,
		metric.WithDescription(description),
	)
	if err != nil {
		slog.Error("Error registering up/down counter", "Error", err)
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
		slog.Error("Error registering counter callback", "Error", err)
	}
}


func RegisterHistogram(meter metric.Meter, name string, description string) (metric.Int64Histogram, error) {
    histogram, err := meter.Int64Histogram(
        name,
        metric.WithDescription(description),
    )
    if err != nil {
        slog.Error("Error creating histogram", "Error", err)
        return nil, err
    }

    return histogram, nil
}
