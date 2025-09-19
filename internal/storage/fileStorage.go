// storage/file_storage.go
package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"webdev-90-days/internal/models"
)

type FileStorage struct {
	filePath string
	mu       sync.Mutex // Для потокобезопасности
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}
	return &FileStorage{filePath: filePath}, nil
}

func (s *FileStorage) SaveContact(contact *models.ContactForm) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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

func (s *FileStorage) Close() error {
	// Для файлового хранилища не нужно закрывать соединение
	return nil
}
