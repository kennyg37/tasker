// Package telegram delivers reminders to a single chat. It is the outbound side
// of the bot — the concrete Sender the scheduler calls. No firing logic lives
// here; it only sends a message.
package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

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

// SendText sends a plain message to the configured chat.
func (c *Client) SendText(text string) error {
	if _, err := c.bot.Send(tgbotapi.NewMessage(c.chatID, text)); err != nil {
		return fmt.Errorf("telegram: send: %w", err)
	}
	return nil
}

// Send delivers a reminder. Its signature matches scheduler.Sender.
func (c *Client) Send(t task.Task) error {
	return c.SendText("Reminder: " + t.Content)
}

// Listen streams incoming text messages from the configured chat (ignoring all
// others — this is a single-user bot) and calls handle for each. It blocks
// until ctx is cancelled.
func (c *Client) Listen(ctx context.Context, handle func(messageID int, text string) error) {
	cfg := tgbotapi.NewUpdate(0)
	cfg.Timeout = 30
	updates := c.bot.GetUpdatesChan(cfg)
	for {
		select {
		case <-ctx.Done():
			c.bot.StopReceivingUpdates()
			return
		case u := <-updates:
			if u.Message == nil || u.Message.Chat.ID != c.chatID {
				continue
			}
			text := strings.TrimSpace(u.Message.Text)
			if text == "" || strings.HasPrefix(text, "/") {
				continue // empty or a bot command (e.g. /start), not a task
			}
			if err := handle(u.Message.MessageID, text); err != nil {
				log.Printf("telegram: handling message %d failed: %v", u.Message.MessageID, err)
			}
		}
	}
}
