package storage

import (
	"encoding/json"
	"os"
	"webdev-90-days/internal/models"
)

type AnalysisFileStorage struct {
	filePath string
}

func NewAnalysisFileStorage(filePath string) *AnalysisFileStorage {
	return &AnalysisFileStorage{
		filePath: filePath,
	}
}

func (s *AnalysisFileStorage) SaveAnalysisData(data models.PairsCrypto) error {
	file, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (s *AnalysisFileStorage) LoadAnalysisData() (models.PairsCrypto, error) {
	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return models.PairsCrypto{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var data models.PairsCrypto
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
