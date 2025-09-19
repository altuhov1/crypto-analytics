// storage/pgx_storage.go
package storage

import (
	"context"
	"fmt"
	"time"

	"webdev-90-days/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PGXStorage struct {
	pool *pgxpool.Pool
}

type PGXConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewPGXStorage(cfg PGXConfig) (*PGXStorage, error) {
	// Формируем строку подключения
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	// Создаем пул соединений
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка конфигурации: %w", err)
	}

	// Настройки пула
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour

	// Подключаемся
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения: %w", err)
	}

	// Проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("база не отвечает: %w", err)
	}

	return &PGXStorage{pool: pool}, nil
}

func (s *PGXStorage) SaveContact(contact *models.ContactForm) error {
	query := `
	INSERT INTO contacts (name, email, message) 
	VALUES ($1, $2, $3)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Выполняем INSERT через Go!
	_, err := s.pool.Exec(ctx, query, contact.Name, contact.Email, contact.Message)
	if err != nil {
		return fmt.Errorf("ошибка сохранения: %w", err)
	}

	return nil
}

func (s *PGXStorage) Close() error {
	s.pool.Close()
	return nil
}
