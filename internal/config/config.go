package config

import (
	"log/slog"

	"github.com/joho/godotenv"
)





func InitEnv() {
  // load .env file once during application startup
  err := godotenv.Load(".env")
  if err != nil {
    slog.Error("Load .ENV", "error", err);
  }
}
