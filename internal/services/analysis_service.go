package services

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
	"webdev-90-days/internal/models"
	"webdev-90-days/internal/storage"
)

type AnalysisService struct {
	Pairs      models.PairsCrypto
	store      storage.AnalysisStorage
	binanceAPI *BinanceAPI
	goToApi    bool
	mu         sync.RWMutex
}

func NewAnalysisService(goToApi bool, store storage.AnalysisStorage) *AnalysisService {
	service := &AnalysisService{
		goToApi:    goToApi,
		store:      store,
		binanceAPI: NewBinanceAPI(),
		mu:         sync.RWMutex{},
		Pairs:      make(models.PairsCrypto, 0),
	}

	// Первоначальная загрузка данных
	if goToApi {
		slog.Info("Загрузка данных из API Binance")
		service.Pairs = service.uploadApi()
		// Запускаем асинхронное обновление каждые 2 часа
		go service.asyncUpdatePairs()
	} else {
		slog.Info("Загрузка данных из хранилища")
		service.Pairs = service.uploadFromStorage()
	}

	slog.Info("Сервис анализа успешно инициализирован",
		"goToApi", goToApi,
		"loadedPairs", len(service.Pairs))

	return service
}

// uploadApi загружает данные из API Binance
func (a *AnalysisService) uploadApi() models.PairsCrypto {
	pairs := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
	timeframes := []string{"5m", "1h"}
	base := make(models.PairsCrypto, 0)

	slog.Info("Начало загрузки данных из API Binance",
		"pairs", pairs,
		"timeframes", timeframes)

	for _, p := range pairs {
		for _, t := range timeframes {
			slog.Debug("Загрузка данных для пары",
				"pair", p,
				"timeframe", t)

			candlesApi := a.fetchFromApi(p, t)

			if len(candlesApi) == 0 {
				slog.Warn("Получено 0 свечей от Binance",
					"pair", p,
					"timeframe", t)
				continue
			}

			analysisData := models.AnalysisData{
				Pair:       p,
				Timeframe:  t,
				Candles:    candlesApi,
				Indicators: a.calcIndicator(candlesApi),
				Timestamp:  time.Now().Unix(),
			}
			base = append(base, analysisData)

			slog.Debug("Данные загружены",
				"pair", p,
				"timeframe", t,
				"candlesCount", len(candlesApi))
		}
	}

	// Сохраняем в хранилище через интерфейс
	if err := a.store.SaveAnalysisData(base); err != nil {
		slog.Error("Ошибка при сохранении данных в хранилище",
			"error", err)
	}

	slog.Info("Успешно загружено данных из API Binance",
		"totalRecords", len(base))

	return base
}

// uploadFromStorage загружает данные из хранилища
func (a *AnalysisService) uploadFromStorage() models.PairsCrypto {
	slog.Info("Загрузка данных из хранилища")

	data, err := a.store.LoadAnalysisData()
	if err != nil {
		slog.Error("Ошибка при загрузке данных из хранилища",
			"error", err)
		return models.PairsCrypto{}
	}

	slog.Info("Успешно загружено данных из хранилища",
		"records", len(data))

	return data
}

// fetchFromApi делает запрос к API Binance для получения свечей
func (a *AnalysisService) fetchFromApi(pair, timeframe string) []models.Candle {
	a.mu.Lock()
	defer a.mu.Unlock()

	slog.Debug("Запрос к API Binance",
		"pair", pair,
		"timeframe", timeframe)

	candles, err := a.binanceAPI.fetchCandlesFromBinance(pair, timeframe, 900)
	if err != nil {
		slog.Error("Ошибка при запросе к Binance API",
			"pair", pair,
			"timeframe", timeframe,
			"error", err)
		return []models.Candle{}
	}

	slog.Debug("API запрос выполнен",
		"pair", pair,
		"timeframe", timeframe,
		"candlesReturned", len(candles))

	return candles
}

