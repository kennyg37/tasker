package store

import (
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/kennyg37/tasker/web/internal/task"
)

// newTx returns a Store backed by a transaction that is rolled back when the
// test ends, so tests never leave rows behind in the dev database.
func newTx(t *testing.T) (*Store, *gorm.DB) {
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
	return New(tx), tx
}

func TestUpsertDoesNotDuplicate(t *testing.T) {
	s, tx := newTx(t)

	due := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	if err := s.Upsert(&task.Task{
		SourceKey: "pc:line-1", Content: "Write tests", Origin: task.OriginPC, DueAt: &due,
	}); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	newDue := due.Add(time.Hour)
	if err := s.Upsert(&task.Task{
		SourceKey: "pc:line-1", Content: "Write more tests", Origin: task.OriginPC, DueAt: &newDue,
	}); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	var count int64
	tx.Model(&task.Task{}).Where("source_key = ?", "pc:line-1").Count(&count)
	if count != 1 {
		t.Fatalf("row count = %d, want 1 (upsert must not duplicate on source_key)", count)
	}

	var got task.Task
	if err := tx.Where("source_key = ?", "pc:line-1").First(&got).Error; err != nil {
		t.Fatalf("read back: %v", err)
	}
	if got.Content != "Write more tests" {
		t.Errorf("Content = %q, want it updated on re-sync", got.Content)
	}
	if got.DueAt == nil || !got.DueAt.Equal(newDue) {
		t.Errorf("DueAt = %v, want %v (should be updated)", got.DueAt, newDue)
	}
}

// The single most important rule in CONTEXT.md: a re-sync must not overwrite
// is_done or reminder_sent.
func TestUpsertPreservesDoneAndSentFlags(t *testing.T) {
	s, tx := newTx(t)

	if err := s.Upsert(&task.Task{
		SourceKey: "pc:line-2", Content: "Pay rent", Origin: task.OriginPC,
	}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	// Simulate completion + a fired reminder happening through other code paths.
	if err := tx.Model(&task.Task{}).Where("source_key = ?", "pc:line-2").
		Updates(map[string]any{"is_done": true, "reminder_sent": true}).Error; err != nil {
		t.Fatalf("mark done/sent: %v", err)
	}

	// Re-syncing TODO.md sends both flags false. The upsert must ignore them.
	if err := s.Upsert(&task.Task{
		SourceKey: "pc:line-2", Content: "Pay rent (updated)", Origin: task.OriginPC,
		IsDone: false, ReminderSent: false,
	}); err != nil {
		t.Fatalf("resync: %v", err)
	}

	var got task.Task
	if err := tx.Where("source_key = ?", "pc:line-2").First(&got).Error; err != nil {
		t.Fatalf("read back: %v", err)
	}
	if !got.IsDone {
		t.Error("IsDone was reset to false — completed task resurrected")
	}
	if !got.ReminderSent {
		t.Error("ReminderSent was reset to false — a handled reminder will re-fire")
	}
	if got.Content != "Pay rent (updated)" {
		t.Errorf("Content = %q, want it updated on re-sync", got.Content)
	}
}

// DueReminders must include only due, unsent, incomplete tasks with a reminder.
func TestDueRemindersFiltering(t *testing.T) {
	s, _ := newTx(t)

	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	mk := func(key string, remindAt *time.Time, sent, done bool) {
		if err := s.Upsert(&task.Task{
			SourceKey: key, Content: key, Origin: task.OriginPC,
			ReminderAt: remindAt, ReminderSent: sent, IsDone: done,
		}); err != nil {
			t.Fatalf("seed %s: %v", key, err)
		}
	}
	mk("pc:due", &past, false, false)      // fires
	mk("pc:future", &future, false, false) // not yet due
	mk("pc:sent", &past, true, false)      // already sent
	mk("pc:done", &past, false, true)      // completed
	mk("pc:noremind", nil, false, false)   // no reminder

	due, err := s.DueReminders(now)
	if err != nil {
		t.Fatalf("DueReminders: %v", err)
	}

	got := map[string]bool{}
	for _, d := range due {
		got[d.SourceKey] = true
	}
	if !got["pc:due"] {
		t.Error("pc:due should be returned")
	}
	for _, k := range []string{"pc:future", "pc:sent", "pc:done", "pc:noremind"} {
		if got[k] {
			t.Errorf("%s should NOT be returned by DueReminders", k)
		}
	}
}

// A batch sync re-applied is idempotent: no duplicate rows.
func TestUpsertAllIsIdempotent(t *testing.T) {
	s, tx := newTx(t)

	tasks := []task.Task{
		{SourceKey: "pc:batch-1", Content: "A", Origin: task.OriginPC},
		{SourceKey: "pc:batch-2", Content: "B", Origin: task.OriginPC},
	}
	if err := s.UpsertAll(tasks); err != nil {
		t.Fatalf("first batch: %v", err)
	}
	if err := s.UpsertAll(tasks); err != nil {
		t.Fatalf("second batch: %v", err)
	}

	var count int64
	tx.Model(&task.Task{}).Where("source_key IN ?", []string{"pc:batch-1", "pc:batch-2"}).Count(&count)
	if count != 2 {
		t.Fatalf("row count = %d, want 2 (batch upsert must not duplicate)", count)
	}
}

// origin and date_added are set on insert and must survive a re-sync.
func TestUpsertPreservesOriginAndDateAdded(t *testing.T) {
	s, tx := newTx(t)

	if err := s.Upsert(&task.Task{
		SourceKey: "pc:line-3", Content: "Read book", Origin: task.OriginPC,
	}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	var first task.Task
	if err := tx.Where("source_key = ?", "pc:line-3").First(&first).Error; err != nil {
		t.Fatalf("read first: %v", err)
	}

	if err := s.Upsert(&task.Task{
		SourceKey: "pc:line-3", Content: "Read book (ch. 2)", Origin: task.OriginPhone,
	}); err != nil {
		t.Fatalf("resync: %v", err)
	}

	var got task.Task
	if err := tx.Where("source_key = ?", "pc:line-3").First(&got).Error; err != nil {
		t.Fatalf("read back: %v", err)
	}
	if got.Origin != task.OriginPC {
		t.Errorf("Origin = %q, want it preserved as %q", got.Origin, task.OriginPC)
	}
	if !got.DateAdded.Equal(first.DateAdded) {
		t.Errorf("DateAdded changed: got %v, want %v", got.DateAdded, first.DateAdded)
	}
}
