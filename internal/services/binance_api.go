package services

import (
	"bytes"
	"crypto-analytics/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type BinanceAPI struct {
	baseURL string
	client  *http.Client
}

func NewBinanceAPI() *BinanceAPI {
	return &BinanceAPI{
		baseURL: "https://api.binance.com/api/v3",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type BinanceCandleResponse []interface{}

func (b *BinanceAPI) fetchCandlesFromBinance(symbol, interval string, limit int) ([]models.Candle, error) {
	url := fmt.Sprintf("%s/klines?symbol=%s&interval=%s&limit=%d",
		b.baseURL, symbol, interval, limit)

	slog.Debug("Запрос к Binance API", "url", url)

	resp, err := b.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к Binance API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("inance API вернул статус: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	var rawCandles []BinanceCandleResponse

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields() 

	if err := decoder.Decode(&rawCandles); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	candles := make([]models.Candle, 0, len(rawCandles))
	for _, raw := range rawCandles {
		if len(raw) < 12 {
			continue
		}

		candle, err := b.parseCandle(raw)
		if err != nil {
			slog.Warn("Ошибка парсинга свечи", "error", err)
			continue
		}
		candles = append(candles, candle)
	}

	return candles, nil
}

func (b *BinanceAPI) parseCandle(raw []interface{}) (models.Candle, error) {
	var candle models.Candle

	// OpenTime
	if openTime, ok := raw[0].(float64); ok {
		candle.OpenTime = int64(openTime)
	} else {
		return candle, fmt.Errorf("неверный формат OpenTime: %v", raw[0])
	}

	// Open
	if open, ok := raw[1].(string); ok {
		if _, err := fmt.Sscanf(open, "%f", &candle.Open); err != nil {
			return candle, fmt.Errorf("ошибка парсинга Open: %w", err)
		}
	}

	// High
	if high, ok := raw[2].(string); ok {
		if _, err := fmt.Sscanf(high, "%f", &candle.High); err != nil {
			return candle, fmt.Errorf("ошибка парсинга High: %w", err)
		}
	}

	// Low
	if low, ok := raw[3].(string); ok {
		if _, err := fmt.Sscanf(low, "%f", &candle.Low); err != nil {
			return candle, fmt.Errorf("ошибка парсинга Low: %w", err)
		}
	}

	// Close
	if close, ok := raw[4].(string); ok {
		if _, err := fmt.Sscanf(close, "%f", &candle.Close); err != nil {
			return candle, fmt.Errorf("ошибка парсинга Close: %w", err)
		}
	}

	// Volume
	if volume, ok := raw[5].(string); ok {
		if _, err := fmt.Sscanf(volume, "%f", &candle.Volume); err != nil {
			return candle, fmt.Errorf("ошибка парсинга Volume: %w", err)
		}
	}

	// CloseTime
	if closeTime, ok := raw[6].(float64); ok {
		candle.CloseTime = int64(closeTime)
	}

	return candle, nil
}
