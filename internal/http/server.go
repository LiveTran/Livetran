package api

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vijayvenkatj/LiveTran/internal/http/handlers"
	"github.com/vijayvenkatj/LiveTran/internal/ingest"
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

	routeHandler := handlers.NewHandler(tm)

	streamRoutes := routeHandler.StreamRoutes()
	videoRoutes := routeHandler.VideoRoutes()

	router := http.NewServeMux()
	router.Handle("/api/",http.StripPrefix("/api",streamRoutes))
	router.Handle("/video/",http.StripPrefix("/video",videoRoutes))

	slog.Info(fmt.Sprintf("Server is listening on port %s",a.address))

	server := &http.Server{
		Addr:    a.address,
		Handler: router,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			// HTTP/2 is enabled by default in Go with TLS >= 1.2 ( We need this for efficient HLS )
		},
	}

	return server.ListenAndServeTLS("keys/localhost.pem", "keys/localhost-key.pem") // TODO: Replace with valid prod secrets

}