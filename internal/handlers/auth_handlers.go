package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"webdev-90-days/internal/models"

	"github.com/gorilla/sessions"
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

	err := h.userService.RegisterUser(contact)
	if err != nil {
		fmt.Println("------", err, "------")

		errText := err.Error()
		switch {
		case strings.Contains(errText, "user already exists"):
			http.Redirect(w, r, "/static/FormNewUser.html?err=alreadyExistsName", http.StatusSeeOther)
			return
		case strings.Contains(errText, "user name already exists"):
			http.Redirect(w, r, "/static/FormNewUser.html?err=alreadyExistsAcc", http.StatusSeeOther)
			return
		default:
			log.Printf("Ошибка сохранения: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	// УВЕДОМЛЕНИЕ (асинхронно)
	go h.notifier.NotifyAdmNewUserForm(contact)

	http.Redirect(w, r, "/static/FormRegUser.html", http.StatusSeeOther)
}

func (h *Handler) CheckAuthHandler(w http.ResponseWriter, r *http.Request) {
	session, err := h.storeSessions.Get(r, "user-session")
	if err != nil {
		// Если ошибка - считаем что не авторизован
		json.NewEncoder(w).Encode(map[string]interface{}{"authenticated": false})
		return
	}

	// 2. Проверяем, вошел ли пользователь
	if auth, ok := session.Values["loggedIn"].(bool); ok && auth {
		// 3. Если вошел - возвращаем имя пользователя
		username := session.Values["username"].(string)
		response := map[string]interface{}{
			"authenticated": true,
			"username":      username,
		}
		json.NewEncoder(w).Encode(response)
	} else {
		// 4. Если не вошел
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

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Ваша проверка логина/пароля (оставьте как есть)
	err := h.userService.LoginUser(username, password)
	if err != nil {
		http.Redirect(w, r, "/static/FormRegUser.html?err=password", http.StatusSeeOther)
		return
	}

	// ПРОСТАЯ ЧАСТЬ: создаем сессию
	session, _ := h.storeSessions.Get(r, "user-session")
	session.Values["loggedIn"] = true
	session.Values["username"] = username
	session.Options = &sessions.Options{
		HttpOnly: true,  // Защита от XSS
		MaxAge:   86400, // Сессия на 1 день
		Path:     "/",   // Действует для всего сайта
	}
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := h.storeSessions.Get(r, "user-session")
	session.Values["loggedIn"] = false
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
