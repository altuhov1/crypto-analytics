package models

type PairsCrypto []AnalysisData

type AnalysisData struct {
	Pair       string              `json:"pair"`
	Timeframe  string              `json:"timeframe"`
	Candles    []Candle            `json:"candles"`
	Indicators TechnicalIndicators `json:"indicators"`
	Timestamp  int64               `json:"timestamp"`
}

type Candle struct {
	OpenTime  int64   `json:"openTime"`
	CloseTime int64   `json:"closeTime"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

type TechnicalIndicators struct {
	SMA20     float64 `json:"sma20"`
	SMA50     float64 `json:"sma50"`
	EMA12     float64 `json:"ema12"`
	EMA26     float64 `json:"ema26"`
	RSI       float64 `json:"rsi"`
	MACD      float64 `json:"macd"`
	Signal    float64 `json:"signal"`
	Histogram float64 `json:"histogram"`
}
