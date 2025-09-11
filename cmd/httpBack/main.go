package main

import (
	"fmt"
	"net/http"
	"webdev-90-days/internal/handlers"
)

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/submit-form", handlers.SubmitFormHandler)
	http.HandleFunc("/about", handlers.AboutHandler) // Добавь эту строку
	http.HandleFunc("/form", handlers.FormHandler)
	// Запускаем веб-сервер на порту 8080
	fmt.Println("Сервер запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err) // В случае ошибки просто "паникуем"
	}
}
