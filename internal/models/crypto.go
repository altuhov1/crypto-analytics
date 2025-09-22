package models

// Coin представляет данные о криптовалюте от CoinGecko API
type Coin struct {
	ID            string  `json:"id"`
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	CurrentPrice  float64 `json:"current_price"`
	PriceChange24 float64 `json:"price_change_percentage_24h"`
	MarketCap     float64 `json:"market_cap"`   // Измените на float64
	Volume24h     float64 `json:"total_volume"` // Измените на float64
	LastUpdated   string  `json:"last_updated"`
}

// CoinGeckoResponse представляет ответ от CoinGecko API
type CoinGeckoResponse []Coin
