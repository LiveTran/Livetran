package config

import (
	"log/slog"
	"os"
)


func InitSlog() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout,nil));
	slog.SetDefault(logger);
}