// Package scheduler fires due reminders. It is a plain polling loop, not a job
// queue: the task row is the single source of firing truth. No LLM is ever in
// this path — firing is deterministic.
package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/kennyg37/tasker/web/internal/store"
	"github.com/kennyg37/tasker/web/internal/task"
)

// Sender delivers a reminder. It is the seam between deterministic firing and
// the delivery mechanism (Telegram). A non-nil error leaves the task unsent so
// the next tick retries it.
type Sender func(task.Task) error

type Scheduler struct {
	store    *store.Store
	send     Sender
	interval time.Duration
}

func New(s *store.Store, send Sender, interval time.Duration) *Scheduler {
	return &Scheduler{store: s, send: send, interval: interval}
}

// Run polls until ctx is cancelled. Ticks run sequentially, so reminders are
// never processed concurrently.
func (sc *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(sc.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := sc.tick(time.Now().UTC()); err != nil {
				log.Printf("scheduler: tick failed: %v", err)
			}
		}
	}
}

// tick fires each due reminder once. On a send failure the task is left unmarked
// so the next tick retries; reminder_sent is set only after a successful send.
func (sc *Scheduler) tick(now time.Time) error {
	due, err := sc.store.DueReminders(now)
	if err != nil {
		return err
	}
	for _, t := range due {
		if err := sc.send(t); err != nil {
			log.Printf("scheduler: send task %d failed, will retry: %v", t.ID, err)
			continue
		}
		if err := sc.store.MarkReminderSent(t.ID); err != nil {
			log.Printf("scheduler: mark task %d sent failed: %v", t.ID, err)
		}
	}
	return nil
}
