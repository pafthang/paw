package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pafthang/paw/internal/agent"
	"github.com/pafthang/paw/internal/channels"
	"github.com/pafthang/paw/internal/config"
	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/llm"
	"github.com/pafthang/paw/internal/tools"
)

type Channel struct {
	settings config.Settings

	mu     sync.Mutex
	bot    *tgbotapi.BotAPI
	cancel context.CancelFunc
	status agentStatus
}

type agentStatus struct {
	running   bool
	startedAt time.Time
	lastError string
}

func New(settings config.Settings) *Channel {
	return &Channel{settings: settings}
}

func (c *Channel) Name() string { return "telegram" }

func (c *Channel) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel != nil {
		return errors.New("already running")
	}
	if c.settings.TelegramBotToken == "" {
		return errors.New("missing telegram_bot_token")
	}
	bot, err := tgbotapi.NewBotAPI(c.settings.TelegramBotToken)
	if err != nil {
		c.status.lastError = err.Error()
		return err
	}
	c.bot = bot
	runCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.status.running = true
	c.status.startedAt = time.Now().UTC()
	c.status.lastError = ""

	go c.loop(runCtx, bot)
	return nil
}

func (c *Channel) Stop(ctx context.Context) error {
	c.mu.Lock()
	cancel := c.cancel
	c.cancel = nil
	c.status.running = false
	c.mu.Unlock()
	_ = ctx
	if cancel != nil {
		cancel()
	}
	return nil
}

func (c *Channel) Status() channels.ChannelStatus {
	c.mu.Lock()
	defer c.mu.Unlock()
	return channels.ChannelStatus{
		Name:      c.Name(),
		Running:   c.status.running,
		LastError: c.status.lastError,
		StartedAt: c.status.startedAt,
	}
}

func (c *Channel) loop(ctx context.Context, bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := bot.GetUpdatesChan(u)
	for {
		select {
		case <-ctx.Done():
			return
		case upd, ok := <-updates:
			if !ok {
				return
			}
			c.handleUpdate(ctx, bot, upd)
		}
	}
}

func (c *Channel) handleUpdate(ctx context.Context, bot *tgbotapi.BotAPI, upd tgbotapi.Update) {
	if upd.Message == nil || upd.Message.Text == "" {
		return
	}
	userID := int64(0)
	if upd.Message.From != nil {
		userID = upd.Message.From.ID
	}
	if c.settings.AllowedUserID != 0 && userID != c.settings.AllowedUserID {
		c.auditEvent(0, "channel.message.received", map[string]any{
			"channel": "telegram",
			"chat_id": upd.Message.Chat.ID,
			"user_id": userID,
			"text":    snippet(upd.Message.Text),
		}, nil, errors.New("unauthorized user"))
		_, _ = bot.Send(tgbotapi.NewMessage(upd.Message.Chat.ID, "Unauthorized."))
		return
	}

	prompt := upd.Message.Text
	database, err := db.Open()
	if err != nil {
		c.setErr(err)
		c.auditEvent(0, "channel.error", map[string]any{"channel": "telegram"}, nil, err)
		_, _ = bot.Send(tgbotapi.NewMessage(upd.Message.Chat.ID, "Error: "+err.Error()))
		return
	}
	c.auditEvent(0, "channel.message.received", map[string]any{
		"channel": "telegram",
		"chat_id": upd.Message.Chat.ID,
		"user_id": userID,
		"text":    snippet(prompt),
	}, nil, nil)
	client, err := llm.NewClient(c.settings)
	if err != nil {
		c.setErr(err)
		c.auditEvent(0, "channel.error", map[string]any{"channel": "telegram"}, nil, err)
		_, _ = bot.Send(tgbotapi.NewMessage(upd.Message.Chat.ID, "Error: "+err.Error()))
		return
	}
	runner := agent.NewRunner(database, tools.DefaultRegistry())
	resp, err := runner.Chat(ctx, client, agent.ChatRequest{
		Prompt:          prompt,
		Model:           llm.DefaultModel(c.settings),
		MaxIterations:   4,
		MaxContextChars: 12000,
	})
	if err != nil {
		c.setErr(err)
		c.auditEvent(resp.SessionID, "channel.error", map[string]any{
			"channel":    "telegram",
			"session_id": resp.SessionID,
		}, nil, err)
		_, _ = bot.Send(tgbotapi.NewMessage(upd.Message.Chat.ID, "Error: "+err.Error()))
		return
	}
	text := resp.FinalResponse.Content
	if !resp.UsedTools {
		text = resp.ModelResponse.Content
	}
	if text == "" {
		text = "(empty response)"
	}
	if _, err := bot.Send(tgbotapi.NewMessage(upd.Message.Chat.ID, text)); err != nil {
		slog.Warn("telegram send failed", "error", err)
		c.setErr(err)
		c.auditEvent(resp.SessionID, "channel.error", map[string]any{
			"channel":    "telegram",
			"session_id": resp.SessionID,
			"chat_id":    upd.Message.Chat.ID,
			"user_id":    userID,
		}, nil, err)
		return
	}
	c.auditEvent(resp.SessionID, "channel.message.sent", map[string]any{
		"channel":    "telegram",
		"session_id": resp.SessionID,
		"chat_id":    upd.Message.Chat.ID,
		"user_id":    userID,
		"text":       snippet(text),
	}, map[string]any{"ok": true}, nil)
}

func (c *Channel) setErr(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil {
		c.status.lastError = err.Error()
	} else {
		c.status.lastError = ""
	}
}

func (c *Channel) auditEvent(sessionID uint, eventType string, input any, output any, err error) {
	database, dbErr := db.Open()
	if dbErr != nil {
		return
	}
	inputJSON, _ := json.Marshal(input)
	outputJSON, _ := json.Marshal(output)
	ev := db.AuditEvent{
		SessionID:  sessionID,
		Type:       eventType,
		ToolName:   "telegram",
		InputJSON:  string(inputJSON),
		OutputJSON: string(outputJSON),
	}
	if err != nil {
		ev.Error = err.Error()
	}
	_, _ = db.CreateAuditEvent(database, ev)
}

func snippet(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 200 {
		return s
	}
	return s[:200]
}
