package config

import (
	"log/slog"
	"os"
)

// (local / prod)
func NewLogger(cfg *Config) *slog.Logger {
	if cfg.LaunchLoc == "prod" && cfg.TgBotToken != "" && cfg.TgChatIDs != "" {
		return newTelegramLogger(cfg)
	}

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: getLogLevelFromString(cfg.LogLevel),
	}))
}
