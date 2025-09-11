package handlers

import (
	"fmt"
	"net/http"
	"text/template"
	"time"
)

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "This is a simple backend server!\n")
}
func FormHandler(w http.ResponseWriter, r *http.Request) {
	// Разрешаем только POST-запросы
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим форму, которая пришла в теле запроса.
	// Этот метод способен парсинговать формы с Content-Type: application/x-www-form-urlencoded
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Извлекаем данные из полей формы
	name := r.FormValue("name")
	email := r.FormValue("email")

	// Отвечаем пользователю
	fmt.Fprintf(w, "POST request successful!\nName: %s\nEmail: %s\n", name, email)
}
func notifyAdmin(form *ContactForm) {
	// Имитация задержки сети при обращении к внешнему сервису
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("=== УВЕДОМЛЕНИЕ ДЛЯ АДМИНА ===\n")
	fmt.Printf("Новое сообщение от: %s (%s)\n", form.Name, form.Email)
	fmt.Printf("Текст сообщения: %s\n", form.Message)
	fmt.Printf("=== КОНЕЦ УВЕДОМЛЕНИЯ ===\n\n")
}
func SubmitFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Создаем ЭКЗЕМПЛЯР структуры ContactForm и заполняем его данными из формы
	formData := ContactForm{
		Name:    r.FormValue("name"),
		Email:   r.FormValue("email"),
		Message: r.FormValue("message"),
	}

	// !!! ВАЖНОЕ НОВОВВЕДЕНИЕ: Сохраняем данные в файл
	err = saveToFile(&formData)
	if err != nil {
		// Если произошла ошибка при сохранении - логируем ее и показываем ошибку пользователю
		fmt.Printf("Error saving data: %v\n", err) // Вывод в консоль сервера для дебага
		http.Error(w, "Internal server error: could not save data", http.StatusInternalServerError)
		return
	}
	go notifyAdmin(&formData)

	tmpl, err := template.ParseFiles("static/answer.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Подготавливаем данные для шаблона
	data := AnswerData{
		Name: formData.Name,
	}

	// Выполняем шаблон с данными
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
