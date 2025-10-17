package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"
	"webdev-90-days/internal/models"
	"webdev-90-days/internal/storage"
)

type AnalysisService struct {
	binanceService *BinanceService
	storage        *storage.AnalysisStorage
	useRealAPI     bool
	preloadedData  map[string]*models.AnalysisData
	mu             sync.RWMutex
}

type BinanceService struct {
	client *http.Client
}

// NewAnalysisService создает сервис с упрощенной логикой
func NewAnalysisService(useRealAPI bool) *AnalysisService {
	binanceService := NewBinanceService()
	analysisStorage := storage.NewAnalysisStorage()

	service := &AnalysisService{
		binanceService: binanceService,
		storage:        analysisStorage,
		useRealAPI:     useRealAPI,
		preloadedData:  make(map[string]*models.AnalysisData),
	}

	// Инициализируем данные согласно логике
	service.initializeData()

	return service
}

func NewBinanceService() *BinanceService {
	return &BinanceService{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// initializeData инициализирует данные согласно флагу
func (s *AnalysisService) initializeData() {
	if s.useRealAPI {
		// Режим true: загружаем из API в ОЗУ и файл + запускаем обновление
		slog.Info("🔧 Initializing data for REAL API mode")
		s.loadAllDataFromAPI()
		s.startPeriodicUpdate()
	} else {
		// Режим false: загружаем из файла в ОЗУ, если нет - ошибка
		slog.Info("🔧 Initializing data for TEST mode")
		s.loadAllDataFromFile()
	}
}

// loadAllDataFromAPI загружает все данные из API в ОЗУ и файл
func (s *AnalysisService) loadAllDataFromAPI() {
	pairs := s.GetAvailablePairs()
	timeframes := s.GetAvailableTimeframes()

	var wg sync.WaitGroup
	successCount := 0
	failCount := 0

	for _, pair := range pairs {
		for _, timeframe := range timeframes {
			wg.Add(1)

			go func(p, tf string) {
				defer wg.Done()

				slog.Info("🌐 Loading from API", "pair", p, "timeframe", tf)

				candles, err := s.fetchCandlesFromBinance(p, tf)
				if err != nil {
					slog.Error("❌ Failed to load from API", "pair", p, "timeframe", tf, "error", err)
					failCount++
					return
				}

				indicators := s.calculateIndicators(candles)

				data := &models.AnalysisData{
					Pair:       p,
					Timeframe:  tf,
					Candles:    candles,
					Indicators: indicators,
					Timestamp:  time.Now().Unix(),
				}

				// Сохраняем в ОЗУ
				key := s.getCacheKey(p, tf)
				s.mu.Lock()
				s.preloadedData[key] = data
				s.mu.Unlock()

				// Сохраняем в файл
				if err := s.storage.SaveAnalysisData(p, tf, data); err != nil {
					slog.Error("❌ Failed to save to file", "pair", p, "timeframe", tf, "error", err)
				} else {
					slog.Info("✅ Loaded from API and saved", "pair", p, "timeframe", tf)
					successCount++
				}

				// Задержка чтобы не спамить API
				time.Sleep(100 * time.Millisecond)
			}(pair, timeframe)
		}
	}

	wg.Wait()
	slog.Info("🎉 API data loading completed",
		"success", successCount,
		"failed", failCount,
		"total", len(pairs)*len(timeframes))
}

// loadAllDataFromFile загружает все данные из файла в ОЗУ
func (s *AnalysisService) loadAllDataFromFile() {
	pairs := s.GetAvailablePairs()
	timeframes := s.GetAvailableTimeframes()

	loadedCount := 0
	errorCount := 0

	for _, pair := range pairs {
		for _, timeframe := range timeframes {
			// Пробуем загрузить из файла
			cachedData, err := s.storage.LoadAnalysisData(pair, timeframe)
			if err != nil {
				slog.Error("❌ Failed to load from file", "pair", pair, "timeframe", timeframe, "error", err)
				errorCount++
				continue
			}

			if cachedData == nil {
				slog.Error("❌ No data in file", "pair", pair, "timeframe", timeframe)
				errorCount++
				continue
			}

			// Успешно загрузили из файла
			key := s.getCacheKey(pair, timeframe)
			s.mu.Lock()
			s.preloadedData[key] = cachedData
			s.mu.Unlock()
			loadedCount++
			slog.Info("📁 Loaded from file", "pair", pair, "timeframe", timeframe)
		}
	}

	slog.Info("🎉 File data loading completed",
		"loaded", loadedCount,
		"errors", errorCount,
		"total", len(pairs)*len(timeframes))

	if errorCount > 0 {
		slog.Error("🚨 Some data failed to load from file. Please run with useRealAPI=true first to populate data.")
	}
}

// startPeriodicUpdate запускает периодическое обновление данных каждые 2 часа
func (s *AnalysisService) startPeriodicUpdate() {
	slog.Info("⏰ Starting periodic data update every 2 hours")

	ticker := time.NewTicker(2 * time.Hour)

	go func() {
		for {
			select {
			case <-ticker.C:
				slog.Info("🔄 Starting scheduled data update from API")
				s.loadAllDataFromAPI() // Просто перезагружаем все данные
			}
		}
	}()
}

// getCacheKey создает ключ для кэша в памяти
func (s *AnalysisService) getCacheKey(pair, timeframe string) string {
	return pair + "_" + timeframe
}

// GetAnalysisData возвращает данные ТОЛЬКО из ОЗУ
func (s *AnalysisService) GetAnalysisData(pair, timeframe string) (*models.AnalysisData, error) {
	key := s.getCacheKey(pair, timeframe)

	s.mu.RLock()
	data, exists := s.preloadedData[key]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("data not found for pair %s and timeframe %s", pair, timeframe)
	}

	slog.Info("⚡ Returning data from memory", "pair", pair, "timeframe", timeframe)
	return data, nil
}

// RefreshData принудительно обновляет данные для конкретной пары
func (s *AnalysisService) RefreshData(pair, timeframe string) error {
	if !s.useRealAPI {
		return fmt.Errorf("refresh is only available in REAL API mode")
	}

	slog.Info("🔄 Manual refresh requested", "pair", pair, "timeframe", timeframe)

	candles, err := s.fetchCandlesFromBinance(pair, timeframe)
	if err != nil {
		return fmt.Errorf("failed to refresh data: %v", err)
	}

	indicators := s.calculateIndicators(candles)

	data := &models.AnalysisData{
		Pair:       pair,
		Timeframe:  timeframe,
		Candles:    candles,
		Indicators: indicators,
		Timestamp:  time.Now().Unix(),
	}

	// Обновляем в ОЗУ
	key := s.getCacheKey(pair, timeframe)
	s.mu.Lock()
	s.preloadedData[key] = data
	s.mu.Unlock()

	// Обновляем в файле
	if err := s.storage.SaveAnalysisData(pair, timeframe, data); err != nil {
		return fmt.Errorf("failed to save refreshed data: %v", err)
	}

	slog.Info("✅ Data refreshed successfully", "pair", pair, "timeframe", timeframe)
	return nil
}

// GetPreloadedStatus возвращает статус загруженных данных
func (s *AnalysisService) GetPreloadedStatus() map[string]bool {
	status := make(map[string]bool)

	pairs := s.GetAvailablePairs()
	timeframes := s.GetAvailableTimeframes()

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, pair := range pairs {
		for _, timeframe := range timeframes {
			key := s.getCacheKey(pair, timeframe)
			status[key] = s.preloadedData[key] != nil
		}
	}

	return status
}

// Остальные методы без изменений
func (s *AnalysisService) fetchCandlesFromBinance(pair, timeframe string) ([]models.Candle, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=%s&limit=500", pair, timeframe)

	resp, err := s.binanceService.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("binance API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance API returned status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var klines [][]interface{}
	if err := json.Unmarshal(body, &klines); err != nil {
		return nil, fmt.Errorf("failed to parse binance response: %v", err)
	}

	if len(klines) == 0 {
		return nil, fmt.Errorf("no data returned from binance for pair %s", pair)
	}

	var candles []models.Candle
	for _, kline := range klines {
		if len(kline) < 12 {
			continue
		}

		candle, err := s.parseBinanceKline(kline)
		if err != nil {
			slog.Warn("⚠️ Failed to parse kline", "error", err)
			continue
		}

		candles = append(candles, candle)
	}

	return candles, nil
}

func (s *AnalysisService) parseBinanceKline(kline []interface{}) (models.Candle, error) {
	candle := models.Candle{}

	if openTime, ok := kline[0].(float64); ok {
		candle.OpenTime = int64(openTime)
	}
	if openStr, ok := kline[1].(string); ok {
		if open, err := strconv.ParseFloat(openStr, 64); err == nil {
			candle.Open = open
		}
	}
	if highStr, ok := kline[2].(string); ok {
		if high, err := strconv.ParseFloat(highStr, 64); err == nil {
			candle.High = high
		}
	}
	if lowStr, ok := kline[3].(string); ok {
		if low, err := strconv.ParseFloat(lowStr, 64); err == nil {
			candle.Low = low
		}
	}
	if closeStr, ok := kline[4].(string); ok {
		if close, err := strconv.ParseFloat(closeStr, 64); err == nil {
			candle.Close = close
		}
	}
	if volumeStr, ok := kline[5].(string); ok {
		if volume, err := strconv.ParseFloat(volumeStr, 64); err == nil {
			candle.Volume = volume
		}
	}
	if closeTime, ok := kline[6].(float64); ok {
		candle.CloseTime = int64(closeTime)
	}

	return candle, nil
}

func (s *AnalysisService) calculateIndicators(candles []models.Candle) models.TechnicalIndicators {
	if len(candles) < 50 {
		return models.TechnicalIndicators{}
	}

	closes := make([]float64, len(candles))
	for i, candle := range candles {
		closes[i] = candle.Close
	}

	return models.TechnicalIndicators{
		SMA20: s.calculateSMA(closes, 20),
		SMA50: s.calculateSMA(closes, 50),
		EMA12: s.calculateEMA(closes, 12),
		EMA26: s.calculateEMA(closes, 26),
		RSI:   s.calculateRSI(closes, 14),
		MACD:  s.calculateMACD(closes),
	}
}

func (s *AnalysisService) calculateSMA(data []float64, period int) float64 {
	if len(data) < period {
		return 0
	}
	sum := 0.0
	for i := len(data) - period; i < len(data); i++ {
		sum += data[i]
	}
	return sum / float64(period)
}

func (s *AnalysisService) calculateEMA(data []float64, period int) float64 {
	if len(data) < period {
		return 0
	}
	multiplier := 2.0 / float64(period+1)
	ema := s.calculateSMA(data[:period], period)
	for i := period; i < len(data); i++ {
		ema = (data[i]-ema)*multiplier + ema
	}
	return ema
}

func (s *AnalysisService) calculateRSI(data []float64, period int) float64 {
	if len(data) <= period {
		return 50.0
	}
	gains, losses := 0.0, 0.0
	for i := len(data) - period; i < len(data)-1; i++ {
		change := data[i+1] - data[i]
		if change > 0 {
			gains += change
		} else {
			losses -= change
		}
	}
	avgGain, avgLoss := gains/float64(period), losses/float64(period)
	if avgLoss == 0 {
		return 100.0
	}
	rs := avgGain / avgLoss
	rsi := 100.0 - (100.0 / (1.0 + rs))
	if rsi > 100 {
		return 100
	}
	if rsi < 0 {
		return 0
	}
	return rsi
}

func (s *AnalysisService) calculateMACD(data []float64) float64 {
	ema12, ema26 := s.calculateEMA(data, 12), s.calculateEMA(data, 26)
	return ema12 - ema26
}

func (s *AnalysisService) GetAvailablePairs() []string {
	return []string{
		"BTCUSDT", "ETHUSDT", "ADAUSDT", "DOTUSDT", "LINKUSDT",
		"BNBUSDT", "XRPUSDT", "SOLUSDT", "DOGEUSDT", "MATICUSDT",
	}
}

func (s *AnalysisService) GetAvailableTimeframes() []string {
	return []string{"1m", "5m", "15m", "1h", "4h", "1d"}
}
