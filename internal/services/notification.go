package services

import (
	"log"
	"time"

	"webdev-90-days/internal/models"
)

// Notifier сервис для отправки уведомлений
type NotifierStruct struct{}

// NewNotifier создает новый Notifier
func NewNotifier() Notifier {
	return &NotifierStruct{}
}

// NotifyAdmContForm уведомляет админа о новом сообщении
func (n *NotifierStruct) NotifyAdmContForm(contact *models.ContactForm) {
	time.Sleep(100 * time.Millisecond) // Имитация задержки
	log.Printf("=== УВЕДОМЛЕНИЕ ДЛЯ АДМИНА ===\n")
	log.Printf("Новое сообщение от: %s (%s)\n", contact.Name, contact.Email)
	log.Printf("Текст сообщения: %s\n", contact.Message)
	log.Printf("=== КОНЕЦ УВЕДОМЛЕНИЯ ===\n\n")
}

func (n *NotifierStruct) NotifyAdmNewUserForm(contact *models.User) {
	time.Sleep(100 * time.Millisecond) // Имитация задержки
	log.Printf("=== УВЕДОМЛЕНИЕ ДЛЯ АДМИНА ===\n")
	log.Printf("Регистрация : %s с почтой (%s)\n", contact.Username, contact.Email)
	log.Printf("=== КОНЕЦ УВЕДОМЛЕНИЯ ===\n\n")
}
