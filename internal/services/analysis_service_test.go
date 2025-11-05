package services

import (
	"crypto-analytics/internal/models"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

type MockTempStorage struct {
	Response *models.AnalysisData
	Error    error
}

func (m *MockTempStorage) GetAnalysisData(pair, timeframe string) (*models.AnalysisData, error) {
	return m.Response, m.Error
}

func (m *MockTempStorage) SaveAnalysisData(data models.AnalysisData) error {
	return nil
}

func (m *MockTempStorage) SavePairs(pairs models.PairsCrypto) error {
	return nil
}

func (m *MockTempStorage) GetStats() string {
	return "test stats"
}

func (m *MockTempStorage) Close(client *redis.Client) {}

type MockStorage struct {
	Response models.PairsCrypto
	Error    error
}

func (m *MockStorage) SaveAnalysisData(data models.PairsCrypto) error {
	return nil
}

func (m *MockStorage) LoadAnalysisData() (models.PairsCrypto, error) {
	return m.Response, m.Error
}

func TestAnalysisService_GetPairInfo_Simple(t *testing.T) {
	createTestData := func(pair, timeframe string) *models.AnalysisData {
		return &models.AnalysisData{
			Pair:      pair,
			Timeframe: timeframe,
			Candles: []models.Candle{
				{
					OpenTime:  time.Now().Unix(),
					CloseTime: time.Now().Add(time.Hour).Unix(),
					Open:      50000.0,
					High:      51000.0,
					Low:       49000.0,
					Close:     50500.0,
					Volume:    1000.0,
				},
			},
			Indicators: models.TechnicalIndicators{
				SMA20:     50200.0,
				SMA50:     50000.0,
				EMA12:     50300.0,
				EMA26:     50100.0,
				RSI:       65.5,
				MACD:      200.0,
				Signal:    180.0,
				Histogram: 20.0,
			},
			Timestamp: time.Now().Unix(),
		}
	}

	tests := []struct {
		name          string
		pair          string
		timeframe     string
		mockResponse  *models.AnalysisData
		mockError     error
		expectedData  *models.AnalysisData
		expectedError string
	}{
		{
			name:          "successful get BTCUSDT",
			pair:          "BTCUSDT",
			timeframe:     "1h",
			mockResponse:  createTestData("BTCUSDT", "1h"),
			mockError:     nil,
			expectedData:  createTestData("BTCUSDT", "1h"),
			expectedError: "",
		},
		{
			name:          "successful get ETHUSDT",
			pair:          "ETHUSDT",
			timeframe:     "5m",
			mockResponse:  createTestData("ETHUSDT", "5m"),
			mockError:     nil,
			expectedData:  createTestData("ETHUSDT", "5m"),
			expectedError: "",
		},
		{
			name:          "data not found in temp store",
			pair:          "UNKNOWN",
			timeframe:     "1h",
			mockResponse:  nil,
			mockError:     errors.New("not found"),
			expectedData:  nil,
			expectedError: "данные для пары UNKNOWN и таймфрейма 1h не найдены",
		},
		{
			name:          "temp store returns error",
			pair:          "BTCUSDT",
			timeframe:     "1h",
			mockResponse:  nil,
			mockError:     errors.New("storage error"),
			expectedData:  nil,
			expectedError: "данные для пары BTCUSDT и таймфрейма 1h не найдены",
		},
		{
			name:          "empty pair",
			pair:          "",
			timeframe:     "1h",
			mockResponse:  nil,
			mockError:     nil,
			expectedData:  nil,
			expectedError: "",
		},
		{
			name:          "empty timeframe",
			pair:          "BTCUSDT",
			timeframe:     "",
			mockResponse:  nil,
			mockError:     nil,
			expectedData:  nil,
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем моки
			mockTempStorage := &MockTempStorage{
				Response: tt.mockResponse,
				Error:    tt.mockError,
			}

			mockStorage := &MockStorage{
				Response: models.PairsCrypto{},
				Error:    nil,
			}

			// Создаем сервис
			service := &AnalysisService{
				tempStore:  mockTempStorage,
				store:      mockStorage,
				binanceAPI: NewBinanceAPI(),
				goToApi:    false,
			}

			// Вызываем метод
			result, err := service.GetPairInfo(tt.pair, tt.timeframe)

			// Проверяем ошибку
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Проверяем данные
			if tt.expectedData != nil {
				if result == nil {
					t.Errorf("Expected data, got nil")
				} else {
					if result.Pair != tt.expectedData.Pair {
						t.Errorf("Expected pair %s, got %s", tt.expectedData.Pair, result.Pair)
					}
					if result.Timeframe != tt.expectedData.Timeframe {
						t.Errorf("Expected timeframe %s, got %s", tt.expectedData.Timeframe, result.Timeframe)
					}
				}
			} else {
				if result != nil {
					t.Errorf("Expected nil data, got %v", result)
				}
			}
		})
	}
}
func TestAnalysisService_calculateRSI_NewAlgorithm(t *testing.T) {
	createCandles := func(prices []float64) []models.Candle {
		candles := make([]models.Candle, len(prices))
		for i, price := range prices {
			candles[i] = models.Candle{Close: price}
		}
		return candles
	}

	tests := []struct {
		name     string
		candles  []models.Candle
		period   int
		expected float64
	}{
		{
			name:     "not enough candles",
			candles:  createCandles([]float64{100, 101, 102}),
			period:   14,
			expected: 50.0,
		},
		{
			name:     "exactly period candles",
			candles:  createCandles([]float64{100, 101, 102, 103}),
			period:   4,
			expected: 50.0,
		},
		{
			name: "all gains",
			candles: createCandles([]float64{
				100, 101, 102, 103, 104, 105,
			}),
			period:   3,
			expected: 100.0,
		},
		{
			name: "all losses",
			candles: createCandles([]float64{
				105, 104, 103, 102, 101, 100,
			}),
			period:   3,
			expected: 0.0,
		},
		{
			name: "equal gains and losses",
			candles: createCandles([]float64{
				100, 102,
				100,
				102,
				100,
			}),
			period:   2,
			expected: 50.0,
		},
		{
			name: "more gains than losses",
			candles: createCandles([]float64{
				100, 102,
				101,
				103,
			}),
			period:   2,
			expected: 66.67,
		},
		{
			name: "real RSI calculation",
			candles: createCandles([]float64{
				100, 102,
				101,
				103,
				102,
				105,
			}),
			period:   4,
			expected: 71.43,
		},
		{
			name:     "empty candles",
			candles:  []models.Candle{},
			period:   14,
			expected: 50.0,
		},
		{
			name:     "single candle",
			candles:  createCandles([]float64{100}),
			period:   14,
			expected: 50.0,
		},
	}

	service := &AnalysisService{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateRSI(tt.candles, tt.period)

			if abs(result-tt.expected) > 0.01 {
				t.Errorf("calculateRSI() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
