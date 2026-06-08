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

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.DatabaseURL != "postgres://localhost/tasker" {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://localhost/tasker")
	}
}
