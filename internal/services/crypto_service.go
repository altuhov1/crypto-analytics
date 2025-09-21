package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"webdev-90-days/internal/models"
)

type CryptoService struct {
	baseURL string
	client  *http.Client
}

func NewCryptoService() *CryptoService {
	return &CryptoService{
		baseURL: "https://api.coingecko.com/api/v3",
		client: &http.Client{
			Timeout: 10 * time.Second, // Таймаут для запросов
		},
	}
}

// GetTopCryptos получает топ N криптовалют по рыночной капитализации
func (s *CryptoService) GetTopCryptos(limit int) ([]models.Coin, error) {
	url := fmt.Sprintf("%s/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=%d&page=1&sparkline=false",
		s.baseURL, limit)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к CoinGecko: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CoinGecko API вернул статус: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	var coins []models.Coin
	if err := json.Unmarshal(body, &coins); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	log.Printf("Успешно получено %d криптовалют", len(coins))
	return coins, nil
}

// GetCoinPrice получает цену конкретной криптовалюты
// func (s *CryptoService) GetCoinPrice(coinID string) (float64, error) {
// 	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", s.baseURL, coinID)

// 	resp, err := s.client.Get(url)
// 	if err != nil {
// 		return 0, fmt.Errorf("ошибка запроса: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	var result map[string]map[string]float64
// 	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
// 		return 0, fmt.Errorf("ошибка парсинга JSON: %w", err)
// 	}

// 	if coinData, exists := result[coinID]; exists {
// 		return coinData["usd"], nil
// 	}

// 	return 0, fmt.Errorf("криптовалюта не найдена")
// }
