// storage/pgx_storage.go
package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
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

func (s *PGXStorage) SaveContactFrom(contact *models.ContactForm) error {
	query := `
    INSERT INTO contacts (name, email, message)
    VALUES ($1, $2, $3)
    `

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Логируем данные которые пытаемся сохранить
	log.Printf("Attempting to save: Name=%s, Email=%s, Message=%s",
		contact.Name, contact.Email, contact.Message)

	// Выполняем INSERT
	result, err := s.pool.Exec(ctx, query, contact.Name, contact.Email, contact.Message)
	if err != nil {
		// Детальное логирование ошибки
		log.Printf("SQL Error: %v", err)
		log.Printf("Query: %s", query)
		log.Printf("Params: %s, %s, %s", contact.Name, contact.Email, contact.Message)
		return fmt.Errorf("ошибка сохранения: %w", err)
	}

	// Логируем результат
	rowsAffected := result.RowsAffected()
	log.Printf("Save successful. Rows affected: %d", rowsAffected)

	return nil
}

func (s *PGXStorage) Close() error {
	s.pool.Close()
	return nil
}

func (s *PGXStorage) CheckAndCreateTables() error {
	// Проверяем существует ли таблица contacts
	var tableExists bool
	err := s.pool.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'contacts')",
	).Scan(&tableExists)

	if err != nil {
		return err
	}

	if !tableExists {
		log.Println("Таблица contacts не найдена, создаем...")
		// Создаем таблицу
		_, err = s.pool.Exec(context.Background(), `
            CREATE TABLE contacts (
                id SERIAL PRIMARY KEY,
                name TEXT NOT NULL,
                email TEXT NOT NULL,
                message TEXT NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        `)
		if err != nil {
			return err
		}
		log.Println("Таблица contacts создана")
	}

	return nil
}

// ExportContactsToJSON экспортирует контакты в JSON файл (альтернативный вариант)
func (s *PGXStorage) ExportContactsToJSON(filename string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Выполняем запрос для получения всех контактов
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, email, message, created_at
		FROM contacts
		ORDER BY created_at DESC
	`)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	// Создаем файл для записи
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("ошибка создания файла: %w", err)
	}
	defer file.Close()

	var contacts []map[string]interface{}
	var contactCount int

	// Обрабатываем каждую строку результата
	for rows.Next() {
		var contact struct {
			ID        int
			Name      string
			Email     string
			Message   string
			CreatedAt time.Time
		}

		err := rows.Scan(&contact.ID, &contact.Name, &contact.Email, &contact.Message, &contact.CreatedAt)
		if err != nil {
			log.Printf("Ошибка чтения строки: %v", err)
			continue
		}

		// Формируем JSON объект
		contactData := map[string]interface{}{
			"id":         contact.ID,
			"name":       contact.Name,
			"email":      contact.Email,
			"message":    contact.Message,
			"created_at": contact.CreatedAt.Format(time.RFC3339),
		}
		contacts = append(contacts, contactData)
		contactCount++
	}

	// Проверяем ошибки после итерации
	if err := rows.Err(); err != nil {
		return fmt.Errorf("ошибка при чтении строк: %w", err)
	}

	// Записываем данные в файл
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(contacts); err != nil {
		return fmt.Errorf("ошибка кодирования JSON: %w", err)
	}

	log.Printf("Экспорт завершен. Экспортировано %d контактов в JSON файл: %s", contactCount, filename)
	return nil
}

// GetContactsStats возвращает статистику по контактам (дополнительный метод)
func (s *PGXStorage) GetContactsStats() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stats := make(map[string]interface{})

	// Получаем общее количество контактов
	var totalContacts int
	err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM contacts").Scan(&totalContacts)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения количества контактов: %w", err)
	}
	stats["total_contacts"] = totalContacts

	// Получаем дату самого старого контакта
	var oldestContact time.Time
	err = s.pool.QueryRow(ctx, "SELECT MIN(created_at) FROM contacts").Scan(&oldestContact)
	if err == nil {
		stats["oldest_contact"] = oldestContact.Format("2006-01-02")
	}

	// Получаем дату самого нового контакта
	var newestContact time.Time
	err = s.pool.QueryRow(ctx, "SELECT MAX(created_at) FROM contacts").Scan(&newestContact)
	if err == nil {
		stats["newest_contact"] = newestContact.Format("2006-01-02")
	}

	return stats, nil
}
