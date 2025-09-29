package storage

import (
	"sync"
)

type NewsFileStorage struct {
	filename string
	mu       sync.RWMutex
}

func NewNewsFileStorage(filename string) *NewsFileStorage {
	return &NewsFileStorage{
		filename: filename,
		mu:       sync.RWMutex{},
	}
}

// func (s *NewsFileStorage) loadNews() ([]*models.NewsItem, error) {
// 	s.mu.RLock()
// 	defer s.mu.RUnlock()

// 	file, err := os.Open(s.filename)
// 	if err != nil {
// 		if os.IsNotExist(err) {
// 			return []*models.NewsItem{}, nil
// 		}
// 		return nil, fmt.Errorf("failed to open file: %w", err)
// 	}
// 	defer file.Close()

// 	var news []*models.NewsItem
// 	decoder := json.NewDecoder(file)
// 	if err := decoder.Decode(&news); err != nil {
// 		return nil, fmt.Errorf("failed to decode JSON: %w", err)
// 	}

// 	return news, nil
// }

// func (s *NewsFileStorage) saveNews(news []*models.NewsItem) error {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	file, err := os.Create(s.filename)
// 	if err != nil {
// 		return fmt.Errorf("failed to create file: %w", err)
// 	}
// 	defer file.Close()

// 	encoder := json.NewEncoder(file)
// 	encoder.SetIndent("", "  ")
// 	if err := encoder.Encode(news); err != nil {
// 		return fmt.Errorf("failed to encode JSON: %w", err)
// 	}

// 	return nil
// }
