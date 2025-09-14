package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"webdev-90-days/internal/models" // Импортируем модели
)

// FileStorage представляет хранилище в файле
type FileStorage struct {
	filePath string
}

// Storage определяет контракт для работы с данными.
// Теперь наш обработчик будет зависеть от этого интерфейса, а не от конкретной реализации (FileStorage).
type Storage interface {
	SaveContact(contact *models.ContactForm) error
	// GetContact(id int) (*models.ContactForm, error) // На будущее
	// GetAllContacts() ([]*models.ContactForm, error) // На будущее
}

// NewFileStorage создает новый экземпляр FileStorage
func NewFileStorage(filePath string) (*FileStorage, error) {
	// Создаем все директории в пути, если их нет
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}
	return &FileStorage{filePath: filePath}, nil
}

// SaveContact сохраняет контакт в файл
func (s *FileStorage) SaveContact(contact *models.ContactForm) error {
	data, err := json.MarshalIndent(contact, "", "  ")
	if err != nil {
		return err
	}

	file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(string(data) + ",\n")
	return err
}
