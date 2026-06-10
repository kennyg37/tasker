package telegram

import (
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/kennyg37/tasker/web/internal/draft"
	"github.com/kennyg37/tasker/web/internal/task"
)

// Replier is the outbound surface the inbound conversation needs. The real
// Client satisfies it; tests fake it.
type Replier interface {
	SendButtons(chatID int64, text string, kb tgbotapi.InlineKeyboardMarkup) error
	AnswerCallback(callbackID, text string) error
	SendForceReply(chatID int64, prompt string) error
	SendText(chatID int64, text string) error
}

// Inbound runs the phone-capture conversation: stage a draft, collect a
// reminder time via buttons, commit one complete task. It holds no firing
// logic; commit is the existing store.Upsert, injected.
type Inbound struct {
	r      Replier
	drafts *draft.Store
	commit func(*task.Task) error
	now    func() time.Time
}

func NewInbound(r Replier, drafts *draft.Store, commit func(*task.Task) error, now func() time.Time) *Inbound {
	return &Inbound{r: r, drafts: drafts, commit: commit, now: now}
}

// OnMessage handles an inbound text message from the authorized chat.
func (in *Inbound) OnMessage(chatID int64, messageID int, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	if text == "/cancel" {
		in.drafts.Clear(chatID)
		logErr(in.r.SendText(chatID, "Cancelled."))
		return
	}
	if strings.HasPrefix(text, "/") {
		return // other bot command — never a task or a field
	}

	d, ok := in.drafts.Get(chatID)
	if !ok {
		in.drafts.Start(chatID, fmt.Sprintf("phone:%d", messageID), text)
		logErr(in.r.SendButtons(chatID, "Remind you about "+quote(text)+" — when?", presetKeyboard()))
		return
	}

	switch d.Step {
	case draft.StepAwaitingTime:
		logErr(in.r.SendText(chatID, "Finish setting a time for "+quote(d.Content)+" first, or tap Cancel."))
	case draft.StepAwaitingCustomTime:
		at, err := draft.ParseCustomTime(text, in.now())
		if err != nil {
			logErr(in.r.SendForceReply(chatID, "Couldn't read that. Send a time like 22:11 (24h) or 10:11pm, or 2026-06-12 09:00."))
			return
		}
		if !at.After(in.now()) {
			logErr(in.r.SendForceReply(chatID, "That time has already passed. Use 24-hour (e.g. 22:11) or add am/pm (10:11pm)."))
			return
		}
		in.commitTask(chatID, d, &at)
	}
}

// OnCallback handles an inline button tap from the authorized chat.
func (in *Inbound) OnCallback(chatID int64, callbackID, data string) {
	d, ok := in.drafts.Get(chatID)
	if !ok {
		logErr(in.r.AnswerCallback(callbackID, "This form expired — send the task again."))
		return
	}
	switch data {
	case "t:cancel":
		in.drafts.Clear(chatID)
		logErr(in.r.AnswerCallback(callbackID, "Cancelled."))
	case "t:pick":
		in.drafts.Advance(chatID)
		logErr(in.r.AnswerCallback(callbackID, ""))
		logErr(in.r.SendForceReply(chatID, "Send a time like 22:11 (24h) or 10:11pm, or 2026-06-12 09:00."))
	case "t:none":
		logErr(in.r.AnswerCallback(callbackID, ""))
		in.commitTask(chatID, d, nil)
	default:
		at, ok := draft.ResolvePreset(data, in.now())
		if !ok {
			logErr(in.r.AnswerCallback(callbackID, "Unknown option."))
			return
		}
		logErr(in.r.AnswerCallback(callbackID, ""))
		in.commitTask(chatID, d, &at)
	}
}

func (in *Inbound) commitTask(chatID int64, d *draft.Draft, at *time.Time) {
	t := task.Task{
		SourceKey:  d.SourceKey,
		Content:    d.Content,
		Origin:     task.OriginPhone,
		ReminderAt: at,
	}
	if err := in.commit(&t); err != nil {
		log.Printf("telegram inbound: commit %s: %v", d.SourceKey, err)
		logErr(in.r.SendText(chatID, "Sorry, couldn't save that — try again."))
		return
	}
	in.drafts.Clear(chatID)
	if at != nil {
		logErr(in.r.SendText(chatID, "Reminder set for "+draft.Human(*at)))
	} else {
		logErr(in.r.SendText(chatID, "Captured (no reminder)"))
	}
}

func presetKeyboard() tgbotapi.InlineKeyboardMarkup {
	b := tgbotapi.NewInlineKeyboardButtonData
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(b("In 1 hour", "t:1h"), b("In 3 hours", "t:3h")),
		tgbotapi.NewInlineKeyboardRow(b("This evening", "t:eve"), b("Tomorrow 9am", "t:tom")),
		tgbotapi.NewInlineKeyboardRow(b("Pick time", "t:pick"), b("No reminder", "t:none")),
		tgbotapi.NewInlineKeyboardRow(b("Cancel", "t:cancel")),
	)
}

func quote(s string) string { return "'" + s + "'" }

func logErr(err error) {
	if err != nil {
		log.Printf("telegram inbound: %v", err)
	}
}
