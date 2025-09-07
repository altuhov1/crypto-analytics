package main

import (
	"fmt"
	"net/http"
)

// Обработчик для главной страницы
func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		// Если это НЕ GET-запрос, возвращаем ошибку 405 "Method Not Allowed"
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprint(w, "Hello, World! (This was a GET request)\n")
}
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "This is a simple backend server!\n")
}

func main() {
	// Регистрируем наш обработчик для корневого URL "/"
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/about", aboutHandler) // Добавь эту строку

	// Запускаем веб-сервер на порту 8080
	fmt.Println("Сервер запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err) // В случае ошибки просто "паникуем"
	}
}
