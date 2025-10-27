package services

import (
	"crypto-analytics/internal/storage"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type CryptoPairsService struct {
	store         storage.CacheStorage
	pairs         []string
	isInitialized bool
}

type PairsResponse struct {
	Symbols []struct {
		Symbol string `json:"symbol"`
	} `json:"symbols"`
}

const BinanceAPIURL = "https://api.binance.com/api/v3/exchangeInfo"

func NewCryptoPairsService(storePairs storage.CacheStorage,
	downloadOnStart bool) *CryptoPairsService {
	service := &CryptoPairsService{
		store: storePairs,
		pairs: []string{},
	}

	if downloadOnStart {
		slog.Info("Downloading crypto pairs from API")
		if err := service.downloadAndCachePairs(); err != nil {
			slog.Error("Failed to download pairs on startup", "error", err)
		}
	} else {
		slog.Info("Loading crypto pairs from cache")
		if strings, err := service.store.Load(); err == nil {
			service.pairs = strings
		} else {
			slog.Error("Cache load failed, downloading from API", "error", err)
			if err := service.downloadAndCachePairs(); err != nil {
				slog.Error("Failed to download pairs after cache failure", "error", err)
			}
		}
	}

	service.isInitialized = true
	slog.Info("Crypto pairs service initialized", "pairs_count", len(service.pairs))

	return service
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
	return s.store.Save(data, len(s.pairs))

}

// filterUSDTOairs оставляет только пары, заканчивающиеся на USDT
func (s *CryptoPairsService) filterUSDTOairs(response PairsResponse) []string {
	var usdtPairs []string

	for _, symbol := range response.Symbols {
		if strings.HasSuffix(symbol.Symbol, "USDT") {
			usdtPairs = append(usdtPairs, symbol.Symbol)
		}
	}
	return usdtPairs
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