// calcIndicator рассчитывает технические индикаторы на основе свечей
func (a *AnalysisService) calcIndicator(candles []models.Candle) models.TechnicalIndicators {
	if len(candles) < 50 {
		slog.Warn("Недостаточно свечей для расчета индикаторов",
			"candlesCount", len(candles),
			"required", 50)
		return models.TechnicalIndicators{}
	}

	slog.Debug("Расчет технических индикаторов",
		"candlesCount", len(candles))

	// Расчет SMA20
	var sum20 float64
	for i := 0; i < 20; i++ {
		sum20 += candles[len(candles)-1-i].Close
	}
	sma20 := sum20 / 20

	// Расчет SMA50
	var sum50 float64
	for i := 0; i < 50; i++ {
		sum50 += candles[len(candles)-1-i].Close
	}
	sma50 := sum50 / 50

	// Расчет EMA12
	ema12 := a.calculateEMA(candles, 12)

	// Расчет EMA26
	ema26 := a.calculateEMA(candles, 26)

	// Расчет MACD
	macd := ema12 - ema26

	// Расчет Signal (EMA9 от MACD)
	signal := a.calculateMACDSignal(candles, ema12, ema26)

	// Расчет Histogram
	histogram := macd - signal

	// Расчет RSI
	rsi := a.calculateRSI(candles, 14)

	indicators := models.TechnicalIndicators{
		SMA20:     sma20,
		SMA50:     sma50,
		EMA12:     ema12,
		EMA26:     ema26,
		RSI:       rsi,
		MACD:      macd,
		Signal:    signal,
		Histogram: histogram,
	}

	slog.Debug("Индикаторы рассчитаны",
		"sma20", sma20,
		"sma50", sma50,
		"ema12", ema12,
		"ema26", ema26,
		"rsi", rsi,
		"macd", macd)

	return indicators
}

// calculateEMA рассчитывает Exponential Moving Average
func (a *AnalysisService) calculateEMA(candles []models.Candle, period int) float64 {
	if len(candles) < period {
		return 0
	}

	// Начинаем с SMA
	var sum float64
	for i := 0; i < period; i++ {
		sum += candles[len(candles)-1-i].Close
	}
	ema := sum / float64(period)

	// Продолжаем с EMA
	multiplier := 2.0 / (float64(period) + 1)
	for i := period; i < len(candles); i++ {
		ema = (candles[len(candles)-1-i].Close-ema)*multiplier + ema
	}

	return ema
}

// calculateMACDSignal рассчитывает сигнальную линию MACD
func (a *AnalysisService) calculateMACDSignal(_ []models.Candle, ema12, ema26 float64) float64 {
	// Упрощенный расчет сигнальной линии
	// В реальности нужно рассчитывать EMA9 от значений MACD
	return (ema12 + ema26) / 2 * 0.9 // Упрощенная формула для примера
}

// calculateRSI рассчитывает Relative Strength Index
func (a *AnalysisService) calculateRSI(candles []models.Candle, period int) float64 {
	if len(candles) <= period {
		return 50.0 // Нейтральное значение при недостатке данных
	}

	gains := 0.0
	losses := 0.0

	for i := 1; i <= period; i++ {
		change := candles[len(candles)-i].Close - candles[len(candles)-i-1].Close
		if change > 0 {
			gains += change
		} else {
			losses -= change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	if avgLoss == 0 {
		return 100.0
	}

	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}

// asyncUpdatePairs асинхронно обновляет данные каждые 2 часа
func (a *AnalysisService) asyncUpdatePairs() {
	ticker := time.NewTicker(2 * time.Hour)
	defer ticker.Stop()

	slog.Info("Запуск асинхронного обновления данных",
		"interval", "2 hours")

	for range ticker.C {
		slog.Info("Начало планового обновления данных")

		startTime := time.Now()
		newPairs := a.uploadApi()

		a.mu.Lock()
		a.Pairs = newPairs
		a.mu.Unlock()

		duration := time.Since(startTime)

		slog.Info("Данные успешно обновлены",
			"duration", duration,
			"records", len(newPairs))
	}
}

// GetPairInfo возвращает информацию по паре и таймфрейму
func (a *AnalysisService) GetPairInfo(pair, timeframe string) (*models.AnalysisData, error) {
	slog.Debug("Поиск данных по паре",
		"pair", pair,
		"timeframe", timeframe)

	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, data := range a.Pairs {
		if data.Pair == pair && data.Timeframe == timeframe {
			slog.Debug("Данные найдены",
				"pair", pair,
				"timeframe", timeframe)
			return &data, nil
		}
	}

	slog.Warn("Данные не найдены",
		"pair", pair,
		"timeframe", timeframe)

	return nil, fmt.Errorf("данные для пары %s и таймфрейма %s не найдены", pair, timeframe)
}

// GetAllPairs возвращает все доступные пары
func (a *AnalysisService) GetAllPairs() models.PairsCrypto {
	a.mu.RLock()
	defer a.mu.RUnlock()

	slog.Debug("Запрос всех пар",
		"availablePairs", len(a.Pairs))

	return a.Pairs
}
