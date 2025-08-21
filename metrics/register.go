package metrics

import (
	"context"
	"log/slog"

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