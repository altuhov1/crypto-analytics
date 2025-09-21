package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"webdev-90-days/internal/models"
	"webdev-90-days/internal/services"
	"webdev-90-days/internal/storage"
)

// Handler структурка, которая хранит зависимости (сервисы, хранилища)
// Это называется "Dependency Injection"
type Handler struct {
	storage  storage.Storage
	notifier services.Notifier
	tmpl     *template.Template
}

// NewHandler создает новый экземпляр Handler
func NewHandler(storage storage.Storage, notifier services.Notifier) (*Handler, error) {
	tmpl, err := template.ParseFiles(filepath.Join("static", "answer.html"))
	if err != nil {
		return nil, err
	}
	return &Handler{storage: storage, notifier: notifier, tmpl: tmpl}, nil
}

func (h *Handler) ContactFormHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Обработка POST запроса на /Contact")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("Ошибка парсинга формы: %v", err)
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	contact := models.ContactForm{
		Name:    r.FormValue("name"),
		Email:   r.FormValue("email"),
		Message: r.FormValue("message"),
	}
	fmt.Println(contact)

	// ВАЛИДАЦИЯ
	if contact.Name == "" || contact.Email == "" || contact.Message == "" {
		log.Printf("Невалидные данные: %+v", contact)
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// СОХРАНЕНИЕ
	if err := h.storage.SaveContact(&contact); err != nil {
		log.Printf("Ошибка сохранения: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// УВЕДОМЛЕНИЕ (асинхронно)
	go h.notifier.NotifyAdmin(&contact)

	// ОТВЕТ ПОЛЬЗОВАТЕЛЮ
	data := struct{ Name string }{Name: contact.Name}
	if err := h.tmpl.Execute(w, data); err != nil {
		log.Printf("Ошибка рендеринга шаблона: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
