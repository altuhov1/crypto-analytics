package services

import (
	"log/slog"
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
	slog.Info("=== УВЕДОМЛЕНИЕ ДЛЯ АДМИНА ===",
		"user_name", contact.Name,
		"user_email", contact.Email,
		"message", contact.Message,
	)
}

func (n *NotifierStruct) NotifyAdmNewUserForm(contact *models.User) {
	time.Sleep(100 * time.Millisecond) // Имитация задержки
	slog.Info("=== УВЕДОМЛЕНИЕ ДЛЯ АДМИНА ===",
		"username", contact.Username,
		"email", contact.Email,
		"event", "registration",
	)
}
