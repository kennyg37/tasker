package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/kennyg37/tasker/web/internal/store"
	"github.com/kennyg37/tasker/web/internal/task"
)

// newTxApp builds a Fiber app whose store is backed by a transaction that rolls
// back when the test ends, so the dev database is left untouched.
func newTxApp(t *testing.T) (*fiber.App, *gorm.DB) {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping Postgres integration test")
	}
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := gdb.AutoMigrate(&task.Task{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	tx := gdb.Begin()
	t.Cleanup(func() { tx.Rollback() })

	app := fiber.New()
	New(store.New(tx), "").Register(app)
	return app, tx
}

func TestSyncIngestsTasks(t *testing.T) {
	app, tx := newTxApp(t)

	// Unique marker keeps the assertion robust against any pre-existing rows.
	marker := fmt.Sprintf("synctest-%d", time.Now().UnixNano())
	body := fmt.Sprintf("10/06/2026\n1. %s A - pending\n2. %s B - pending\n", marker, marker)

	req := httptest.NewRequest("POST", "/api/sync", strings.NewReader(body))
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body = %s", resp.StatusCode, b)
	}

	var out struct {
		Parsed   int `json:"parsed"`
		Upserted int `json:"upserted"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Parsed != 2 || out.Upserted != 2 {
		t.Fatalf("response = %+v, want parsed=2 upserted=2", out)
	}

	var count int64
	tx.Model(&task.Task{}).Where("content LIKE ?", marker+"%").Count(&count)
	if count != 2 {
		t.Fatalf("rows persisted = %d, want 2", count)
	}
}

func TestSyncRequiresTokenWhenConfigured(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping Postgres integration test")
	}
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	tx := gdb.Begin()
	t.Cleanup(func() { tx.Rollback() })

	app := fiber.New()
	New(store.New(tx), "topsecret").Register(app)

	body := "10/06/2026\n1. authed task - pending\n"

	cases := []struct {
		name   string
		header string
		want   int
	}{
		{"no token", "", 401},
		{"wrong token", "Bearer nope", 401},
		{"correct token", "Bearer topsecret", 200},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/sync", strings.NewReader(body))
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tc.want {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.want)
			}
		})
	}
}
