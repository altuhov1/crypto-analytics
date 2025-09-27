package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	// Добавляем настройки DB
	DBHost          string
	DBPort          int
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	KeyUsersGorilla string
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	port := getEnv("PORT", "8080")

	// Получаем настройки DB из environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	dbUser := getEnv("DB_USER", "webuser")
	dbPassword := getEnv("DB_PASSWORD", "1111")
	dbName := getEnv("DB_NAME", "webdev")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")
	KEY_USER_GORILLA := getEnv("DB_SSLMODE", "my-super-secret-key-12345")

	return &Config{
		ServerPort:      port,
		DBHost:          dbHost,
		DBPort:          dbPort,
		DBUser:          dbUser,
		DBPassword:      dbPassword,
		DBName:          dbName,
		DBSSLMode:       dbSSLMode,
		KeyUsersGorilla: KEY_USER_GORILLA,
	}
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return defaultValue
	}
	return value
}
