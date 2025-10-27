package handlers

import (
	"crypto-analytics/internal/models"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func (h *Handler) AuthUserFormHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		slog.Warn("Ошибка парсинга формы:", "error", err)
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	contact := &models.User{
		Username:      r.FormValue("username"),
		Email:         r.FormValue("email"),
		Password:      r.FormValue("password"),
		FavoriteCoins: make([]string, 0),
	}

	if contact.Username == "" || contact.Email == "" || contact.Password == "" {
		slog.Warn("Невалидные данные:", "contact", contact)
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	err := h.userService.RegisterUser(contact)
	if err != nil {

		errText := err.Error()
		switch {
		case strings.Contains(errText, "user already exists"):
			http.Redirect(w, r, "/static/FormNewUser.html?err=alreadyExistsName", http.StatusSeeOther)
			return
		case strings.Contains(errText, "user name already exists"):
			http.Redirect(w, r, "/static/FormNewUser.html?err=alreadyExistsAcc", http.StatusSeeOther)
			return
		default:
			slog.Warn("Ошибка сохранения", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
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

	if auth, ok := session.Values["loggedIn"].(bool); ok && auth {
		username := session.Values["username"].(string)
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

	username := r.FormValue("username")
	password := r.FormValue("password")

	err := h.userService.LoginUser(username, password)
	if err != nil {
		http.Redirect(w, r, "/static/FormRegUser.html?err=password", http.StatusSeeOther)
		return
	}

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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, _ := h.storeSessions.Get(r, "user-session")
	// Полностью удаляем куку
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) ChangeFavorite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username, authenticated := h.getCurrentUser(r)
	if !authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	// Парсим JSON запрос
	var req models.FavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Action == "add" {
		err := h.userService.AddFavorite(username, req.CoinID)
		if err != nil {
			slog.Warn("Ошибка", "error", err)
			http.Error(w, "Cant create", http.StatusBadRequest)
			return
		}
	} else {
		err := h.userService.RemoveFavorite(username, req.CoinID)
		if err != nil {
			slog.Warn("Ошибка", "error", err)
			http.Error(w, "Cant create", http.StatusBadRequest)
			return
		}

	}

	response := map[string]interface{}{
		"success": true,
		"message": "Favorite updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetFavorites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username, authenticated := h.getCurrentUser(r)
	if !authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	favorites, err := h.userService.GetFavorites(username)
	if err != nil {
		slog.Warn("Ошибка", "error", err)
		http.Error(w, "Cant create", http.StatusBadRequest)
	}

	response := APIResponse{
		Success: true,
		Message: "Favorites retrieved",
		Data:    favorites,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Вспомогательный метод для получения текущего пользователя
func (h *Handler) getCurrentUser(r *http.Request) (string, bool) {
	session, err := h.storeSessions.Get(r, "user-session")
	if err != nil {
		return "", false
	}

	if auth, ok := session.Values["loggedIn"].(bool); !ok || !auth {
		return "", false
	}

	username, ok := session.Values["username"].(string)
	return username, ok
}

func (h *Handler) InfoOfUsers(w http.ResponseWriter, r *http.Request) {
	h.userService.PrintJsonAllUsers("storage/user.json")
	fmt.Fprint(w, "OK")
}
