package main

import (
	"log"
	"net/http"

	"webdev-90-days/internal/config"
	"webdev-90-days/internal/handlers"
	"webdev-90-days/internal/services"
	"webdev-90-days/internal/storage"
)

func main() {
	// Настройка логирования
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Загружаем конфигурацию
	cfg := config.MustLoad()

	// Теперь cfg.StoragePath содержит путь из переменной окружения, а не "зашитый" в коде.

	// Инициализация хранилища ПЕРЕДАЕМ ПУТЬ ИЗ КОНФИГА
	fileStorage, err := storage.NewFileStorage(cfg.StoragePath)
	if err != nil {
		log.Fatal("Failed to create storage:", err)
	}

	notifier := services.NewNotifier()
	handler, err := handlers.NewHandler(fileStorage, notifier)
	if err != nil {
		log.Fatal("Failed to create handler:", err)
	}

	// Настройка маршрутов
	http.HandleFunc("/about", handler.AboutHandler)
	http.HandleFunc("/submit-form", handler.SubmitFormHandler)

	// Отдача статических файлов
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	log.Printf("Server starting on :%s...", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, nil))
}
