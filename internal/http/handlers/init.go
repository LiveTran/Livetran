package handlers

import (
	"net/http"

	"github.com/vijayvenkatj/LiveTran/internal/http/middlewares"
	"github.com/vijayvenkatj/LiveTran/internal/ingest"
)


type Handler struct {
	tm		*ingest.TaskManager
}


// Constructor for Handler
func NewHandler(tm *ingest.TaskManager) *Handler {
	return &Handler{
		tm: tm,
	}
}

// Sub-Router for Stream APIs
func (h *Handler) StreamRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /start-stream",h.StartStream)
	mux.HandleFunc("POST /stop-stream",h.StopStream)
	mux.HandleFunc("POST /status", h.Status)

	handler := middlewares.CORSMiddleware(middlewares.VerifyRequest(mux))

	return handler
}

func (h *Handler) VideoRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /",h.GetVideoChunks)

	handler := middlewares.CORSMiddleware(mux)

	return handler
}
/*
	This File manages all the SubRouters 
	Flow : 
		APIServer => RouteHandler => Individual Route Handlers based on Path
*/

