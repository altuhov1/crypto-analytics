package storage

import (
	"fmt"
	"webdev-90-days/internal/models"
)

type UserFileStorage struct {
	filename string
}

func NewUserFileStorage(filename string) *UserFileStorage {
	return &UserFileStorage{filename: filename}
}

func (s *UserFileStorage) CreateUser(user *models.User) error {
	// Пока просто заглушка - всегда возвращаем ошибку "пользователь существует"
	return fmt.Errorf("user already exists")
}

func (s *UserFileStorage) GetUserByEmail(email string) (*models.User, error) {
	// Пока заглушка - пользователь не найден
	return nil, fmt.Errorf("user not found")
}

func (s *UserFileStorage) GetUserByID(id int) (*models.User, error) {
	// Пока заглушка - пользователь не найден
	return nil, fmt.Errorf("user not found")
}
