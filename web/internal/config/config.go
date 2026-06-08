// Package config loads runtime settings from environment variables. Secrets
// (DB credentials, bot token) come from the environment only — never hardcoded.
package config

import "os"

const defaultPort = "3000"

type Config struct {
	Port        string
	DatabaseURL string
}

// Load reads configuration from the environment, defaulting values that are
// safe to default. It does not open or validate any connection.
func Load() Config {
	return Config{
		Port:        getenv("PORT", defaultPort),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
