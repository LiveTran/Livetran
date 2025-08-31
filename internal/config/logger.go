package config

import (
    "context"
    "log"
    "log/slog"
    "os"
    "time"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func InitSlogOTLP() {
    ctx := context.Background()

    // Create OTLP gRPC exporter
    exporter, err := otlptracegrpc.New(ctx)
    if err != nil {
        log.Fatalf("failed to create OTLP exporter: %v", err)
    }

    // Create Resource with service attributes
    res, err := resource.New(ctx,
        resource.WithSchemaURL(semconv.SchemaURL),
        resource.WithAttributes(
            attribute.String("service.name", "LiveTran"),
        ),
    )
    if err != nil {
        log.Fatalf("failed to create resource: %v", err)
    }

    // Create Tracer Provider
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
    )
    otel.SetTracerProvider(tp)

    // --- Write slog logs to a specific file ---
    logFile, err := os.OpenFile("/Users/vijayvenkatj/Livetran/metrics/deployment/log/livetran.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatalf("failed to open log file: %v", err)
    }

    handler := slog.NewJSONHandler(logFile, nil) // JSONHandler is recommended for structured logs
    slog.SetDefault(slog.New(handler))

    // Flush tracer on exit
    go func() {
        <-ctx.Done()
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := tp.Shutdown(ctx); err != nil {
            log.Printf("Error shutting down tracer provider: %v", err)
        }
        logFile.Close()
    }()
}
