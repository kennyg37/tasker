// Package draft stages an in-progress phone task until its reminder time is
// chosen, so the tasks table never holds a partial row. It is pure logic: no
// Telegram, no DB, no wall clock (now is injected). All times resolve in the
// fixed Kigali zone and are returned in UTC.
package draft

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// kigali is UTC+2 (CAT) with no DST. A fixed zone is used deliberately: the
// distroless runtime image has no tzdata, so time.LoadLocation would fail.
var kigali = time.FixedZone("Kigali", 2*3600)

type Step int

const (
	StepAwaitingTime       Step = iota // presets shown, waiting for a button tap
	StepAwaitingCustomTime             // "Pick time" tapped, waiting for an HH:MM reply
)

type Draft struct {
	ChatID     int64
	SourceKey  string // phone:<message_id> of the content message; the task key
	Content    string
	ReminderAt *time.Time
	Step       Step
	UpdatedAt  time.Time
}

type Store struct {
	mu     sync.Mutex
	drafts map[int64]*Draft
	now    func() time.Time
	ttl    time.Duration
}

func New(now func() time.Time, ttl time.Duration) *Store {
	return &Store{drafts: make(map[int64]*Draft), now: now, ttl: ttl}
}

// Start stages a new draft, replacing any existing one for the chat.
func (s *Store) Start(chatID int64, sourceKey, content string) *Draft {
	s.mu.Lock()
	defer s.mu.Unlock()
	d := &Draft{
		ChatID:    chatID,
		SourceKey: sourceKey,
		Content:   content,
		Step:      StepAwaitingTime,
		UpdatedAt: s.now(),
	}
	s.drafts[chatID] = d
	return d
}

// Get returns the active draft, or false if there is none or it has gone stale.
// A stale draft is cleared on access, so an abandoned form self-heals the
// one-draft gate.
func (s *Store) Get(chatID int64) (*Draft, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.drafts[chatID]
	if !ok {
		return nil, false
	}
	if s.now().Sub(d.UpdatedAt) > s.ttl {
		delete(s.drafts, chatID)
		return nil, false
	}
	return d, true
}

// Advance moves an active draft to StepAwaitingCustomTime and refreshes its
// freshness. Returns false if there is no active draft.
func (s *Store) Advance(chatID int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.drafts[chatID]
	if !ok {
		return false
	}
	d.Step = StepAwaitingCustomTime
	d.UpdatedAt = s.now()
	return true
}

func (s *Store) Clear(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.drafts, chatID)
}

// ResolvePreset returns the reminder time for a time-producing preset token,
// computed from now in Kigali and returned in UTC. ok is false for tokens that
// are not times (t:pick, t:none, t:cancel).
func ResolvePreset(token string, now time.Time) (time.Time, bool) {
	k := now.In(kigali)
	switch token {
	case "t:1h":
		return now.Add(time.Hour).UTC(), true
	case "t:3h":
		return now.Add(3 * time.Hour).UTC(), true
	case "t:eve":
		return time.Date(k.Year(), k.Month(), k.Day(), 18, 0, 0, 0, kigali).UTC(), true
	case "t:tom":
		n := k.AddDate(0, 0, 1)
		return time.Date(n.Year(), n.Month(), n.Day(), 9, 0, 0, 0, kigali).UTC(), true
	default:
		return time.Time{}, false
	}
}

// ParseCustomTime parses a time in Kigali and returns UTC. It accepts 24-hour
// ("22:11") and 12-hour ("10:11pm", "10:11 PM") clock times — resolved to today
// — plus a full "YYYY-MM-DD HH:MM" (24h or am/pm). Rigid format parsing, not
// natural language. The caller decides what to do if the result is in the past.
func ParseCustomTime(s string, now time.Time) (time.Time, error) {
	up := strings.ToUpper(strings.TrimSpace(s))

	// Absolute date+time forms.
	for _, layout := range []string{"2006-01-02 15:04", "2006-01-02 3:04PM", "2006-01-02 3:04 PM"} {
		if t, err := time.ParseInLocation(layout, up, kigali); err == nil {
			return t.UTC(), nil
		}
	}
	// Time-of-day forms → today in Kigali.
	for _, layout := range []string{"15:04", "3:04PM", "3:04 PM"} {
		if t, err := time.ParseInLocation(layout, up, kigali); err == nil {
			k := now.In(kigali)
			return time.Date(k.Year(), k.Month(), k.Day(), t.Hour(), t.Minute(), 0, 0, kigali).UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time %q", s)
}

// Human formats a UTC instant as Kigali local time for confirmations.
func Human(t time.Time) string {
	return t.In(kigali).Format("Mon 2 Jan 15:04")
}
