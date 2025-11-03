package services

import (
	"crypto-analytics/internal/models"
	"crypto-analytics/internal/storage"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type AnalysisService struct {
	store      storage.AnalysisStorage
	tempStore  storage.AnalysisTempStorage
	binanceAPI *BinanceAPI
	goToApi    bool
	mu         sync.RWMutex
}

func NewAnalysisService(goToApi bool, store storage.AnalysisStorage, tempS storage.AnalysisTempStorage) *AnalysisService {
	service := &AnalysisService{
		goToApi:    goToApi,
		store:      store,
		binanceAPI: NewBinanceAPI(),
		mu:         sync.RWMutex{},
		tempStore:  tempS,
	}


	if goToApi {
		slog.Info("Загрузка данных из API Binance")
		service.uploadApi()
		go service.asyncUpdatePairs()
	} else {
		err := service.tempStore.SavePairs(service.uploadFromStorage())
		if err != nil {
			slog.Error("err in uploadApi()", "err", err)
		}
	}

	slog.Info("Сервис анализа успешно инициализирован",
		"goToApi", goToApi,
		"loadedPairs", service.tempStore.GetStats())

	return service
}

func (a *AnalysisService) uploadApi() {
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
			err := a.tempStore.SaveAnalysisData(analysisData)
			if err != nil {
				slog.Error("err in uploadApi()", "err", err)
			}

			slog.Debug("Данные загружены",
				"pair", p,
				"timeframe", t,
				"candlesCount", len(candlesApi))
		}
	}

	if err := a.store.SaveAnalysisData(base); err != nil {
		slog.Error("Ошибка при сохранении данных в хранилище",
			"error", err)
	}

}

func (a *AnalysisService) uploadFromStorage() models.PairsCrypto {

	data, err := a.store.LoadAnalysisData()
	if err != nil {
		slog.Error("Ошибка при загрузке данных из хранилища",
			"error", err)
		return models.PairsCrypto{}
	}
	return data
}

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

func (a *AnalysisService) calculateMACDSignal(_ []models.Candle, ema12, ema26 float64) float64 {

	return (ema12 + ema26) / 2 * 0.9 
}

func (a *AnalysisService) calculateRSI(candles []models.Candle, period int) float64 {
	if len(candles) <= period {
		return 50.0 
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

func (a *AnalysisService) asyncUpdatePairs() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {

		startTime := time.Now()
		a.uploadApi()

		duration := time.Since(startTime)

		slog.Info("Данные для анализа пар с usdt успешно обновлены",
			"duration", duration,
			"records", a.tempStore.GetStats())
	}
}

func (a *AnalysisService) GetPairInfo(pair, timeframe string) (*models.AnalysisData, error) {
	slog.Debug("Поиск данных по паре",
		"pair", pair,
		"timeframe", timeframe)
	res, err := a.tempStore.GetAnalysisData(pair, timeframe)
	if err == nil {
		fmt.Println(res)
		return res, err
	}
	slog.Warn("Данные не найдены",
		"pair", pair,
		"timeframe", timeframe,
		"err", err,
	)

	return nil, fmt.Errorf("данные для пары %s и таймфрейма %s не найдены", pair, timeframe)
}
