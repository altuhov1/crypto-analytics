package services

import (
	"log/slog"
	"time"

	"crypto-analytics/internal/models"
)

type NotifierStruct struct{}

func NewNotifier() Notifier {
	return &NotifierStruct{}
}


func (n *NotifierStruct) NotifyAdmContForm(contact *models.ContactForm) {
	time.Sleep(100 * time.Millisecond)
	slog.Info("<-> УВЕДОМЛЕНИЕ ДЛЯ АДМИНА <->",
		"user_name", contact.Name,
		"user_email", contact.Email,
		"message", contact.Message,
	)
}

func (n *NotifierStruct) NotifyAdmNewUserForm(contact *models.User) {
	time.Sleep(100 * time.Millisecond) 
	slog.Info("<-> УВЕДОМЛЕНИЕ ДЛЯ АДМИНА <->",
		"username", contact.Username,
		"email", contact.Email,
		"event", "registration",
	)
}
