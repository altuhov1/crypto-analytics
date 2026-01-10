package handlers

import (
	"crypto-analytics/internal/models"
	"log/slog"
	"net/http"
)

func (h *Handler) ContactFormHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		slog.Warn("Ошибка парсинга формы:", "error", err)
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	contact := models.ContactForm{
		Name:    r.FormValue("name"),
		Email:   r.FormValue("email"),
		Message: r.FormValue("message"),
	}

	if contact.Name == "" || contact.Email == "" || contact.Message == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	if err := h.storage.SaveContactFrom(&contact); err != nil {
		slog.Warn("Ошибка сохранения:", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// УВЕДОМЛЕНИЕ (асинхронно)
	go h.notifier.NotifyAdmContForm(&contact)

	data := struct{ Name string }{Name: contact.Name}
	if err := h.tmpl.ExecuteTemplate(w, "answerForm.html", data); err != nil {
		slog.Warn("Ошибка рендаренга", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
