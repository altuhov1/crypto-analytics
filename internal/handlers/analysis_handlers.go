package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type PairsDataResponse struct {
	Success bool     `json:"success"`
	Pairs   []string `json:"pairs,omitempty"`
	Error   string   `json:"error,omitempty"`
}

type SelectPairRequest struct {
	Pair string `json:"pair"`
}

type SelectPairResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// CryptoPairsPageHandler отображает страницу выбора пар
func (h *Handler) CryptoPairsPageHandler(w http.ResponseWriter, r *http.Request) {
	// pairs, err := h.pairs.GetAllPairs()
	// if err != nil {
	// 	slog.Error("Failed to get pairs for page", "error", err)
	// 	http.Error(w, "Failed to load crypto pairs", http.StatusInternalServerError)
	// 	return
	// }

	// Здесь можно использовать шаблонизатор или просто отдать HTML
	// Для простоты будем отдавать готовый HTML с данными
	http.ServeFile(w, r, "static/crypto_pairs.html")
}

// GetAllPairsHandler возвращает все пары для клиентского поиска
func (h *Handler) GetAllPairsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pairs, err := h.pairs.GetAllPairs()
	if err != nil {
		slog.Error("Failed to get pairs", "error", err)
		json.NewEncoder(w).Encode(PairsDataResponse{
			Success: false,
			Error:   "Failed to load crypto pairs",
		})
		return
	}

	json.NewEncoder(w).Encode(PairsDataResponse{
		Success: true,
		Pairs:   pairs,
	})
}

// SelectPairHandler отправляет выбранную пару на внешний бэкенд
func (h *Handler) SelectPairHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SelectPairRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode request", "error", err)
		json.NewEncoder(w).Encode(SelectPairResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	if req.Pair == "" {
		json.NewEncoder(w).Encode(SelectPairResponse{
			Success: false,
			Error:   "Pair is required",
		})
		return
	}

	// Валидация пары
	pairs, err := h.pairs.GetAllPairs()
	if err != nil {
		slog.Error("Failed to get pairs for validation", "error", err)
		json.NewEncoder(w).Encode(SelectPairResponse{
			Success: false,
			Error:   "Failed to validate pair",
		})
		return
	}

	// Проверяем, существует ли такая пара
	validPair := false
	for _, pair := range pairs {
		if pair == req.Pair {
			validPair = true
			break
		}
	}

	if !validPair {
		json.NewEncoder(w).Encode(SelectPairResponse{
			Success: false,
			Error:   "Invalid crypto pair",
		})
		return
	}

	// TODO: Здесь будет отправка на внешний бэкенд
	slog.Info("Pair selected for external backend", "pair", req.Pair)

	// Временная заглушка
	if err := h.sendToExternalBackend(req.Pair); err != nil {
		slog.Error("Failed to send to external backend", "error", err, "pair", req.Pair)
		json.NewEncoder(w).Encode(SelectPairResponse{
			Success: false,
			Error:   "Failed to send to external service",
		})
		return
	}

	json.NewEncoder(w).Encode(SelectPairResponse{
		Success: true,
		Message: fmt.Sprintf("Pair %s successfully sent for processing", req.Pair),
	})
}

// Временная заглушка для отправки на внешний бэкенд
func (h *Handler) sendToExternalBackend(pair string) error {
	// TODO: Реализовать отправку на ваш внешний бэкенд
	slog.Info("Sending to external backend (stub)", "pair", pair)
	return nil
}

// // //sdfasfasfkasl;fkjadfkasd;dfka;ldfka;'
// type AnalysisRequest struct {
// 	Pair      string `json:"pair"`
// 	Timeframe string `json:"timeframe"`
// 	UseCache  bool   `json:"useCache"`
// }

// type AnalysisResponse struct {
// 	Success bool                   `json:"success"`
// 	Data    *services.AnalysisData `json:"data,omitempty"`
// 	Error   string                 `json:"error,omitempty"`
// }

// func NewAnalysisHandler(analysisService *services.AnalysisService) *AnalysisHandler {
// 	return &AnalysisHandler{
// 		analysisService: analysisService,
// 	}
// }

// // AnalysisPageHandler отображает страницу анализа
// func (h *AnalysisHandler) AnalysisPageHandler(w http.ResponseWriter, r *http.Request) {
// 	pair := r.URL.Query().Get("pair")
// 	if pair == "" {
// 		pair = "BTCUSDT" // значение по умолчанию
// 	}

// 	// Можно передать пару в шаблон или использовать JavaScript
// 	http.ServeFile(w, r, "templates/analysis.html")
// }

// // GetAnalysisDataHandler API для получения данных анализа
// func (h *Handler) GetAnalysisDataHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	var req AnalysisRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		slog.Error("Failed to decode analysis request", "error", err)
// 		json.NewEncoder(w).Encode(AnalysisResponse{
// 			Success: false,
// 			Error:   "Invalid request format",
// 		})
// 		return
// 	}

// 	if req.Pair == "" {
// 		req.Pair = "BTCUSDT"
// 	}
// 	if req.Timeframe == "" {
// 		req.Timeframe = "1h"
// 	}

// 	slog.Info("Analysis request", "pair", req.Pair, "timeframe", req.Timeframe, "useCache", req.UseCache)

// 	analysisData, err := h.analysisService.GetAnalysisData(req.Pair, req.Timeframe, req.UseCache)
// 	if err != nil {
// 		slog.Error("Failed to get analysis data", "error", err, "pair", req.Pair)
// 		json.NewEncoder(w).Encode(AnalysisResponse{
// 			Success: false,
// 			Error:   fmt.Sprintf("Failed to load analysis data: %v", err),
// 		})
// 		return
// 	}

// 	json.NewEncoder(w).Encode(AnalysisResponse{
// 		Success: true,
// 		Data:    analysisData,
// 	})
// }
