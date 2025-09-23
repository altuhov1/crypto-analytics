package handlers

import (
	"fmt"
	"log"
	"net/http"
	"webdev-90-days/internal/models"
)

func (h *Handler) AuthUserFormHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Обработка POST запроса на /register")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("Ошибка парсинга формы: %v", err)
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	contact := &models.User{
		Username: r.FormValue("username"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}
	fmt.Println(contact)

	// ВАЛИДАЦИЯ
	if contact.Username == "" || contact.Email == "" || contact.Password == "" {
		log.Printf("Невалидные данные: %+v", contact)
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// СОХРАНЕНИЕ
	if err := h.userService.RegisterUser(contact); err != nil {
		log.Printf("Ошибка сохранения: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// УВЕДОМЛЕНИЕ (асинхронно)
	go h.notifier.NotifyAdmNewUserForm(contact)

	// ОТВЕТ ПОЛЬЗОВАТЕЛЮ
	// data := struct{ Name string }{Name: contact.Name}
	// if err := h.tmpl.ExecuteTemplate(w, "answerForm.html", data); err != nil {
	// 	log.Printf("Ошибка рендеринга шаблона: %v", err)
	// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
	// }
}
