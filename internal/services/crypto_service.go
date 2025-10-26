package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"webdev-90-days/internal/models"
)

type CryptoService struct {
	baseURL    string
	client     *http.Client
	cache      []models.Coin
	cacheMutex sync.RWMutex
	cacheTime  time.Time
	cacheFile  string
	useAPI     bool // Режим работы: true - API, false - файл
}

func NewCryptoService(useAPI bool, cacheFile string) *CryptoService {
	svc := &CryptoService{
		baseURL:   "https://api.coingecko.com/api/v3",
		client:    &http.Client{Timeout: 10 * time.Second},
		cacheFile: cacheFile,
		useAPI:    useAPI,
	}

	if useAPI {
		// Режим API: загружаем и кэшируем
		svc.refreshCache()
		go svc.startCacheUpdater()
	} else {
		// Режим файла: загружаем из файла
		svc.loadFromFile()
	}

	return svc
}

func (s *CryptoService) loadFromFile() error {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	data, err := os.ReadFile(s.cacheFile)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	if err := json.Unmarshal(data, &s.cache); err != nil {
		return fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	s.cacheTime = time.Now()
	slog.Info("Загружены монеты из файла",
		"кол", len(s.cache),
		"из файла", s.cacheFile)
	return nil
}

func (s *CryptoService) saveToFile() error {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	data, err := json.MarshalIndent(s.cache, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка создания JSON: %w", err)
	}

	if err := os.WriteFile(s.cacheFile, data, 0644); err != nil {
		return fmt.Errorf("ошибка записи файла: %w", err)
	}
	slog.Info("Сохранены монеты в файл",
		"кол", len(s.cache),
		"из файла", s.cacheFile)
	return nil
}

// refreshCache обновляет кэш (с сохранением в файл если нужно)
func (s *CryptoService) refreshCache() {
	coins, err := s.getTopCryptosFromAPI(250)
	if err != nil {
		slog.Error("Ошибка обновления кэша:", "error", err)
		return
	}

	s.cacheMutex.Lock()
	s.cache = coins
	s.cacheTime = time.Now()
	s.cacheMutex.Unlock()

	// Сохраняем в файл для будущего использования
	if err := s.saveToFile(); err != nil {
		slog.Error("Ошибка сохранения в файл:", "error", err)
	}

}

func (s *CryptoService) startCacheUpdater() {
	ticker := time.NewTicker(1 * time.Hour) // Обновление каждый час

	for range ticker.C {
		s.refreshCache()
	}
}

func (s *CryptoService) getTopCryptosFromAPI(limit int) ([]models.Coin, error) {
	url := fmt.Sprintf("%s/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=%d&page=1&sparkline=false",
		s.baseURL, limit)
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API вернул статус: %d", resp.StatusCode)
	}

	var coins []models.Coin
	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return nil, fmt.Errorf("ошибка парсинга: %w", err)
	}

	return coins, nil
}

func (s *CryptoService) GetTopCryptos(limit int) ([]models.Coin, error) {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	if len(s.cache) == 0 {
		return nil, fmt.Errorf("кэш пустой")
	}

	// Возвращаем запрошенное количество монет
	if limit > len(s.cache) {
		limit = len(s.cache)
	}

	return s.cache[:limit], nil
}

func (s *CryptoService) GetCacheInfo() (int, time.Time) {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	return len(s.cache), s.cacheTime
}
