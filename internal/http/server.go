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

	"github.com/vijayvenkatj/LiveTran/internal/config"
	"github.com/vijayvenkatj/LiveTran/internal/http/handlers"
	"github.com/vijayvenkatj/LiveTran/internal/ingest"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

	mp, err := config.InitMeterProvider(ctx)
	if err != nil {
		return fmt.Errorf("init metrics: %w", err)
	}
	defer func() { // ensure metrics are flushed on shutdown
		shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mp.Shutdown(shCtx)
	}()

	routeHandler := handlers.NewHandler(tm)

	streamRoutes := routeHandler.StreamRoutes()
	videoRoutes := routeHandler.VideoRoutes()

	router := http.NewServeMux()
	router.Handle("/api/",http.StripPrefix("/api",streamRoutes))
	router.Handle("/video/",http.StripPrefix("/video",videoRoutes))

	instrumented := otelhttp.NewHandler(router, "http.server",
		otelhttp.WithMeterProvider(mp),
	)

	server := &http.Server{
		Addr:      a.address,
		Handler:   instrumented,
		TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}

	// Goroutine: shutdown on signal
	go func() {
		<-ctx.Done()
		slog.Info("shutting down gracefully...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown error", "err", err)
		}
	}()

	slog.Info(fmt.Sprintf("Server is listening on %s", a.address))
	if err := server.ListenAndServeTLS("keys/localhost.pem", "keys/localhost-key.pem"); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}