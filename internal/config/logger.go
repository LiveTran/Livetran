package config

import (
    "context"
    "io"
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

    // --- Create OTLP gRPC exporter ---
    exporter, err := otlptracegrpc.New(ctx)
    if err != nil {
        log.Fatalf("failed to create OTLP exporter: %v", err)
    }

    // --- Create Resource with service attributes ---
    res, err := resource.New(ctx,
        resource.WithSchemaURL(semconv.SchemaURL),
        resource.WithAttributes(
            attribute.String("service.name", "LiveTran"),
        ),
    )
    if err != nil {
        log.Fatalf("failed to create resource: %v", err)
    }

    // --- Create Tracer Provider ---
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
    )
    otel.SetTracerProvider(tp)

    // --- Open log file ---
    logFile, err := os.OpenFile("/tmp/logs/livetran.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatalf("failed to open log file: %v", err)
    }

    // --- MultiWriter to write both to console and file ---
    mw := io.MultiWriter(os.Stdout, logFile)

    // --- Create slog handler that writes JSON logs to both file and console ---
    handler := slog.NewJSONHandler(mw, nil)
    slog.SetDefault(slog.New(handler))

    // --- Graceful shutdown of tracer and log file ---
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
