package scheduler

import (
	"errors"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/kennyg37/tasker/web/internal/store"
	"github.com/kennyg37/tasker/web/internal/task"
)

func newTxStore(t *testing.T) (*store.Store, *gorm.DB) {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping Postgres integration test")
	}
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := gdb.AutoMigrate(&task.Task{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	tx := gdb.Begin()
	t.Cleanup(func() { tx.Rollback() })
	return store.New(tx), tx
}

func seedDue(t *testing.T, s *store.Store, key string) *task.Task {
	t.Helper()
	past := time.Now().UTC().Add(-time.Hour)
	tk := &task.Task{SourceKey: key, Content: key, Origin: task.OriginPC, ReminderAt: &past}
	if err := s.Upsert(tk); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return tk
}

func TestTickSendsAndMarksSent(t *testing.T) {
	s, tx := newTxStore(t)
	tk := seedDue(t, s, "pc:sched-success")

	sent := map[uint]bool{}
	sc := New(s, func(x task.Task) error { sent[x.ID] = true; return nil }, time.Minute)

	if err := sc.tick(time.Now().UTC()); err != nil {
		t.Fatalf("tick: %v", err)
	}
	if !sent[tk.ID] {
		t.Fatal("sender was not called for the due task")
	}

	var got task.Task
	tx.First(&got, tk.ID)
	if !got.ReminderSent {
		t.Error("reminder_sent should be true after a successful send")
	}
}

func TestTickLeavesUnmarkedOnSendFailure(t *testing.T) {
	s, tx := newTxStore(t)
	tk := seedDue(t, s, "pc:sched-fail")

	sc := New(s, func(task.Task) error { return errors.New("delivery down") }, time.Minute)

	if err := sc.tick(time.Now().UTC()); err != nil {
		t.Fatalf("tick: %v", err)
	}

	var got task.Task
	tx.First(&got, tk.ID)
	if got.ReminderSent {
		t.Error("reminder_sent must stay false when the send fails, so it retries")
	}
}

func TestTickDoesNotRefireSentTask(t *testing.T) {
	s, tx := newTxStore(t)
	tk := seedDue(t, s, "pc:sched-once")

	calls := 0
	sc := New(s, func(x task.Task) error {
		if x.ID == tk.ID {
			calls++
		}
		return nil
	}, time.Minute)

	now := time.Now().UTC()
	_ = sc.tick(now)
	_ = sc.tick(now) // second tick: task is now marked sent

	if calls != 1 {
		t.Errorf("sender called %d times for one reminder, want 1", calls)
	}
	var got task.Task
	tx.First(&got, tk.ID)
	if !got.ReminderSent {
		t.Error("reminder_sent should remain true")
	}
}
