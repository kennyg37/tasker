// Package task defines the core domain model, mirroring the C++ Task class.
package task

// Status mirrors the TaskStatus enum from include/Task.h.
type Status string

const (
	StatusDone      Status = "done"
	StatusPostponed Status = "postponed"
	StatusPending   Status = "pending"
	StatusOther     Status = "other"
)

// Task is the web equivalent of the C++ Task. The optional RemindAt field is
// new: the CLI tracked only a date, but reminders need a concrete time.
type Task struct {
	ID          int        `json:"id"`
	Date        string     `json:"date"`
	Number      int        `json:"number"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	RemindAt    *string    `json:"remind_at,omitempty"` // RFC3339; nil means no reminder
}
