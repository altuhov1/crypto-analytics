// storage/storage.go
package storage

import "webdev-90-days/internal/models"

// Storage определяет контракт для работы с данными
type FormStorage interface {
	SaveContactFrom(contact *models.ContactForm) error
	// GetContact(id int) (*models.ContactForm, error) // На будущее
	// GetAllContacts() ([]*models.ContactForm, error) // На будущее
	Close() error // Для закрытия соединений
}

type UserStorage interface {
	CreateUser(user *models.User) error
	GetUserByName(nameU string) (*models.User, error)
	// UpdateUser(user *models.User) error
	// DeleteUser(id int) error
}
