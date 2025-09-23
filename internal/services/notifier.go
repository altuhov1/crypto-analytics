package services

import "webdev-90-days/internal/models"

// Notifier определяет контракт для отправки уведомлений.
type Notifier interface {
	NotifyAdmContForm(contact *models.ContactForm)
	NotifyAdmNewUserForm(contact *models.User)
}
