package handlers

import (
	"html/template"
	"path/filepath"
	"regexp"

	"crypto-analytics/internal/services"
	"crypto-analytics/internal/storage"

	"github.com/gorilla/sessions"
)

type Handler struct {
	storage       storage.FormStorage
	notifier      services.Notifier
	cryptoSvc     services.GetAllPairsService
	userService   services.UserLogService
	tmpl          *template.Template
	storeSessions *sessions.CookieStore
	newsStorage   services.NewsRssService
	pairs         services.AIAnalysisService
	Analysis      services.AnalysisGService
	postsService  services.PostPService
}

func NewHandler(storage storage.FormStorage,
	notifier services.Notifier,
	cryptoSvc services.GetAllPairsService,
	userService services.UserLogService,
	KeyUsersGorilla string,
	newsStor services.NewsRssService,
	pairss services.AIAnalysisService,
	analys services.AnalysisGService,
	post services.PostPService) (*Handler, error) {

	tmpl := template.New("").Funcs(template.FuncMap{
		"formatNumber": formatNumber,
		"add":          add,
		"formatMoney":  formatMoney,
		"parseTime":    parseTime,
		"stripHTML": func(html string) string {
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
		newsStorage:  newsStor,
		pairs:        pairss,
		Analysis:     analys,
		postsService: post,
	}, nil
}
