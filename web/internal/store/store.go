// Package store owns persistence for tasks. The DB is the single source of
// firing truth, so the upsert here is the most safety-critical code in the
// project: re-syncing the same task must update it, never duplicate it, and
// never resurrect completion or re-arm a fired reminder.
package store

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/kennyg37/tasker/web/internal/task"
)

type Store struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

// Upsert inserts t, or updates the existing row with the same source_key. Only
// the fields a re-sync is allowed to change are updated: content, due_at,
// reminder_at, updated_at. is_done and reminder_sent are deliberately excluded
// — overwriting them would resurrect completed tasks or re-fire handled
// reminders. origin and date_added are set once on insert and preserved.
func (s *Store) Upsert(t *task.Task) error {
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "source_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"content", "due_at", "reminder_at", "updated_at"}),
	}).Create(t).Error
}
