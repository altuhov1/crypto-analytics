package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"webdev-90-days/internal/config"
	"webdev-90-days/internal/handlers"
	"webdev-90-days/internal/services"
	"webdev-90-days/internal/storage"
)

type App struct {
	cfg      *config.Config
	logger   *slog.Logger
	server   *http.Server
	services *Services
	storages *Storages
}

type Services struct {
	notifier services.Notifier
	crypto   *services.CryptoService
	news     *services.NewsService
	users    *services.UserService
	pairs    *services.CryptoPairsService
	analysis *services.AnalysisService
}

type Storages struct {
	contacts storage.FormStorage
	users    storage.UserStorage
	news     storage.NewsStorage
	pairs    storage.CacheStorage
	anslysis storage.AnalysisStorage
}

func main() {
	app := NewApp()
	app.Run()
}

func NewApp() *App {
	cfg := config.MustLoad()
	logger := config.NewLogger(cfg)
	slog.SetDefault(logger)

	app := &App{
		cfg:    cfg,
		logger: logger,
	}

	app.initStorages()
	app.initServices()
	app.initHTTP()

	return app
}

func (a *App) initStorages() {
	dbConfig := storage.PGXConfig{
		Host:     a.cfg.DBHost,
		Port:     a.cfg.DBPort,
		User:     a.cfg.DBUser,
		Password: a.cfg.DBPassword,
		DBName:   a.cfg.DBName,
		SSLMode:  a.cfg.DBSSLMode,
	}

	contactsStorage, err := storage.NewPGXStorage(dbConfig)
	if err != nil {
		slog.Error("Failed to initialize contacts storage", "error", err)
		os.Exit(1)
	}

	usersStorage, err := storage.NewUserPostgresStorage(dbConfig)
	if err != nil {
		slog.Error("Failed to initialize users storage", "error", err)
		os.Exit(1)
	}

	newsStorage := storage.NewNewsFileStorage("storage/news_cache.json")

	pairsStorage := storage.NewPairsFileStorage("storage/pairs_cache.json")

	analysisStorage := storage.NewAnalysisFileStorage("storage/analysis_cache.json")

	a.storages = &Storages{
		contacts: contactsStorage,
		users:    usersStorage,
		news:     newsStorage,
		pairs:    pairsStorage,
		anslysis: analysisStorage,
	}
}

func (a *App) initServices() {
	a.services = &Services{
		notifier: services.NewNotifier(),
		crypto:   services.NewCryptoService(false, "storage/crypto_cache.json"),
		news:     services.NewNewsService(a.storages.news, false),
		users:    services.NewUserService(a.storages.users),
		pairs:    services.NewCryptoPairsService(a.storages.pairs, false),
		analysis: services.NewAnalysisService(false, a.storages.anslysis),
	}
}

func (a *App) initHTTP() {
	handler, err := handlers.NewHandler(
		a.storages.contacts,
		a.services.notifier,
		a.services.crypto,
		a.services.users,
		a.cfg.KeyUsersGorilla,
		a.services.news,
		a.services.pairs,
		a.services.analysis,
	)
	if err != nil {
		slog.Error("Failed to create handler", "error", err)
		os.Exit(1)
	}

	router := a.setupRoutes(handler)

	a.server = &http.Server{
		Addr:         ":" + a.cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
}

func (a *App) setupRoutes(handler *handlers.Handler) http.Handler {
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// API routes
	apiRoutes := map[string]http.HandlerFunc{
		"/api/allFavoriteCoin":    handler.GetFavorites,
		"/api/changeFavoriteCoin": handler.ChangeFavorite,
		"/api/printUserstInfo":    handler.InfoOfUsers,
		"/api/printContactInfo":   handler.InfoOfContacts,
		"/api/cache-info":         handler.CacheInfoHandler,
		"/api/all-pairs":          handler.GetAllPairsHandler,
		"/api/select-pair":        handler.SelectPairHandler,
		"/api/pair":               handler.GetPairInfo,
		"/api/pairs":              handler.GetAllPairs,
		"/api/available":          handler.GetAvailablePairs,
	}

	for path, handlerFunc := range apiRoutes {
		mux.HandleFunc(path, handlerFunc)
	}

	// Web routes
	webRoutes := map[string]http.HandlerFunc{
		"/news":          handler.NewsPage,
		"/pairs":         handler.CryptoPairsPageHandler,
		"/logout":        handler.LogoutHandler,
		"/login":         handler.LoginHandler,
		"/check-Sess-Id": handler.CheckAuthHandler,
		"/register":      handler.AuthUserFormHandler,
		"/contact":       handler.ContactFormHandler,
		"/crypto-top":    handler.CryptoTopHandler,
	}

	for path, handlerFunc := range webRoutes {
		mux.HandleFunc(path, handlerFunc)
	}

	// Root route
	mux.HandleFunc("/", a.rootHandler)

	return mux
}

func (a *App) rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "static/index.html")
}

func (a *App) Run() {
	go a.startServer()
	a.waitForShutdown()
}

func (a *App) startServer() {
	slog.Info("Server starting", "port", a.server.Addr)
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func (a *App) waitForShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Shutting down server gracefully...")
	a.shutdown()
}

func (a *App) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	a.storages.contacts.Close()
	a.storages.users.Close()
	slog.Info("Server stopped")
	if a.cfg.LaunchLoc == "prod" {
		time.Sleep(1 * time.Second)
	}
}
