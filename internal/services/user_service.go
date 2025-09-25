package services

import (
	"fmt"
	"webdev-90-days/internal/models"
	"webdev-90-days/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userStorage storage.UserStorage // отдельное хранилище для пользователей
}

func NewUserService(userStorage storage.UserStorage) *UserService {
	return &UserService{userStorage: userStorage}
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

		return fmt.Errorf("failed to register user: %w", err)
	}
	return nil
}

// LoginUser - вход пользователя
func (s *UserService) LoginUser(email, password string) (*models.User, error) {
	// TODO: будет реализовывать на следующем этапе

	//err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	//return err == nil
	return nil, nil
}

// остальные методы без изменений...
func (s *UserService) HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}
