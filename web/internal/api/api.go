// Package api wires HTTP routes to the store. Handlers stay thin: parse the
// request, call the store, return JSON.
package api

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/kennyg37/tasker/web/internal/parser"
	"github.com/kennyg37/tasker/web/internal/store"
)

type Handler struct {
	store *store.Store
}

func New(s *store.Store) *Handler {
	return &Handler{store: s}
}

// Register mounts the routes under /api.
func (h *Handler) Register(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/sync", h.sync)
}

// sync ingests a raw TODO.md body pushed by the PC watcher: parse it, then
// upsert every task in one transaction.
func (h *Handler) sync(c *fiber.Ctx) error {
	tasks := parser.Parse(string(c.Body()))
	if err := h.store.UpsertAll(tasks); err != nil {
		log.Printf("sync: upsert failed: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "upsert failed")
	}
	return c.JSON(fiber.Map{"parsed": len(tasks), "upserted": len(tasks)})
}
