package storage

import (
	"context"
	"crypto-analytics/internal/config"
	"crypto-analytics/internal/storage"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewPoolPg(cfg *storage.PGXConfig) (*pgxpool.Pool, error) {
	// Формируем строку подключения
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.PG_SSLMode)
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

	return pool, nil
}

func NewMongoClient(cfg *config.Config) (*mongo.Client, error) {
	// connStr := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
	// 	cfg.MG_DBUser, cfg.MG_DBPassword, cfg.MG_DBHost, cfg.MG_DBPort, cfg.MG_DBName)

	// clientOptions := options.Client().ApplyURI(connStr)
	// clientOptions.SetMaxPoolSize(100)
	// clientOptions.SetMinPoolSize(5)
	// clientOptions.SetMaxConnIdleTime(30 * time.Minute)

	// // Подключаемся напрямую через Connect
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	// client, err := mongo.Connect(ctx, clientOptions)
	// if err != nil {
	// 	return nil, fmt.Errorf("ошибка подключения к MongoDB: %w", err)
	// }

	// // Проверяем соединение
	// pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer pingCancel()

	// if err := client.Ping(pingCtx, nil); err != nil {
	// 	_ = client.Disconnect(context.Background())
	// 	return nil, fmt.Errorf("MongoDB не отвечает: %w", err)
	// }

	// return client, nil
}
