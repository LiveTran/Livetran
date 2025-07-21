package config

import (
	"fmt"
	"github.com/joho/godotenv"
)





func InitEnv() {
  // load .env file once during application startup
  err := godotenv.Load(".env")
  if err != nil {
    fmt.Println("Error loading .env file",err)
  }
}
