package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "Relay"),
			Env:  getEnv("APP_ENV", "local"),
			Port: getEnv("APP_PORT", "8080"),
		},

		DB: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USERNAME", "relay"),
			Password: getEnv("DB_PASSWORD", "relay"),
			Name:     getEnv("DB_DATABASE", "relay"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}

	if cfg.App.Name == "" {
		return nil, fmt.Errorf("App Name is required")
	}

	return cfg, nil

}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}
