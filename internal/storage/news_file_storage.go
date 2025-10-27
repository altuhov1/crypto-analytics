package storage

import (
	"crypto-analytics/internal/models"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type NewsFileStorage struct {
	filename string
	mu       sync.RWMutex
}

func NewNewsFileStorage(filename string) *NewsFileStorage {
	return &NewsFileStorage{
		filename: filename,
	}
}

func (s *NewsFileStorage) loadNews() ([]models.NewsItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, err := os.Open(s.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.NewsItem{}, nil
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var news []models.NewsItem
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&news); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return news, nil
}

func (s *NewsFileStorage) saveNews(news []models.NewsItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Create(s.filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(news); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// generateID генерирует уникальный ID на основе GUID, Title и PublishedAt
func (s *NewsFileStorage) generateID(item models.NewsItem) string {
	if item.GUID != "" {
		return item.GUID
	}
	data := item.Title + item.PublishedAt
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:16])
}

func (s *NewsFileStorage) AddNews(items []models.NewsItem) error {
	if len(items) == 0 {
		return nil
	}

	existingNews, err := s.loadNews()
	if err != nil {
		return err
	}

	// Создаем мапу для быстрого поиска существующих новостей
	existingMap := make(map[string]bool)
	for _, item := range existingNews {
		existingMap[s.generateID(item)] = true
	}

	// Добавляем только новые новости
	for _, item := range items {
		id := s.generateID(item)
		if !existingMap[id] {
			existingNews = append(existingNews, item)
			existingMap[id] = true
		}
	}

	return s.saveNews(existingNews)
}

func (s *NewsFileStorage) GetAllNews() ([]models.NewsItem, error) {
	return s.loadNews()
}

func (s *NewsFileStorage) UpdateNews(items []models.NewsItem) error {
	// Для файлового хранилища обновление = добавление новых записей
	return s.AddNews(items)
}
