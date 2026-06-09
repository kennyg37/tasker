// Package config loads runtime settings from environment variables. Secrets
// (DB credentials, bot token) come from the environment only — never hardcoded.
package config

import "os"

const (
	defaultPort   = "3000"
	defaultAPIURL = "http://localhost:5001"
)

type Config struct {
	Port        string
	DatabaseURL string

	// Telegram delivery (server).
	TelegramBotToken string
	TelegramChatID   string

	// PC watcher.
	TodoFilePath string
	APIURL       string

	// Shared secret protecting /api/sync (server checks it, watcher sends it).
	SyncToken string
}

// Load reads configuration from the environment, defaulting values that are
// safe to default. It does not open or validate any connection.
func Load() Config {
	return Config{
		Port:             getenv("PORT", defaultPort),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
		TodoFilePath:     os.Getenv("TODO_FILE_PATH"),
		APIURL:           getenv("TASKER_API_URL", defaultAPIURL),
		SyncToken:        os.Getenv("SYNC_TOKEN"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
