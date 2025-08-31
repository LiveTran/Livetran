package main

import (
	"log/slog"
	"time"

	"github.com/vijayvenkatj/LiveTran/internal/config"
	api "github.com/vijayvenkatj/LiveTran/internal/http"
	"github.com/vijayvenkatj/LiveTran/internal/ingest"
)

var tm *ingest.TaskManager



func init() {
	tm = ingest.NewTaskManager()
	config.InitEnv()

	config.InitSlogOTLP()

	slog.Info("App started")
    slog.Error("Test error")
    slog.Warn("Test warning")
    
    time.Sleep(2 * time.Second) // Give time for logs to flush
    
    slog.Info("App ending")
}

func main() {
	apiServer := api.NewAPIServer(":8080")
	err := apiServer.StartAPIServer(tm);
	if err != nil {
		slog.Error("SERVER STARTUP", "error", err)
		return
	}
}