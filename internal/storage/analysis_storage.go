package storage

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"webdev-90-days/internal/models"
)

type AnalysisStorage struct {
	analysisFile string // полный путь к файлу
}

func NewAnalysisStorage(filename string) *AnalysisStorage {
	// Создаем файл в стандартной папке storage

	return &AnalysisStorage{
		analysisFile: filename,
	}
}

func (s *AnalysisStorage) SaveAnalysisData(pair, timeframe string, data *models.AnalysisData) error {
	// Создаем директорию если нужно
	dir := filepath.Dir(s.analysisFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	file, err := os.Create(s.analysisFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Используем энкодер для постепенной записи
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode data: %v", err)
	}

	slog.Info("Analysis data saved", "pair", pair, "timeframe", timeframe, "file", s.analysisFile)
	return nil
}

func (s *AnalysisStorage) LoadAnalysisData(pair, timeframe string) (*models.AnalysisData, error) {
	file, err := os.Open(s.analysisFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Файл не существует - это нормально
		}
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var data models.AnalysisData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode data: %v", err)
	}

	// Проверяем, что данные соответствуют запрошенной паре и таймфрейму
	if data.Pair != pair || data.Timeframe != timeframe {
		slog.Info("Cached data doesn't match request",
			"cached_pair", data.Pair, "requested_pair", pair,
			"cached_timeframe", data.Timeframe, "requested_timeframe", timeframe)
		return nil, nil
	}

	slog.Info("Analysis data loaded from cache", "pair", pair, "timeframe", timeframe)
	return &data, nil
}
