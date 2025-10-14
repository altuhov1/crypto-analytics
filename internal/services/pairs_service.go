package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CryptoPairsService struct {
	cacheFile     string
	pairs         []string
	isInitialized bool
}

type PairsResponse struct {
	Symbols []struct {
		Symbol string `json:"symbol"`
	} `json:"symbols"`
}

const BinanceAPIURL = "https://api.binance.com/api/v3/exchangeInfo"

func NewCryptoPairsService(cacheFile string, downloadOnStart bool) (*CryptoPairsService, error) {
	service := &CryptoPairsService{
		cacheFile: cacheFile,
		pairs:     []string{},
	}

	if downloadOnStart {
		slog.Info("Downloading crypto pairs from API")
		if err := service.downloadAndCachePairs(); err != nil {
			return nil, fmt.Errorf("failed to download pairs: %v", err)
		}
	} else {
		slog.Info("Loading crypto pairs from cache")
		if err := service.loadPairsFromCache(); err != nil {
			slog.Warn("Cache load failed, downloading from API")
			if err := service.downloadAndCachePairs(); err != nil {
				return nil, fmt.Errorf("failed to download pairs: %v", err)
			}
		}
	}

	service.isInitialized = true
	slog.Info("Crypto pairs service initialized", "pairs_count", len(service.pairs))

	return service, nil
}

func (s *CryptoPairsService) downloadAndCachePairs() error {
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(BinanceAPIURL)
	if err != nil {
		return fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var apiResponse PairsResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Фильтруем только USDT пары
	s.pairs = s.filterUSDTOairs(apiResponse)

	// Сохраняем в кэш
	data, err := json.Marshal(s.pairs)
	if err != nil {
		return err
	}

	// Создаем директории рекурсивно
	if err := s.ensureCacheDir(); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	if err := os.WriteFile(s.cacheFile, data, 0644); err != nil {
		return err
	}

	slog.Info("USDT pairs downloaded and cached", "count", len(s.pairs))
	return nil
}

// ensureCacheDir создает директорию для кэш файла если её нет
func (s *CryptoPairsService) ensureCacheDir() error {
	dir := filepath.Dir(s.cacheFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	slog.Debug("Cache directory ensured", "path", dir)
	return nil
}

// filterUSDTOairs оставляет только пары, заканчивающиеся на USDT
func (s *CryptoPairsService) filterUSDTOairs(response PairsResponse) []string {
	var usdtPairs []string

	for _, symbol := range response.Symbols {
		if strings.HasSuffix(symbol.Symbol, "USDT") {
			usdtPairs = append(usdtPairs, symbol.Symbol)
		}
	}

	slog.Info("Filtered USDT pairs", "total", len(response.Symbols), "usdt_pairs", len(usdtPairs))
	return usdtPairs
}

func (s *CryptoPairsService) loadPairsFromCache() error {
	data, err := os.ReadFile(s.cacheFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.pairs)
}

func (s *CryptoPairsService) GetAllPairs() ([]string, error) {
	if !s.isInitialized {
		return nil, fmt.Errorf("service not initialized")
	}
	return s.pairs, nil
}

func (s *CryptoPairsService) GetPairsCount() int {
	return len(s.pairs)
}
