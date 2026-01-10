package config

import (
	"log/slog"
	"os"
)

// NewEarlyLogger создаёт логгер ДО загрузки полного конфига
func NewEarlyLogger() *slog.Logger {
	launchLoc := os.Getenv("LAUNCH_LOC")
	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = "info"
	}

	if launchLoc == "prod" {
		tgToken := os.Getenv("TG_BOT_TOKEN")
		tgChatID := os.Getenv("TG_CHAT_ID")
		if tgToken != "" && tgChatID != "" {
			// TODO: Telegram-логгер
			return newTelegramLoggerFromEnv(tgToken, tgChatID, logLevelStr)
		}
	}

	// fallback на stdout
	level := getLogLevelFromString(logLevelStr)
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
