package config

import (
	"log/slog"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort      string `env:"PORT" envDefault:"8080"`
	PG_DBHost       string `env:"DB_PG_HOST" envDefault:"postgres"`
	PG_DBUser       string `env:"DB_PG_USER" envDefault:"webuser"`
	PG_DBPassword   string `env:"DB_PG_PASSWORD" envDefault:"1111"`
	PG_DBName       string `env:"DB_PG_NAME" envDefault:"webdev"`
	PG_DBSSLMode    string `env:"DB_PG_SSLMODE" envDefault:"disable"`
	PG_PORT         string `env:"DB_PG_PORT" envDefault:"5432"`
	MG_DBUser       string `env:"MG_DB_USER" envDefault:"appuser"`
	MG_DBPassword   string `env:"MG_DB_PASSWORD" envDefault:"apppassword"`
	MG_DBHost       string `env:"MG_DB_HOST" envDefault:"mongodb"`
	MG_DBName       string `env:"MG_DB_NAME" envDefault:"cryptodb"`
	MG_Auth         string `env:"MG_DB_AUTH_SOURCE" envDefault:"appuser"`
	MG_Port         string `env:"MG_DB_PORT" envDefault:"27017"`
	KeyUsersGorilla string `env:"KEY_USERS_GORILLA" envDefault:"my-super-secret-key-12345"`
	LaunchLoc       string `env:"LAUNCH_LOC" envDefault:"prod"`
	LogLevel        string `env:"LOG_LEVEL" envDefault:"info"`
	TgBotToken      string `env:"TG_BOT_TOKEN" envDefault:""`
	TgChatIDs       string `env:"TG_CHAT_IDS" envDefault:""`
	RedisHost       string `env:"REDIS_HOST" envDefault:"redis"`
	RedisPassword   string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB         int    `env:"REDIS_DB" envDefault:"0"`
	RedisPoolSize   int    `env:"REDIS_POOL_SIZE" envDefault:"10"`
	RedisPort       string `env:"REDIS_PORT" envDefault:"6379"`
	ProfFlag        int    `env:"PROF_FLAG" envDefault:"0"`
}

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

func MustLoad() *Config {

	logger := NewEarlyLogger()

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
