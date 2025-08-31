package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Data    string `json:"data,omitempty"`
}

type StreamRequest struct {
	StreamId	string	    `json:"stream_id"`
	WebhookUrls []string 	`json:"webhook_urls,omitempty"`
	Abr			bool		`json:"abr,omitempty"`
}


func (handler *Handler) StartStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var streamBody StreamRequest
	err := json.NewDecoder(r.Body).Decode(&streamBody)
	if err != nil {
		slog.Error("failed to decode start stream request body",
			"error", err,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.Header.Get("User-Agent"),
		)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error:   "Cannot read Request body!",
		})
		return
	}

	slog.Info("received start stream request",
		"stream_id", streamBody.StreamId,
		"webhook_urls", streamBody.WebhookUrls,
		"abr", streamBody.Abr,
		"remote_addr", r.RemoteAddr,
		"user_agent", r.Header.Get("User-Agent"),
	)

	handler.tm.StartTask(streamBody.StreamId, streamBody.WebhookUrls, streamBody.Abr)

	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data:    "Stream launching!",
	})
}

func (handler *Handler) StopStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var streamBody StreamRequest
	err := json.NewDecoder(r.Body).Decode(&streamBody)
	if err != nil {
		slog.Error("failed to decode stop stream request body",
			"error", err,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.Header.Get("User-Agent"),
		)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error:   "Cannot read Request body!",
		})
		return
	}

	slog.Info("received stop stream request",
		"stream_id", streamBody.StreamId,
		"remote_addr", r.RemoteAddr,
		"user_agent", r.Header.Get("User-Agent"),
	)

	handler.tm.StopTask(streamBody.StreamId, errors.New("user initiated request"))

	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data:    "Stream stopped!",
	})
}

func (handler *Handler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var streamBody StreamRequest
	err := json.NewDecoder(r.Body).Decode(&streamBody)
	if err != nil {
		slog.Error("failed to decode status request body",
			"error", err,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.Header.Get("User-Agent"),
		)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error:   "Cannot read Request body!",
		})
		return
	}

	slog.Info("received status request",
		"stream_id", streamBody.StreamId,
		"remote_addr", r.RemoteAddr,
		"user_agent", r.Header.Get("User-Agent"),
	)

	task, exists := handler.tm.TaskMap[streamBody.StreamId]
	if exists {
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(Response{
			Success: true,
			Data:    fmt.Sprintf("Status: %s", task.Status),
		})
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Error:   "Task not found",
	})
}

func (handler *Handler) GetVideoChunks(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join("output", r.URL.Path)

	slog.Info("video chunk requested",
		"path", r.URL.Path,
		"resolved_file", filePath,
		"remote_addr", r.RemoteAddr,
		"user_agent", r.Header.Get("User-Agent"),
	)

	file, err := os.Open(filePath)
	if err != nil {
		slog.Error("failed to open video chunk",
			"path", filePath,
			"error", err,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.Header.Get("User-Agent"),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error:   "Error accessing file",
		})
		return
	}
	defer file.Close()

	if filepath.Ext(filePath) == ".m3u8" {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	} else if filepath.Ext(filePath) == ".ts" {
		w.Header().Set("Content-Type", "video/MP2T")
	}

	w.Header().Set("Accept-Ranges", "bytes")

	info, _ := file.Stat()
	modtime := info.ModTime()

	http.ServeContent(w, r, filepath.Base(filePath), modtime, file)
}

/*
	TODO:
		To Make a basic stream management with In-Memory DB 
		To add JWT for auth and StreamKey for validation
			-> Have a JWT secret key, Validate it against client's key. 
			-> If it succeeds then go for connection based on the streamId
			-> Use USER-API_SECRET for this
		To add AUTH (later)
*/