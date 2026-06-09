// Package store owns persistence for tasks. The DB is the single source of
// firing truth, so the upsert here is the most safety-critical code in the
// project: re-syncing the same task must update it, never duplicate it, and
// never resurrect completion or re-arm a fired reminder.
package store

import (
	"time"

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
	return upsert(s.db, t)
}

// UpsertAll applies a batch of tasks in a single transaction: either every task
// is upserted, or none are. A re-synced TODO.md must never land half-applied.
func (s *Store) UpsertAll(tasks []task.Task) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for i := range tasks {
			if err := upsert(tx, &tasks[i]); err != nil {
				return err
			}
		}
		return nil
	})
}

func upsert(db *gorm.DB, t *task.Task) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "source_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"content", "due_at", "reminder_at", "updated_at"}),
	}).Create(t).Error
}

// DueReminders returns tasks whose reminder is due and not yet handled. This is
// the scheduler's firing predicate — it must never include a sent, completed,
// or reminder-less task.
func (s *Store) DueReminders(now time.Time) ([]task.Task, error) {
	var tasks []task.Task
	err := s.db.
		Where("reminder_at IS NOT NULL AND reminder_at <= ? AND reminder_sent = false AND is_done = false", now).
		Find(&tasks).Error
	return tasks, err
}

// MarkReminderSent records that a reminder fired, so the scheduler never
// re-fires it.
func (s *Store) MarkReminderSent(id uint) error {
	return s.db.Model(&task.Task{}).Where("id = ?", id).Update("reminder_sent", true).Error
}
