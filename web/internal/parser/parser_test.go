package parser

import (
	"testing"
	"time"

	"github.com/kennyg37/tasker/web/internal/task"
)

const sample = `22/06/2026
1. Onway redis - Done
2. Polymorphism c++ - Postponed
3. Nicomachean ethics - Done

23/06/2026

1. Plausible - Done
2. Tenex - pending
`

func TestParseBasic(t *testing.T) {
	tasks := Parse(sample)

	if len(tasks) != 5 {
		t.Fatalf("got %d tasks, want 5", len(tasks))
	}
	for _, tk := range tasks {
		if tk.Origin != task.OriginPC {
			t.Errorf("task %q origin = %q, want pc", tk.Content, tk.Origin)
		}
		if tk.SourceKey == "" {
			t.Errorf("task %q has empty source_key", tk.Content)
		}
		if tk.ReminderAt == nil {
			t.Errorf("task %q has nil reminder_at", tk.Content)
		}
	}
}

func TestReminderAtIsDefaultHourKigaliInUTC(t *testing.T) {
	tasks := Parse("10/06/2026\n1. Submit report - pending\n")
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}
	// 09:00 Kigali (CAT, UTC+2) on 2026-06-10 == 07:00 UTC.
	want := time.Date(2026, 6, 10, 7, 0, 0, 0, time.UTC)
	got := tasks[0].ReminderAt
	if got == nil || !got.Equal(want) {
		t.Fatalf("reminder_at = %v, want %v", got, want)
	}
	if got.Location() != time.UTC {
		t.Errorf("reminder_at not stored in UTC: %v", got.Location())
	}
}

func TestIsDoneDerivedFromStatus(t *testing.T) {
	tasks := Parse("10/06/2026\n1. Done task - Done\n2. Active task - pending\n3. Later - Postponed\n")
	if len(tasks) != 3 {
		t.Fatalf("got %d tasks, want 3", len(tasks))
	}
	if !tasks[0].IsDone {
		t.Error("task marked Done should have is_done=true")
	}
	if tasks[1].IsDone {
		t.Error("pending task should have is_done=false")
	}
	if tasks[2].IsDone {
		t.Error("postponed task should have is_done=false")
	}
}

func TestSourceKeyStableAcrossNumberAndStatus(t *testing.T) {
	// Same date + description, but reordered (different number) and a different
	// status, must produce the same source_key.
	a := Parse("10/06/2026\n1. Buy milk - pending\n")
	b := Parse("10/06/2026\n5. Buy milk - Done\n")
	if a[0].SourceKey != b[0].SourceKey {
		t.Errorf("source_key changed on number/status edit: %q vs %q", a[0].SourceKey, b[0].SourceKey)
	}
}

func TestSourceKeyStableAcrossDateFormat(t *testing.T) {
	a := Parse("1/6/2026\n1. Buy milk - pending\n")
	b := Parse("01/06/2026\n1. Buy milk - pending\n")
	if a[0].SourceKey != b[0].SourceKey {
		t.Errorf("source_key changed on date-format variant: %q vs %q", a[0].SourceKey, b[0].SourceKey)
	}
}

func TestSourceKeyChangesOnDescriptionEdit(t *testing.T) {
	a := Parse("10/06/2026\n1. Buy milk - pending\n")
	b := Parse("10/06/2026\n1. Buy oat milk - pending\n")
	if a[0].SourceKey == b[0].SourceKey {
		t.Error("source_key should change when description changes")
	}
}

func TestLinesBeforeAnyDateHeaderIgnored(t *testing.T) {
	tasks := Parse("1. Orphan task - pending\n10/06/2026\n1. Real task - pending\n")
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1 (line before any date header must be ignored)", len(tasks))
	}
	if tasks[0].Content != "Real task" {
		t.Errorf("content = %q, want %q", tasks[0].Content, "Real task")
	}
}

func TestDescriptionWithInternalDashKept(t *testing.T) {
	// "commerce setup" is not a recognized status, so the dash stays in the
	// description rather than being treated as a separator.
	tasks := Parse("10/06/2026\n1. e-commerce setup\n")
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}
	if tasks[0].Content != "e-commerce setup" {
		t.Errorf("content = %q, want %q", tasks[0].Content, "e-commerce setup")
	}
}

func TestNonTaskLinesIgnored(t *testing.T) {
	tasks := Parse("10/06/2026\njust a note, no number\n- bullet without number\n1. Real - pending\n")
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}
}
