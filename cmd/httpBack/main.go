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

func main() {

	cfg := config.MustLoad()

	logger := config.NewLogger(cfg)
	slog.SetDefault(logger)

	configDB := storage.PGXConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	}

	pgStorageContacts, err := storage.NewPGXStorage(configDB)
	if err != nil {
		slog.Error("Ошибка подключения", "error", err)
		os.Exit(1)
	}
	defer pgStorageContacts.Close()

	pgSorageUser, err := storage.NewUserPostgresStorage(configDB)
	if err != nil {
		panic("can not connect to db's table of users")
	} else {
		defer pgSorageUser.Close()
	}

	notifier := services.NewNotifier()
	cryptoSvc := services.NewCryptoService(false, "storage/crypto_cache.json")

	newsStorage := storage.NewNewsFileStorage("storage/news_cache.json")

	newsService := services.NewNewsService(newsStorage, false)

	userSvc := services.NewUserService(pgSorageUser)
	handler, err := handlers.NewHandler(pgStorageContacts, notifier, cryptoSvc, userSvc, cfg.KeyUsersGorilla, newsService)
	if err != nil {
		slog.Error("Failed to create handler:", "error", err)
		os.Exit(1)
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/news", handler.NewsPage)
	http.HandleFunc("/logout", handler.LogoutHandler)
	http.HandleFunc("/login", handler.LoginHandler)
	http.HandleFunc("/check-Sess-Id", handler.CheckAuthHandler)
	http.HandleFunc("/register", handler.AuthUserFormHandler)
	http.HandleFunc("/contact", handler.ContactFormHandler)
	http.HandleFunc("/crypto-top", handler.CryptoTopHandler)

	http.HandleFunc("/api/allFavoriteCoin", handler.GetFavorites)
	http.HandleFunc("/api/changeFavoriteCoin", handler.ChangeFavorite)
	http.HandleFunc("/api/printUserstInfo", handler.InfoOfUsers)
	http.HandleFunc("/api/printContactInfo", handler.InfoOfContacts)
	http.HandleFunc("/api/cache-info", handler.CacheInfoHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r) //404
			return
		}
		http.ServeFile(w, r, "static/index.html")
	})

	// Создаем HTTP сервер с таймаутами
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		slog.Info("Server starting", "port", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	slog.Info("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Останавливаем tcp подключения
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown:", "error", err)
		os.Exit(1)
	}
	slog.Info("Server stopped")
}
