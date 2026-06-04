// Command tasker-web starts the HTTP server. This is the wiring layer: build
// dependencies, register routes, listen. Keep business logic out of here.
package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/kennyg37/tasker/web/internal/api"
	"github.com/kennyg37/tasker/web/internal/store"
)

func main() {
	app := fiber.New(fiber.Config{AppName: "tasker-web"})
	app.Use(logger.New())

	// Health check — verify the server is alive before anything else.
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	s := store.New()
	api.New(s).Register(app)

	log.Fatal(app.Listen(":3000"))
}
