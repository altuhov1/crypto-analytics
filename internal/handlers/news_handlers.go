package handlers

import (
	"crypto-analytics/internal/models"
	"net/http"
)

func (h *Handler) NewsPage(w http.ResponseWriter, r *http.Request) {
	news, err := h.newsStorage.GetNews()
	if err != nil {
		http.Error(w, "Failed to load news: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		News []models.NewsItem
	}{
		News: news,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "news.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
