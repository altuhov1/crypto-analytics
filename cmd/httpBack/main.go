package main

import (
	"context"
	"log"
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
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Загружаем конфигурацию
	cfg := config.MustLoad()

	// Инициализация хранилища
	configDB := storage.PGXConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	}

	pgStorage, err := storage.NewPGXStorage(configDB)
	if err != nil {
		log.Fatal("Ошибка подключения:", err)
	}
	defer pgStorage.Close()

	notifier := services.NewNotifier()
	handler, err := handlers.NewHandler(pgStorage, notifier)
	if err != nil {
		log.Fatal("Failed to create handler:", err)
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/contact", handler.ContactFormHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Если запрос не к корню - отдаем 404
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "static/index.html")
	})

	// Создаем HTTP сервер с таймаутами
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Запускаем сервер в горутине
	go func() {
		log.Printf("Server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Канал для перехвата сигналов ОС
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Ждем сигнал о завершении
	<-stop
	log.Println("Shutting down server gracefully...")

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Останавливаем сервер
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
