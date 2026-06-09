// Package api wires HTTP routes to the store. Handlers stay thin: parse the
// request, call the store, return JSON.
package api

import (
	"crypto/subtle"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/kennyg37/tasker/web/internal/parser"
	"github.com/kennyg37/tasker/web/internal/store"
)

type Handler struct {
	store     *store.Store
	syncToken string
}

func New(s *store.Store, syncToken string) *Handler {
	return &Handler{store: s, syncToken: syncToken}
}

// Register mounts the routes under /api. When a sync token is configured, the
// group requires a matching bearer token.
func (h *Handler) Register(app *fiber.App) {
	api := app.Group("/api")
	if h.syncToken != "" {
		api.Use(h.requireToken)
	}
	api.Post("/sync", h.sync)
}

// requireToken rejects requests without a matching "Authorization: Bearer"
// token, compared in constant time.
func (h *Handler) requireToken(c *fiber.Ctx) error {
	got := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
	if subtle.ConstantTimeCompare([]byte(got), []byte(h.syncToken)) != 1 {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	return c.Next()
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
