// storage/storage.go
package storage

import "webdev-90-days/internal/models"

// Storage определяет контракт для работы с данными
type Storage interface {
	SaveContact(contact *models.ContactForm) error
	// GetContact(id int) (*models.ContactForm, error) // На будущее
	// GetAllContacts() ([]*models.ContactForm, error) // На будущее
	Close() error // Для закрытия соединений
}
