// Command tasker-web starts the HTTP server. Wiring only: load config, build
// the Fiber app, register the health route, listen. Business logic lives in the
// internal packages.
package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/kennyg37/tasker/web/internal/api"
	"github.com/kennyg37/tasker/web/internal/config"
	"github.com/kennyg37/tasker/web/internal/db"
	"github.com/kennyg37/tasker/web/internal/draft"
	"github.com/kennyg37/tasker/web/internal/scheduler"
	"github.com/kennyg37/tasker/web/internal/store"
	"github.com/kennyg37/tasker/web/internal/task"
	"github.com/kennyg37/tasker/web/internal/telegram"
)

func main() {
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	gdb, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	if err := db.Migrate(gdb); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("database connected and migrated")

	app := fiber.New(fiber.Config{AppName: "tasker-web"})
	app.Use(logger.New())

	// Health check — verify the server is alive before anything else.
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	s := store.New(gdb)
	if cfg.SyncToken == "" {
		log.Println("WARNING: SYNC_TOKEN not set; /api/sync is UNAUTHENTICATED")
	}
	api.New(s, cfg.SyncToken).Register(app)

	ctx := context.Background()
	send := startDelivery(ctx, cfg, s)
	go scheduler.New(s, send, 30*time.Second).Run(ctx)

	log.Printf("tasker-web listening on :%s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}

// startDelivery wires Telegram outbound delivery and inbound capture when a bot
// token and chat id are configured, returning the scheduler's send function.
// Without Telegram it falls back to logging, so the server still runs in dev.
func startDelivery(ctx context.Context, cfg config.Config, s *store.Store) scheduler.Sender {
	if cfg.TelegramBotToken == "" || cfg.TelegramChatID == "" {
		log.Println("TELEGRAM_BOT_TOKEN/TELEGRAM_CHAT_ID not set; reminders log to stdout, no inbound capture")
		return func(t task.Task) error {
			log.Printf("REMINDER task %d: %s", t.ID, t.Content)
			return nil
		}
	}
	tg, err := telegram.New(cfg.TelegramBotToken, cfg.TelegramChatID)
	if err != nil {
		log.Fatalf("telegram: %v", err)
	}
	log.Println("reminders deliver via Telegram; inbound capture enabled")

	drafts := draft.New(time.Now, 10*time.Minute)
	inbound := telegram.NewInbound(tg, drafts, s.Upsert, time.Now)
	go tg.Listen(ctx, inbound)
	return tg.Send
}
