package config

import (
	"log/slog"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// getLogLevelFromString преобразует строку в slog.Level
// Используется и в MustLoad (до парсинга cfg), и в GetLogLevel (после)
func getLogLevelFromString(levelStr string) slog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type Config struct {
	ServerPort      string `env:"PORT" envDefault:"8080"`
	DBHost          string `env:"DB_HOST" envDefault:"localhost"`
	DBPort          int    `env:"DB_PORT" envDefault:"5432"`
	DBUser          string `env:"DB_USER" envDefault:"webuser"`
	DBPassword      string `env:"DB_PASSWORD" envDefault:"1111"`
	DBName          string `env:"DB_NAME" envDefault:"webdev"`
	DBSSLMode       string `env:"DB_SSLMODE" envDefault:"disable"`
	KeyUsersGorilla string `env:"KEY_USERS_GORILLA" envDefault:"my-super-secret-key-12345"`
	LaunchLoc       string `env:"LAUNCH_LOC" envDefault:"local"`
	LogLevel        string `env:"LOG_LEVEL" envDefault:"info"`
	TgBotToken      string `env:"TG_BOT_TOKEN" envDefault:""`
	TgChatIDs       string `env:"TG_CHAT_IDS" envDefault:""`
}

func MustLoad() *Config {

	logger := NewEarlyLogger()

	// Загружаем .env (если есть)
	if err := godotenv.Load(); err != nil {

		logger.Debug("Failed to load .env file", "error", err)
	} else {
		logger.Info("Loaded configuration from .env file")
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		logger.Error("Failed to parse environment variables", "error", err)
		panic("configuration error: " + err.Error())
	}
	logger.Info("Application started", "mode", cfg.LaunchLoc)

	return &cfg
}

func (c *Config) GetLogLevel() slog.Level {
	return getLogLevelFromString(c.LogLevel)
}
