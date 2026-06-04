// Package api wires HTTP routes to the store. Handlers stay thin: parse the
// request, call the store, return JSON.
package api

import (
	"github.com/gofiber/fiber/v2"

	"github.com/kennyg37/tasker/web/internal/store"
	"github.com/kennyg37/tasker/web/internal/task"
)

type Handler struct {
	store *store.Store
}

func New(s *store.Store) *Handler {
	return &Handler{store: s}
}

// Register mounts the task routes under /api.
func (h *Handler) Register(app *fiber.App) {
	api := app.Group("/api")
	api.Get("/tasks", h.listTasks)
	api.Post("/tasks", h.createTask)
}

func (h *Handler) listTasks(c *fiber.Ctx) error {
	return c.JSON(h.store.All())
}

func (h *Handler) createTask(c *fiber.Ctx) error {
	var t task.Task
	if err := c.BodyParser(&t); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task payload")
	}
	if t.Status == "" {
		t.Status = task.StatusPending
	}
	return c.Status(fiber.StatusCreated).JSON(h.store.Add(t))
}
