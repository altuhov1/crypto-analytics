package storage

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type PairsFileStorage struct {
	cacheFile string
}

func NewPairsFileStorage(filename string) *PairsFileStorage {
	return &PairsFileStorage{
		cacheFile: filename,
	}
}

func (p *PairsFileStorage) Save(data []byte, amountPairs int) error {
	if err := p.ensureCacheDir(); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	if err := os.WriteFile(p.cacheFile, data, 0644); err != nil {
		return err
	}

	return nil
}

// ensureCacheDir создает директорию для кэш файла если её нет
func (p *PairsFileStorage) ensureCacheDir() error {
	dir := filepath.Dir(p.cacheFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	slog.Debug("Cache directory ensured", "path", dir)
	return nil
}

func (p *PairsFileStorage) Load() ([]string, error) {
	data, err := os.ReadFile(p.cacheFile)
	if err != nil {
		return nil, err
	}
	pairs := make([]string, 0, 250)
	err = json.Unmarshal(data, &pairs)
	return pairs, err
}
