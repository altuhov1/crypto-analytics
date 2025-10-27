// storage/storage.go
package storage

import "crypto-analytics/internal/models"

// Storage определяет контракт для работы с данными
type FormStorage interface {
	SaveContactFrom(contact *models.ContactForm) error
	ExportContactsToJSON(filename string) error
	Close()
}

type UserStorage interface {
	CreateUser(user *models.User) error
	GetUserByName(nameU string) (*models.User, error)
	GetAllFavoriteCoins(nameU string) ([]string, error)
	NewFavoriteCoin(nameU string, nameCoin string) error
	RemoveFavoriteCoin(nameU string, nameCoin string) error
	ExportUsersToJSON(filename string) error
	Close()
}

type NewsStorage interface {
	AddNews([]models.NewsItem) error
	GetAllNews() ([]models.NewsItem, error)
	UpdateNews([]models.NewsItem) error
}

type CacheStorage interface {
	Save(data []byte, amountPairs int) error
	Load() ([]string, error)
}

type AnalysisStorage interface {
	SaveAnalysisData(data models.PairsCrypto) error
	LoadAnalysisData() (models.PairsCrypto, error)
}
