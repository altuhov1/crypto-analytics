package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) GetPairInfo(w http.ResponseWriter, r *http.Request) {
	pair := r.URL.Query().Get("pair")
	timeframe := r.URL.Query().Get("timeframe")

	if pair == "" || timeframe == "" {
		http.Error(w, "Параметры pair и timeframe обязательны", http.StatusBadRequest)
		return
	}

	data, err := h.amalysis.GetPairInfo(pair, timeframe)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) GetAvailablePairs(w http.ResponseWriter, r *http.Request) {
	pairs := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
	timeframes := []string{"5m", "1h"}

	response := map[string]interface{}{
		"pairs":      pairs,
		"timeframes": timeframes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
