package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
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
	WebhookUrls  []string 	`json:"webhook_urls,omitempty"`
}



func (handler *Handler) StartStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var streamBody StreamRequest

	err := json.NewDecoder(r.Body).Decode(&streamBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
		Success: false,
		Error: "Cannot read Request body!",
		})
		return
	}

	handler.tm.StartTask(streamBody.StreamId,streamBody.WebhookUrls)
	
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
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
		Success: false,
		Error: "Cannot read Request body!",
		})
		return
	}

	handler.tm.StopTask(streamBody.StreamId,errors.New("user initiated request"))
	
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data:    "Stream stopped!",
	})
}

func (handler *Handler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type","application/json")

	var streamBody StreamRequest

	err := json.NewDecoder(r.Body).Decode(&streamBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
		Success: false,
		Error: "Cannot read Request body!",
		})
		return
	}

	task,exists := handler.tm.TaskMap[streamBody.StreamId]
	if exists {
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(Response{
		Success: true,
		Data: fmt.Sprintf("Status: %s",task.Status),
		})
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Error: "Task not found",
	})
}

func (handler *Handler) GetVideoChunks(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement LL-HLS 
	
	filePath := filepath.Join("output",r.URL.Path)
	
	file, err := os.Open(filePath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error: "Error accessing file",
		})
		return
	}
	defer file.Close()

	if filepath.Ext(filePath) == ".m3u8" {
        w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
    } else if filepath.Ext(filePath) == ".ts" {
        w.Header().Set("Content-Type", "video/MP2T")
    }

	w.Header().Set("Accept-Ranges", "bytes") // Allowing for partial requests to be done

	info, _ := file.Stat()
	modtime := info.ModTime() // This is to control caching
	http.ServeContent(w, r, filepath.Base(filePath), modtime , file)
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