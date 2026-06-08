// Command tasker-web starts the HTTP server. Wiring only: load config, build
// the Fiber app, register the health route, listen. Business logic lives in the
// internal packages.
package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/kennyg37/tasker/web/internal/config"
	"github.com/kennyg37/tasker/web/internal/db"
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

	log.Printf("tasker-web listening on :%s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
