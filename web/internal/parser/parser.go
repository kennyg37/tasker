// Package parser turns TODO.md text into tasks ready to upsert. It is the PC
// capture layer: it assigns a stable source_key per task and converts the local
// date header into a UTC reminder_at.
package parser

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/kennyg37/tasker/web/internal/task"
)

// defaultReminderHour is the local time-of-day a task reminds at, since the
// date header carries no clock time. Change here to retune.
const defaultReminderHour = 9

// kigali is the capture timezone (UTC+3, no DST). Local dates are interpreted
// here, then stored in UTC.
var kigali = time.FixedZone("Africa/Kigali", 3*60*60)

// Parse reads TODO.md content and returns the tasks it contains. Lines before
// the first date header are ignored, matching the source format.
func Parse(content string) []task.Task {
	var tasks []task.Task
	var date time.Time
	haveDate := false

	sc := bufio.NewScanner(strings.NewReader(content))
	for sc.Scan() {
		line := strings.TrimSpace(strings.TrimRight(sc.Text(), "\r"))
		if line == "" {
			continue
		}
		if d, ok := parseDate(line); ok {
			date, haveDate = d, true
			continue
		}
		if !haveDate {
			continue
		}
		if tk, ok := parseTask(line, date); ok {
			tasks = append(tasks, tk)
		}
	}
	return tasks
}

// parseDate recognizes a "D/M/YYYY" header (tolerating leading zeros and
// spaces) and returns it as midnight in Kigali.
func parseDate(s string) (time.Time, bool) {
	parts := strings.Split(s, "/")
	if len(parts) != 3 {
		return time.Time{}, false
	}
	day, e1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	month, e2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	year, e3 := strconv.Atoi(strings.TrimSpace(parts[2]))
	if e1 != nil || e2 != nil || e3 != nil {
		return time.Time{}, false
	}
	if month < 1 || month > 12 || day < 1 || day > 31 || year < 1 {
		return time.Time{}, false
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, kigali), true
}

// parseTask parses a "N. Description - Status" line. The trailing "- Status" is
// stripped only when Status is a recognized keyword, so dashes inside a
// description are preserved.
func parseTask(line string, date time.Time) (task.Task, bool) {
	dot := strings.IndexByte(line, '.')
	if dot < 0 {
		return task.Task{}, false
	}
	if _, err := strconv.Atoi(strings.TrimSpace(line[:dot])); err != nil {
		return task.Task{}, false
	}

	rest := line[dot+1:]
	desc := strings.TrimSpace(rest)
	isDone := false
	if dash := strings.LastIndexByte(rest, '-'); dash >= 0 {
		if done, ok := statusFromTail(rest[dash+1:]); ok {
			desc = strings.TrimSpace(rest[:dash])
			isDone = done
		}
	}
	if desc == "" {
		return task.Task{}, false
	}

	remindAt := time.Date(date.Year(), date.Month(), date.Day(), defaultReminderHour, 0, 0, 0, kigali).UTC()
	return task.Task{
		SourceKey:  sourceKey(date, desc),
		Content:    desc,
		Origin:     task.OriginPC,
		ReminderAt: &remindAt,
		IsDone:     isDone,
	}, true
}

// statusFromTail reports whether the trailing segment is a recognized status
// keyword, and if so whether it means done.
func statusFromTail(tail string) (done bool, recognized bool) {
	switch strings.ToLower(strings.Trim(strings.TrimSpace(tail), ".!✓")) {
	case "done":
		return true, true
	case "postponed", "pending":
		return false, true
	default:
		return false, false
	}
}

// sourceKey hashes the canonical date and description so the same task keeps a
// stable key across reordering, status changes, and date-format variants.
func sourceKey(date time.Time, desc string) string {
	sum := sha256.Sum256([]byte(date.Format("2006-01-02") + "\n" + desc))
	return "pc:" + hex.EncodeToString(sum[:])
}
