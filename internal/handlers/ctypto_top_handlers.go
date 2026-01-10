package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
)

func (h *Handler) CryptoTopHandler(w http.ResponseWriter, r *http.Request) {

	// Получаем параметр limit из query string (по умолчанию 10)
	limitStr := r.URL.Query().Get("limit")
	limit := 250
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Получаем данные из CoinGecko
	coins, err := h.cryptoSvc.GetTopCryptos(limit)
	if err != nil {
		slog.Warn("Ошибка получения криптовалюты:", "error", err)
		http.Error(w, "Временные проблемы с получением данных", http.StatusInternalServerError)
		return
	}

	// Рендерим шаблон
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "crypto_top.html", coins); err != nil {
		slog.Warn("Ошибка рендеринга шаблона:", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
