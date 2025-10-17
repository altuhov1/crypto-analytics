package storage

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// FileCacheStorage реализация CacheStorage для файловой системы
type FileCacheStorage struct {
	cacheFile string
}

// NewFileCacheStorage создает новый файловый кэш
func NewFileCacheStorage(cacheFile string) *FileCacheStorage {
	return &FileCacheStorage{
		cacheFile: cacheFile,
	}
}

// Save сохраняет пары в файл
func (f *FileCacheStorage) Save(pairs []string) error {
	data, err := json.Marshal(pairs)
	if err != nil {
		return err
	}

	// Создаем директории рекурсивно
	if err := f.ensureCacheDir(); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	if err := os.WriteFile(f.cacheFile, data, 0644); err != nil {
		return err
	}

	slog.Debug("Pairs saved to cache", "file", f.cacheFile, "count", len(pairs))
	return nil
}

// Load загружает пары из файла
func (f *FileCacheStorage) Load() ([]string, error) {
	data, err := os.ReadFile(f.cacheFile)
	if err != nil {
		return nil, err
	}

	var pairs []string
	if err := json.Unmarshal(data, &pairs); err != nil {
		return nil, err
	}

	slog.Debug("Pairs loaded from cache", "file", f.cacheFile, "count", len(pairs))
	return pairs, nil
}

// ensureCacheDir создает директорию для кэш файла если её нет
func (f *FileCacheStorage) ensureCacheDir() error {
	dir := filepath.Dir(f.cacheFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	slog.Debug("Cache directory ensured", "path", dir)
	return nil
}
