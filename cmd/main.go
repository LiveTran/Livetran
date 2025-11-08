package main

import (
	"log/slog"

	"github.com/vijayvenkatj/LiveTran/internal/config"
	api "github.com/vijayvenkatj/LiveTran/internal/http"
	"github.com/vijayvenkatj/LiveTran/internal/ingest"
)

var tm *ingest.TaskManager



func init() {
	tm = ingest.NewTaskManager()
	config.InitEnv()

	config.InitSlogOTLP()

}

func main() {
	apiServer := api.NewAPIServer(":8080")
	err := apiServer.StartAPIServer(tm);
	if err != nil {
		slog.Error("SERVER STARTUP", "error", err)
		return
	}
}