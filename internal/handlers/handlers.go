package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"crypto-analytics/internal/models"
	"crypto-analytics/internal/services"
	"crypto-analytics/internal/storage"

	"github.com/gorilla/sessions"
)

type Handler struct {
	storage       storage.FormStorage
	notifier      services.Notifier
	cryptoSvc     *services.CryptoService
	userService   *services.UserService
	tmpl          *template.Template
	storeSessions *sessions.CookieStore
	newsStorage   *services.NewsService
	pairs         *services.CryptoPairsService
	amalysis      *services.AnalysisService
}

// NewHandler создает новый экземпляр Handler
func NewHandler(storage storage.FormStorage,
	notifier services.Notifier,
	cryptoSvc *services.CryptoService,
	userService *services.UserService,
	KeyUsersGorilla string,
	newsStor *services.NewsService,
	pairss *services.CryptoPairsService,
	amalys *services.AnalysisService) (*Handler, error) {

	tmpl := template.New("").Funcs(template.FuncMap{
		"formatNumber": formatNumber,
		"add":          add,
		"formatMoney":  formatMoney,
		"parseTime":    parseTime,
		"stripHTML": func(html string) string {
			// Простая очистка от HTML тегов
			re := regexp.MustCompile(`<[^>]*>`)
			return re.ReplaceAllString(html, "")
		},
	})
	tmpl, err := tmpl.ParseFiles(
		filepath.Join("static", "answerForm.html"),
		filepath.Join("static", "crypto_top.html"),
		filepath.Join("static", "news.html"),
	)
	if err != nil {
		return nil, err
	}
	return &Handler{
		storage:     storage,
		notifier:    notifier,
		cryptoSvc:   cryptoSvc,
		userService: userService,
		tmpl:        tmpl,
		storeSessions: sessions.NewCookieStore(
			[]byte(KeyUsersGorilla)),
		newsStorage: newsStor,
		pairs:       pairss,
		amalysis:    amalys,
	}, nil
}

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

func (h *Handler) CryptoTopHandler(w http.ResponseWriter, r *http.Request) {

	// Получаем параметр limit из query string (по умолчанию 10)
	limitStr := r.URL.Query().Get("limit")
	limit := 250
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Получаем данные из CoinGecko
	coins, err := h.cryptoSvc.GetTopCryptos(limit)
	if err != nil {
		slog.Warn("Ошибка получения криптовалюты:", "error", err)
		http.Error(w, "Временные проблемы с получением данных", http.StatusInternalServerError)
		return
	}

	// Рендерим шаблон
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "crypto_top.html", coins); err != nil {
		slog.Warn("Ошибка рендеринга шаблона:", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) CacheInfoHandler(w http.ResponseWriter, r *http.Request) {
	count, cacheTime := h.cryptoSvc.GetCacheInfo()

	info := map[string]interface{}{
		"cached_coins": count,
		"last_updated": cacheTime.Format("2006-01-02 15:04:05"),
		"age_minutes":  time.Since(cacheTime).Minutes(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (h *Handler) InfoOfContacts(w http.ResponseWriter, r *http.Request) {
	h.storage.ExportContactsToJSON("storage/info_contacts.json")
	fmt.Fprint(w, "OK")
}
