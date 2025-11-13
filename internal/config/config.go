package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
	JWTSecret   string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (for local development)
	// Ignore error if .env doesn't exist
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}
