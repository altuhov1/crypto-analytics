package config

import (
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

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
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Debug("Failed to load .env file", "error", err)

	} else {
		slog.Info("Loaded configuration from .env file")
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		slog.Error("Failed to parse environment variables", "error", err)
		panic(err)
	}

	return &cfg
}
