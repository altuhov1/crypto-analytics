package models

// Coin представляет данные о криптовалюте от CoinGecko API
type Coin struct {
	ID            string  `json:"id"`
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	CurrentPrice  float64 `json:"current_price"`
	PriceChange24 float64 `json:"price_change_percentage_24h"`
	MarketCap     int64   `json:"market_cap"`
	Volume24h     int64   `json:"total_volume"`
	LastUpdated   string  `json:"last_updated"`
}

// CoinGeckoResponse представляет ответ от CoinGecko API
type CoinGeckoResponse []Coin
