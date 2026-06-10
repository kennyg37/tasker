package telegram

import (
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/kennyg37/tasker/web/internal/draft"
	"github.com/kennyg37/tasker/web/internal/task"
)

type fakeReplier struct {
	texts        []string
	callbacks    []string
	buttons      int
	forceReplies int
}

func (f *fakeReplier) SendText(chatID int64, text string) error {
	f.texts = append(f.texts, text)
	return nil
}
func (f *fakeReplier) SendButtons(chatID int64, text string, kb tgbotapi.InlineKeyboardMarkup) error {
	f.buttons++
	return nil
}
func (f *fakeReplier) AnswerCallback(callbackID, text string) error {
	f.callbacks = append(f.callbacks, text)
	return nil
}
func (f *fakeReplier) SendForceReply(chatID int64, prompt string) error {
	f.forceReplies++
	return nil
}

func newTestInbound() (*Inbound, *fakeReplier, *[]task.Task) {
	now := func() time.Time { return time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC) }
	var committed []task.Task
	commit := func(tk *task.Task) error { committed = append(committed, *tk); return nil }
	r := &fakeReplier{}
	in := NewInbound(r, draft.New(now, 10*time.Minute), commit, now)
	return in, r, &committed
}

func TestStartShowsButtons(t *testing.T) {
	in, r, committed := newTestInbound()
	in.OnMessage(1, 100, "buy milk")
	if r.buttons != 1 {
		t.Errorf("buttons = %d, want 1", r.buttons)
	}
	if len(*committed) != 0 {
		t.Error("nothing should commit before a time is chosen")
	}
}

func TestOneDraftGate(t *testing.T) {
	in, r, committed := newTestInbound()
	in.OnMessage(1, 100, "buy milk")
	in.OnMessage(1, 101, "second thing")
	if r.buttons != 1 {
		t.Errorf("buttons = %d, want 1 (no second draft)", r.buttons)
	}
	if len(r.texts) != 1 {
		t.Errorf("expected one gate reply, got %d", len(r.texts))
	}
	if len(*committed) != 0 {
		t.Error("gate must not commit")
	}
}

func TestPresetCommit(t *testing.T) {
	in, r, committed := newTestInbound()
	in.OnMessage(1, 100, "buy milk")
	in.OnCallback(1, "cb1", "t:1h")

	if len(*committed) != 1 {
		t.Fatalf("commits = %d, want 1", len(*committed))
	}
	tk := (*committed)[0]
	if tk.SourceKey != "phone:100" || tk.Origin != task.OriginPhone {
		t.Errorf("task = %+v", tk)
	}
	want := time.Date(2026, 6, 10, 13, 0, 0, 0, time.UTC)
	if tk.ReminderAt == nil || !tk.ReminderAt.Equal(want) {
		t.Errorf("ReminderAt = %v, want %v", tk.ReminderAt, want)
	}
	// Draft must be cleared: a new message starts a fresh draft (buttons again).
	in.OnMessage(1, 102, "next")
	if r.buttons != 2 {
		t.Error("draft not cleared after commit")
	}
}

func TestNoneCommit(t *testing.T) {
	in, _, committed := newTestInbound()
	in.OnMessage(1, 100, "buy milk")
	in.OnCallback(1, "cb", "t:none")
	if len(*committed) != 1 {
		t.Fatalf("commits = %d, want 1", len(*committed))
	}
	if (*committed)[0].ReminderAt != nil {
		t.Error("t:none must commit with nil ReminderAt")
	}
}

func TestCancelButtonAndCommand(t *testing.T) {
	in, r, committed := newTestInbound()

	in.OnMessage(1, 100, "x")
	in.OnCallback(1, "cb", "t:cancel")
	if len(*committed) != 0 {
		t.Error("cancel must not commit")
	}
	in.OnMessage(1, 101, "y") // fresh draft -> buttons again
	if r.buttons != 2 {
		t.Error("t:cancel did not clear the draft")
	}

	in.OnMessage(1, 102, "/cancel")
	if len(*committed) != 0 {
		t.Error("/cancel must not commit")
	}
	in.OnMessage(1, 103, "z")
	if r.buttons != 3 {
		t.Error("/cancel did not clear the draft")
	}
}

func TestStaleCallback(t *testing.T) {
	in, r, committed := newTestInbound()
	in.OnCallback(1, "cb", "t:1h") // no active draft
	if len(*committed) != 0 {
		t.Error("stale callback must not commit")
	}
	if len(r.callbacks) != 1 {
		t.Errorf("expected one (expired) callback answer, got %d", len(r.callbacks))
	}
}

func TestCustomTimeCommit(t *testing.T) {
	in, r, committed := newTestInbound()
	in.OnMessage(1, 100, "buy milk")
	in.OnCallback(1, "cb", "t:pick")
	if r.forceReplies != 1 {
		t.Fatalf("forceReplies = %d, want 1 after t:pick", r.forceReplies)
	}
	in.OnMessage(1, 101, "18:30")
	if len(*committed) != 1 {
		t.Fatalf("commits = %d, want 1", len(*committed))
	}
	want := time.Date(2026, 6, 10, 16, 30, 0, 0, time.UTC) // 18:30 Kigali (UTC+2)
	if got := (*committed)[0].ReminderAt; got == nil || !got.Equal(want) {
		t.Errorf("ReminderAt = %v, want %v", got, want)
	}
}

func TestCustomTimeAmPm(t *testing.T) {
	in, _, committed := newTestInbound() // frozen now = 12:00 UTC (14:00 Kigali)
	in.OnMessage(1, 100, "buy milk")
	in.OnCallback(1, "cb", "t:pick")
	in.OnMessage(1, 101, "10:11pm")
	if len(*committed) != 1 {
		t.Fatalf("commits = %d, want 1", len(*committed))
	}
	want := time.Date(2026, 6, 10, 20, 11, 0, 0, time.UTC) // 22:11 Kigali
	if got := (*committed)[0].ReminderAt; got == nil || !got.Equal(want) {
		t.Errorf("ReminderAt = %v, want %v", got, want)
	}
}

func TestCustomTimePastReprompts(t *testing.T) {
	in, r, committed := newTestInbound() // now = 14:00 Kigali
	in.OnMessage(1, 100, "buy milk")
	in.OnCallback(1, "cb", "t:pick")
	in.OnMessage(1, 101, "10:00") // 10:00 Kigali, already past 14:00
	if len(*committed) != 0 {
		t.Error("a past custom time must not commit (it would fire immediately)")
	}
	if r.forceReplies != 2 { // pick + past re-prompt
		t.Errorf("forceReplies = %d, want 2", r.forceReplies)
	}
}

func TestCustomTimeGarbageReprompts(t *testing.T) {
	in, r, committed := newTestInbound()
	in.OnMessage(1, 100, "buy milk")
	in.OnCallback(1, "cb", "t:pick")
	in.OnMessage(1, 101, "not a time")
	if len(*committed) != 0 {
		t.Error("garbage time must not commit")
	}
	if r.forceReplies != 2 {
		t.Errorf("forceReplies = %d, want 2 (pick + re-prompt)", r.forceReplies)
	}
}

func TestOtherCommandIgnored(t *testing.T) {
	in, r, committed := newTestInbound()
	in.OnMessage(1, 100, "/start")
	if r.buttons != 0 || len(*committed) != 0 || len(r.texts) != 0 {
		t.Error("/start must be ignored entirely")
	}
}
