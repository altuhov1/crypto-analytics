package services

import (
	"fmt"
	"sync"
	"webdev-90-days/internal/models"
	"webdev-90-days/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userStorage storage.UserStorage // отдельное хранилище для пользователей
	// userSessions map[string]Session
	sessionMutex sync.RWMutex
}

func NewUserService(userStorage storage.UserStorage) *UserService {
	return &UserService{userStorage: userStorage,
		sessionMutex: sync.RWMutex{},
	}
}

// RegisterUser - регистрация нового пользователя
func (s *UserService) RegisterUser(user *models.User) error {
	var err error
	user.Password, err = s.HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("failed Hashing: %w", err)
	}
	err = s.userStorage.CreateUser(user)
	if err != nil {

		return err
	}
	return nil
}

func (s *UserService) LoginUser(username, password string) error {
	User, err := s.userStorage.GetUserByName(username)
	if err != nil {
		return fmt.Errorf("we have not this acc")
	}

	err = bcrypt.CompareHashAndPassword([]byte(User.Password), []byte(password))
	return err
}

func (s *UserService) HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (s *UserService) AddFavorite(username, CoinID string) error {
	err := s.userStorage.NewFavoriteCoin(username, CoinID)
	if err != nil {
		return fmt.Errorf("in AddFavorite: %w", err)
	}
	return nil
}
func (s *UserService) RemoveFavorite(username, CoinID string) error {
	err := s.userStorage.RemoveFavoriteCoin(username, CoinID)
	if err != nil {
		return fmt.Errorf("in RemoveFavorite: %w", err)
	}
	return nil
}

func (s *UserService) GetFavorites(username string) ([]string, error) {
	allFavC, err := s.userStorage.GetAllFavoriteCoins(username)
	if err != nil {
		return nil, fmt.Errorf("in RemoveFavorite: %w", err)
	}
	return allFavC, nil
}
