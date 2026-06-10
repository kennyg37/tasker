// Package telegram is the bot transport: outbound sends and inbound update
// dispatch. The conversation logic lives in Inbound (inbound.go); this file is
// the thin client wrapping the API.
package telegram

import (
	"context"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/kennyg37/tasker/web/internal/task"
)

type Client struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

// New validates the token with Telegram (a getMe call) and parses the chat id.
func New(token, chatID string) (*Client, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("telegram: connect: %w", err)
	}
	id, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("telegram: invalid chat id %q: %w", chatID, err)
	}
	return &Client{bot: bot, chatID: id}, nil
}

// SendText sends a plain message to the given chat.
func (c *Client) SendText(chatID int64, text string) error {
	if _, err := c.bot.Send(tgbotapi.NewMessage(chatID, text)); err != nil {
		return fmt.Errorf("telegram: send: %w", err)
	}
	return nil
}

// SendButtons sends a message carrying an inline keyboard.
func (c *Client) SendButtons(chatID int64, text string, kb tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = kb
	if _, err := c.bot.Send(msg); err != nil {
		return fmt.Errorf("telegram: send buttons: %w", err)
	}
	return nil
}

// AnswerCallback acknowledges a button tap so the client's spinner clears.
func (c *Client) AnswerCallback(callbackID, text string) error {
	if _, err := c.bot.Request(tgbotapi.NewCallback(callbackID, text)); err != nil {
		return fmt.Errorf("telegram: answer callback: %w", err)
	}
	return nil
}

// SendForceReply sends a prompt that pre-focuses the user's reply box.
func (c *Client) SendForceReply(chatID int64, prompt string) error {
	msg := tgbotapi.NewMessage(chatID, prompt)
	msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
	if _, err := c.bot.Send(msg); err != nil {
		return fmt.Errorf("telegram: send force reply: %w", err)
	}
	return nil
}

// Send delivers a reminder. Its signature matches scheduler.Sender.
func (c *Client) Send(t task.Task) error {
	return c.SendText(c.chatID, "Reminder: "+t.Content)
}

// Listen streams updates from the configured chat (ignoring all others — this
// is a single-user bot) and dispatches messages and button taps to in. It
// blocks until ctx is cancelled.
func (c *Client) Listen(ctx context.Context, in *Inbound) {
	cfg := tgbotapi.NewUpdate(0)
	cfg.Timeout = 30
	updates := c.bot.GetUpdatesChan(cfg)
	for {
		select {
		case <-ctx.Done():
			c.bot.StopReceivingUpdates()
			return
		case u := <-updates:
			switch {
			case u.CallbackQuery != nil:
				cq := u.CallbackQuery
				if cq.Message == nil || cq.Message.Chat.ID != c.chatID {
					continue
				}
				in.OnCallback(cq.Message.Chat.ID, cq.ID, cq.Data)
			case u.Message != nil:
				if u.Message.Chat.ID != c.chatID {
					continue
				}
				in.OnMessage(u.Message.Chat.ID, u.Message.MessageID, u.Message.Text)
			}
		}
	}
}
