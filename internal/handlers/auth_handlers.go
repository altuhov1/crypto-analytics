package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
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

	http.Redirect(w, r, "/static/FormRegUser.html", http.StatusSeeOther)
}

func (h *Handler) CheckAuthHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем sessionID из куков
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, `{"authenticated": false}`, http.StatusOK)
		return
	}

	// Получаем IP и браузер
	ip := strings.Split(r.RemoteAddr, ":")[0]
	browser := r.UserAgent()

	// Проверяем сессию
	username, valid := h.userService.ValidateSession(cookie.Value, ip, browser)

	if valid {
		response := map[string]interface{}{
			"authenticated": true,
			"username":      username,
		}
		json.NewEncoder(w).Encode(response)
	} else {
		response := map[string]interface{}{
			"authenticated": false,
		}
		json.NewEncoder(w).Encode(response)
	}
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Получаем данные из формы
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Проверяем что данные не пустые
	if username == "" || password == "" {
		http.Error(w, "Username and password required", http.StatusBadRequest)
		return
	}
	var ErrUserNotFound = errors.New("we have not this acc")
	err := h.userService.LoginUser(username, password)
	if errors.Is(err, ErrUserNotFound) {
		// Обработка случая "пользователь не найден"
	} else if err != nil {
		http.Redirect(w, r, "/static/FormRegUser.html?err=password", http.StatusSeeOther)
		return
		//обработка случая "пользователь найден, но"
	}

	// Создаем сессию
	ip := strings.Split(r.RemoteAddr, ":")[0]
	browser := r.UserAgent()
	sessionID := h.userService.CreateSession(username, ip, browser)

	// Устанавливаем куку
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   24 * 60 * 60,
		HttpOnly: true,
		Secure:   false, // true для HTTPS
	})

	// Перенаправляем на главную страницу
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Если нужно удалить сессию из БД
	if cookie, err := r.Cookie("session_id"); err == nil {
		h.userService.DeleteSession(cookie.Value) // удаляем из БД
	}

	// Удаляем куку из браузера
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
