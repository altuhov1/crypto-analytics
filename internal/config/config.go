package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config — структура для хранения всех конфигов приложения
type Config struct {
	ServerPort  string // Порт, на котором запускается сервер
	StoragePath string // Путь к файлу с данными
	// Сюда потом можно добавить настройки базы данных, API-ключи и т.д.
}

// MustLoad — функция, которая загружает конфиг из переменных окружения.
// Если что-то не так — она вызовет panic (Must — распространенное в Go
// название для функций, которые не возвращают error, а паникуют).
func MustLoad() *Config {
	// Загружаем переменные из файла .env
	if err := godotenv.Load(); err != nil {
		// Если ошибка "файл не найден" - это ок, работаем дальше.
		// Любая другая ошибка (например, файл есть, но он битый) - это плохо.
		if !os.IsNotExist(err) {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}

	port := getEnv("PORT", "8080")
	storagePath := getEnv("STORAGE_PATH", "bd/jsonsWithData.json")

	return &Config{
		ServerPort:  port,
		StoragePath: storagePath,
	}
}

// getEnv — вспомогательная функция для чтения переменных окружения.
// Если переменной нет, возвращает значение по умолчанию.
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Printf("Environment variable %s not found, using default: %s", key, defaultValue)
		return defaultValue
	}
	// Если переменная есть, но она пустая, тоже используем default.
	if value == "" {
		log.Printf("Environment variable %s is empty, using default: %s", key, defaultValue)
		return defaultValue
	}
	return value
}
