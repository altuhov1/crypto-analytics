package services

import "crypto-analytics/internal/models"

type Notifier interface {
	NotifyAdmContForm(contact *models.ContactForm)
	NotifyAdmNewUserForm(contact *models.User)
}
