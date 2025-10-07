package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"webdev-90-days/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserPostgresStorage struct {
	pool *pgxpool.Pool
}

func NewUserPostgresStorage(config PGXConfig) (*UserPostgresStorage, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
		config.SSLMode,
	)

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &UserPostgresStorage{pool: pool}, nil
}

func (s *UserPostgresStorage) Close() {
	s.pool.Close()
}

func (s *UserPostgresStorage) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (email, password, username, favorite_coins) 
		VALUES ($1, $2, $3, $4)
	`

	_, err := s.pool.Exec(context.Background(), query, user.Email, user.Password, user.Username, user.FavoriteCoins)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			if strings.Contains(err.Error(), "users_email_key") {
				return fmt.Errorf("user already exists")
			}
			if strings.Contains(err.Error(), "users_username_key") {
				return fmt.Errorf("user name already exists")
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *UserPostgresStorage) GetUserByName(nameU string) (*models.User, error) {
	query := `
		SELECT email, password, username, favorite_coins 
		FROM users 
		WHERE username = $1
	`

	user := &models.User{}
	var favoriteCoins []string

	err := s.pool.QueryRow(context.Background(), query, nameU).Scan(
		&user.Email,
		&user.Password,
		&user.Username,
		&favoriteCoins,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.FavoriteCoins = favoriteCoins
	return user, nil
}

func (s *UserPostgresStorage) GetAllFavoriteCoins(nameU string) ([]string, error) {
	query := `
		SELECT favorite_coins 
		FROM users 
		WHERE username = $1
	`

	var favoriteCoins []string
	err := s.pool.QueryRow(context.Background(), query, nameU).Scan(&favoriteCoins)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get favorite coins: %w", err)
	}

	return favoriteCoins, nil
}

func (s *UserPostgresStorage) NewFavoriteCoin(nameU string, nameCoin string) error {
	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	// Проверяем существование пользователя
	var exists bool
	err = tx.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", nameU).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Проверяем, есть ли уже монета в избранном
	var coinExists bool
	err = tx.QueryRow(context.Background(), `
		SELECT EXISTS(
			SELECT 1 FROM users 
			WHERE username = $1 AND $2 = ANY(favorite_coins)
		)`, nameU, nameCoin).Scan(&coinExists)
	if err != nil {
		return fmt.Errorf("failed to check coin existence: %w", err)
	}
	if coinExists {
		return fmt.Errorf("coin already in list of favorite coins")
	}

	// Добавляем монету в избранное
	_, err = tx.Exec(context.Background(), `
		UPDATE users 
		SET favorite_coins = array_append(favorite_coins, $1) 
		WHERE username = $2`, nameCoin, nameU)
	if err != nil {
		return fmt.Errorf("failed to add favorite coin: %w", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *UserPostgresStorage) RemoveFavoriteCoin(nameU string, nameCoin string) error {
	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	// Проверяем существование пользователя
	var exists bool
	err = tx.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", nameU).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Проверяем, есть ли монета в избранном
	var coinExists bool
	err = tx.QueryRow(context.Background(), `
		SELECT EXISTS(
			SELECT 1 FROM users 
			WHERE username = $1 AND $2 = ANY(favorite_coins)
		)`, nameU, nameCoin).Scan(&coinExists)
	if err != nil {
		return fmt.Errorf("failed to check coin existence: %w", err)
	}
	if !coinExists {
		return fmt.Errorf("coin does not in list")
	}

	// Удаляем монету из избранного
	_, err = tx.Exec(context.Background(), `
		UPDATE users 
		SET favorite_coins = array_remove(favorite_coins, $1) 
		WHERE username = $2`, nameCoin, nameU)
	if err != nil {
		return fmt.Errorf("failed to remove favorite coin: %w", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

type PublicUser struct {
	ID            int       `json:"id"`
	Username      string    `json:"username"`
	FavoriteCoins []string  `json:"favorite_coins"`
	CreatedAt     time.Time `json:"created_at"`
}

// ExportUsersToJSON экспортирует всех пользователей (без паролей и email) в JSON файл
func (s *UserPostgresStorage) ExportUsersToJSON(filename string) error {
	// Выполняем запрос к БД
	rows, err := s.pool.Query(context.Background(), `
        SELECT id, username, favorite_coins, created_at 
        FROM users 
        ORDER BY id
    `)
	if err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []PublicUser

	// Читаем результаты
	for rows.Next() {
		var user PublicUser
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.FavoriteCoins,
			&user.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	// Проверяем ошибки итерации
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error during rows iteration: %w", err)
	}

	// Создаем JSON файл
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Настраиваем JSON encoder для красивого вывода
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	// Записываем данные в файл
	if err := encoder.Encode(users); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	slog.Info("Successfully exported users to",
		"amount", len(users),
		"filename", filename)
	return nil
}
