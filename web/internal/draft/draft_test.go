package draft

import (
	"testing"
	"time"
)

func frozen(t time.Time) func() time.Time { return func() time.Time { return t } }

func TestStartGetClear(t *testing.T) {
	s := New(frozen(time.Now()), 10*time.Minute)

	d := s.Start(7, "phone:7", "buy milk")
	if d.Content != "buy milk" || d.SourceKey != "phone:7" || d.Step != StepAwaitingTime {
		t.Fatalf("unexpected draft: %+v", d)
	}
	got, ok := s.Get(7)
	if !ok || got.Content != "buy milk" {
		t.Fatalf("Get returned ok=%v draft=%+v", ok, got)
	}
	s.Clear(7)
	if _, ok := s.Get(7); ok {
		t.Fatal("draft should be gone after Clear")
	}
}

func TestExpiryClearsStaleOnAccess(t *testing.T) {
	cur := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	s := New(func() time.Time { return cur }, 10*time.Minute)

	s.Start(1, "phone:1", "x")
	cur = cur.Add(11 * time.Minute)
	if _, ok := s.Get(1); ok {
		t.Fatal("stale draft should report absent")
	}
	// And it should be cleared, so a fresh start works cleanly (gate self-heal).
	d := s.Start(1, "phone:2", "y")
	if d.Content != "y" {
		t.Fatalf("fresh draft = %+v", d)
	}
}

func TestAdvance(t *testing.T) {
	cur := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	s := New(func() time.Time { return cur }, 10*time.Minute)
	s.Start(1, "phone:1", "x")

	cur = cur.Add(time.Minute)
	if !s.Advance(1) {
		t.Fatal("Advance should return true for an active draft")
	}
	d, _ := s.Get(1)
	if d.Step != StepAwaitingCustomTime {
		t.Errorf("Step = %v, want StepAwaitingCustomTime", d.Step)
	}
	if !d.UpdatedAt.Equal(cur) {
		t.Errorf("UpdatedAt not bumped on Advance: %v", d.UpdatedAt)
	}
}

func TestResolvePreset(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC) // 14:00 Kigali (UTC+2)
	cases := map[string]time.Time{
		"t:1h":  time.Date(2026, 6, 10, 13, 0, 0, 0, time.UTC),
		"t:3h":  time.Date(2026, 6, 10, 15, 0, 0, 0, time.UTC),
		"t:eve": time.Date(2026, 6, 10, 16, 0, 0, 0, time.UTC), // 18:00 Kigali
		"t:tom": time.Date(2026, 6, 11, 7, 0, 0, 0, time.UTC),  // 09:00 Kigali next day
	}
	for token, want := range cases {
		got, ok := ResolvePreset(token, now)
		if !ok {
			t.Errorf("%s: ok=false, want a time", token)
			continue
		}
		if !got.Equal(want) {
			t.Errorf("%s = %v, want %v", token, got, want)
		}
	}
	if _, ok := ResolvePreset("t:pick", now); ok {
		t.Error("t:pick is not a time preset")
	}
}

func TestResolvePresetEveningCanBePast(t *testing.T) {
	now := time.Date(2026, 6, 10, 17, 0, 0, 0, time.UTC) // 19:00 Kigali, after 18:00
	got, ok := ResolvePreset("t:eve", now)
	if !ok {
		t.Fatal("t:eve should resolve")
	}
	if !got.Before(now) {
		t.Errorf("expected a past time, got %v (now %v)", got, now)
	}
}

func TestParseCustomTime(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC) // 14:00 Kigali (UTC+2)

	got, err := ParseCustomTime("18:30", now)
	if err != nil {
		t.Fatalf("HH:MM: %v", err)
	}
	if want := time.Date(2026, 6, 10, 16, 30, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("HH:MM = %v, want %v", got, want)
	}

	got, err = ParseCustomTime("2026-06-12 09:00", now)
	if err != nil {
		t.Fatalf("full: %v", err)
	}
	if want := time.Date(2026, 6, 12, 7, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("full = %v, want %v", got, want)
	}

	// 12-hour clock: "10:11pm" == 22:11 Kigali == 20:11 UTC.
	pmWant := time.Date(2026, 6, 10, 20, 11, 0, 0, time.UTC)
	for _, in := range []string{"10:11pm", "10:11 PM", "10:11PM"} {
		got, err := ParseCustomTime(in, now)
		if err != nil {
			t.Fatalf("%q: %v", in, err)
		}
		if !got.Equal(pmWant) {
			t.Errorf("%q = %v, want %v", in, got, pmWant)
		}
	}
	// "9:30am" == 09:30 Kigali == 07:30 UTC.
	if got, err := ParseCustomTime("9:30am", now); err != nil || !got.Equal(time.Date(2026, 6, 10, 7, 30, 0, 0, time.UTC)) {
		t.Errorf("9:30am = %v (err %v)", got, err)
	}

	if _, err := ParseCustomTime("not a time", now); err == nil {
		t.Error("garbage should error")
	}
}
