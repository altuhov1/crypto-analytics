package services

import "webdev-90-days/internal/models"

// Notifier определяет контракт для отправки уведомлений.
type Notifier interface {
	NotifyAdmin(contact *models.ContactForm)
}
