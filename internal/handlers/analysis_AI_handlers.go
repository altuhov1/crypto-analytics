package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// анализ ai
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

func (h *Handler) CryptoPairsPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/crypto_pairs.html")
}

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

	pairs, err := h.pairs.GetAllPairs()
	if err != nil {
		slog.Error("Failed to get pairs for validation", "error", err)
		json.NewEncoder(w).Encode(SelectPairResponse{
			Success: false,
			Error:   "Failed to validate pair",
		})
		return
	}

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
