package main

import (
	"testing"

	"github.com/kennyg37/tasker/web/internal/task"
)

func TestPhoneTask(t *testing.T) {
	got := phoneTask(42, "buy milk")

	if got.SourceKey != "phone:42" {
		t.Errorf("SourceKey = %q, want %q", got.SourceKey, "phone:42")
	}
	if got.Origin != task.OriginPhone {
		t.Errorf("Origin = %q, want %q", got.Origin, task.OriginPhone)
	}
	if got.Content != "buy milk" {
		t.Errorf("Content = %q, want %q", got.Content, "buy milk")
	}
	if got.ReminderAt != nil {
		t.Errorf("ReminderAt = %v, want nil (no reminder for phone capture yet)", got.ReminderAt)
	}
}
