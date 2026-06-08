// Package task defines the tasks table — the single source of truth for
// reminders. Fields are declared explicitly; gorm.Model is intentionally NOT
// embedded, because its soft-delete DeletedAt column would conflict with the
// explicit is_done column.
package task

import "time"

// Origin records where a task was captured from.
type Origin string

const (
	OriginPC    Origin = "pc"
	OriginPhone Origin = "phone"
)

// Task mirrors the schema in CONTEXT.md. All datetimes are timestamptz, stored
// in UTC; local time is converted to UTC at the capture layer before saving.
type Task struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	SourceKey      string     `gorm:"uniqueIndex;not null" json:"source_key"`
	Content        string     `gorm:"not null" json:"content"`
	Origin         Origin     `gorm:"not null" json:"origin"`
	DateAdded      time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"date_added"`
	DueAt          *time.Time `gorm:"type:timestamptz" json:"due_at,omitempty"`
	ReminderAt     *time.Time `gorm:"type:timestamptz" json:"reminder_at,omitempty"`
	ReminderBefore *string    `gorm:"type:interval" json:"reminder_before,omitempty"`
	ReminderSent   bool       `gorm:"not null;default:false" json:"reminder_sent"`
	IsDone         bool       `gorm:"not null;default:false" json:"is_done"`
	UpdatedAt      time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"updated_at"`
}

func (Task) TableName() string { return "tasks" }
