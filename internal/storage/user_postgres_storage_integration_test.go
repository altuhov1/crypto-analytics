// storage/user_postgres_storage_integration_test.go
package storage

import (
	"context"
	"testing"
	"time"

	"webdev-90-days/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *UserPostgresStorage {
	ctx := context.Background()

	// Запускаем Postgres в Docker контейнере
	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate postgres container: %v", err)
		}
	})

	connStr, err := postgresContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Создаем хранилище
	config := PGXConfig{
		Host:     "localhost", // testcontainers пробросит порт
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	storage, err := NewUserPostgresStorage(config)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Создаем таблицу для тестов
	_, err = storage.pool.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            email VARCHAR(255) UNIQUE NOT NULL,
            password VARCHAR(255) NOT NULL,
            username VARCHAR(255) UNIQUE NOT NULL,
            favorite_coins TEXT[] DEFAULT '{}',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	return storage
}

func TestUserPostgresStorage_CreateUser_Success(t *testing.T) {
	storage := setupTestDB(t)
	defer storage.Close()

	user := &models.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Username: "testuser",
	}

	err := storage.CreateUser(user)

	assert.NoError(t, err)
}

func TestUserPostgresStorage_CreateUser_DuplicateEmail(t *testing.T) {
	storage := setupTestDB(t)
	defer storage.Close()

	user1 := &models.User{
		Email:    "same@example.com",
		Password: "pass1",
		Username: "user1",
	}

	user2 := &models.User{
		Email:    "same@example.com", // тот же email
		Password: "pass2",
		Username: "user2",
	}

	err := storage.CreateUser(user1)
	assert.NoError(t, err)

	err = storage.CreateUser(user2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user already exists")
}
