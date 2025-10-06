package config

import (
	"log"
	"os"

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
	KeyUsersGorilla string `env:"KEY_USER_GORILLA" envDefault:"my-super-secret-key-12345"`
}

func MustLoad() *Config {
	// Загружаем .env файл если он существует
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("Failed to parse environment variables: %v", err)
	}

	return cfg
}
