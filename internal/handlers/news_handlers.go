package handlers

import (
	"net/http"
	"webdev-90-days/internal/models"
)

func (h *Handler) NewsPage(w http.ResponseWriter, r *http.Request) {
	// Получаем уже отсортированные новости из сервиса
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
