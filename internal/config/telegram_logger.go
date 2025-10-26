package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

type telegramHandler struct {
	botToken string
	chatIDs  []string
	level    slog.Level
	client   *http.Client
}

func (h *telegramHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *telegramHandler) Handle(_ context.Context, r slog.Record) error {
	if len(h.chatIDs) == 0 {
		return nil
	}

	// Формируем базовое сообщение
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("[%s] %s\n", r.Level, r.Time.Format("15:04:05")))
	msg.WriteString(fmt.Sprintf("Message: %s\n", r.Message))

	// Обрабатываем атрибуты
	if r.NumAttrs() > 0 {
		msg.WriteString("\nAttributes:\n")
		r.Attrs(func(attr slog.Attr) bool {
			msg.WriteString(fmt.Sprintf("  %s: %v\n", attr.Key, attr.Value.Any()))
			return true // продолжаем обработку
		})
	}

	// Отправляем асинхронно
	go func(message string) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		for _, chatID := range h.chatIDs {
			if err := h.sendToTelegram(ctx, chatID, message); err != nil {
				fmt.Fprintf(os.Stderr, "[TelegramLogger] Failed to send to %s: %v\n", chatID, err)
			}
		}
	}(msg.String())

	return nil
}
func (h *telegramHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *telegramHandler) WithGroup(name string) slog.Handler {
	return h
}

func parseChatIDs(chatIDsStr string) []string {
	if chatIDsStr == "" {
		return nil
	}
	var ids []string
	for _, id := range strings.Split(chatIDsStr, ",") {
		id = strings.TrimSpace(id)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func newTelegramLogger(cfg *Config) *slog.Logger {
	level := getLogLevelFromString(cfg.LogLevel)
	handler := &telegramHandler{
		botToken: cfg.TgBotToken,
		chatIDs:  parseChatIDs(cfg.TgChatIDs),
		level:    level,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
	return slog.New(handler)
}

func newTelegramLoggerFromEnv(token, chatIDsStr, logLevelStr string) *slog.Logger {
	level := getLogLevelFromString(logLevelStr)
	handler := &telegramHandler{
		botToken: token,
		chatIDs:  parseChatIDs(chatIDsStr),
		level:    level,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
	return slog.New(handler)
}

func (h *telegramHandler) sendToTelegram(ctx context.Context, chatID, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", h.botToken)
	payload := map[string]string{
		"chat_id": chatID,
		"text":    text,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа для диагностики
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Проверяем JSON ответ
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	if !result["ok"].(bool) {
		return fmt.Errorf("telegram API error: %v", result)
	}

	return nil
}
