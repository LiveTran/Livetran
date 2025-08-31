package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vijayvenkatj/LiveTran/internal/http/handlers"
	"github.com/vijayvenkatj/LiveTran/internal/ingest"
	"github.com/vijayvenkatj/LiveTran/metrics"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric"
)



type APIServer struct {
	address string
}


// Constructor for APIServer
func NewAPIServer(address string) *APIServer {
	return &APIServer{
		address: address,
	}
}


func (a *APIServer) StartAPIServer(tm *ingest.TaskManager) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	isMetricsEnabled := metrics.IsMetricsEnabled()

	var mp metric.MeterProvider
	var shutdownFunc func()

	if isMetricsEnabled {
		var err error
		mp, err = metrics.InitMeterProvider(ctx)
		if err != nil {
			return fmt.Errorf("init metrics: %w", err)
		}
		// graceful shutdown for metrics
		shutdownFunc = func() {
			shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := mp.(interface{ Shutdown(context.Context) error }).Shutdown(shCtx); err != nil {
				slog.Error("failed to shutdown metrics", "error", err)
			}
		}
		defer shutdownFunc()

		meter := mp.Meter("live-streaming-api")

		metrics.RegisterStatusGauge(ctx, meter, "streams_info", "stream information based on status", func() (active,idle,stopped int64) {
			return tm.GetAllStreams()
		})
	}

	routeHandler := handlers.NewHandler(tm)
	streamRoutes := routeHandler.StreamRoutes()
	videoRoutes := routeHandler.VideoRoutes()

	router := http.NewServeMux()
	router.Handle("/api/", http.StripPrefix("/api", streamRoutes))
	router.Handle("/video/", http.StripPrefix("/video", videoRoutes))

	var handler http.Handler = router
	if isMetricsEnabled {
		handler = otelhttp.NewHandler(router, "http.server",
			otelhttp.WithMeterProvider(mp),
		)
	}

	server := &http.Server{
		Addr:      a.address,
		Handler:   handler,
		TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}

	// Goroutine: shutdown on signal
	go func() {
		<-ctx.Done()
		slog.Info("shutting down gracefully...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown error", "Error", err)
		}
	}()

	slog.Info(fmt.Sprintf("Server is listening on %s", a.address))
	if err := server.ListenAndServeTLS("keys/localhost.pem", "keys/localhost-key.pem"); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}
