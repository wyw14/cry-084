package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Address        string
	DatabaseURL    string
	RequestTimeout time.Duration
	FileRoot       string
}

func Load() (Config, error) {
	value := Config{Address: env("APP_ADDR", ":8080"), DatabaseURL: os.Getenv("DATABASE_URL"), FileRoot: env("FILE_ROOT", "./data/files")}
	duration, err := time.ParseDuration(env("REQUEST_TIMEOUT", "5s"))
	if err != nil {
		return Config{}, fmt.Errorf("request timeout: %w", err)
	}
	value.RequestTimeout = duration
	return value, nil
}
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
