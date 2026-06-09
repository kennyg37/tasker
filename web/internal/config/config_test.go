package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "")

	cfg := Load()

	if cfg.Port != defaultPort {
		t.Errorf("Port = %q, want %q", cfg.Port, defaultPort)
	}
	if cfg.DatabaseURL != "" {
		t.Errorf("DatabaseURL = %q, want empty", cfg.DatabaseURL)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/tasker")
	t.Setenv("TELEGRAM_BOT_TOKEN", "tok")
	t.Setenv("TELEGRAM_CHAT_ID", "123")
	t.Setenv("TODO_FILE_PATH", "/home/me/TODO.md")
	t.Setenv("TASKER_API_URL", "http://api:9000")
	t.Setenv("SYNC_TOKEN", "s3cret")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.DatabaseURL != "postgres://localhost/tasker" {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://localhost/tasker")
	}
	if cfg.TelegramBotToken != "tok" || cfg.TelegramChatID != "123" {
		t.Errorf("telegram = %q/%q, want tok/123", cfg.TelegramBotToken, cfg.TelegramChatID)
	}
	if cfg.TodoFilePath != "/home/me/TODO.md" {
		t.Errorf("TodoFilePath = %q", cfg.TodoFilePath)
	}
	if cfg.APIURL != "http://api:9000" {
		t.Errorf("APIURL = %q", cfg.APIURL)
	}
	if cfg.SyncToken != "s3cret" {
		t.Errorf("SyncToken = %q", cfg.SyncToken)
	}
}

func TestAPIURLDefault(t *testing.T) {
	t.Setenv("TASKER_API_URL", "")
	if got := Load().APIURL; got != defaultAPIURL {
		t.Errorf("APIURL = %q, want default %q", got, defaultAPIURL)
	}
}
