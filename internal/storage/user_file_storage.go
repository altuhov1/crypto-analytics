package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"webdev-90-days/internal/models"
)

type UserFileStorage struct {
	filename string
	mu       sync.RWMutex
}

func NewUserFileStorage(filename string) *UserFileStorage {
	return &UserFileStorage{
		filename: filename,
		mu:       sync.RWMutex{},
	}
}

func (s *UserFileStorage) LoadUsers() ([]*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, err := os.Open(s.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []*models.User{}, nil // Возвращаем пустой массив если файла нет
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var users []*models.User
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return users, nil
}

func (s *UserFileStorage) saveUsers(users []*models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Create(s.filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(users); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

func (s *UserFileStorage) CreateUser(user *models.User) error {
	users, err := s.LoadUsers()
	if err != nil {
		return err
	}

	// Проверяем, существует ли пользователь с таким email
	for _, u := range users {
		if u.Email == user.Email {
			return fmt.Errorf("user already exists")
		}
		if u.Username == user.Username {
			return fmt.Errorf("user name already exists")
		}
	}

	// Добавляем пользователя
	users = append(users, user)

	// Сохраняем данные
	if err := s.saveUsers(users); err != nil {
		return err
	}

	return nil
}

func (s *UserFileStorage) GetUserByName(nameU string) (*models.User, error) {
	users, err := s.LoadUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Username == nameU {
			return &models.User{
				Email:    user.Email,
				Password: user.Password,
				Username: user.Username,
			}, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

func (s *UserFileStorage) GetAllUsers() ([]*models.User, error) {
	users, err := s.LoadUsers()
	if err != nil {
		return nil, err
	}

	// Возвращаем копии пользователей без паролей для безопасности
	result := make([]*models.User, len(users))
	for i, user := range users {
		result[i] = &models.User{
			Email:    user.Email,
			Password: "", // Не возвращаем пароли
			Username: user.Username,
		}
	}
	return result, nil
}

func (s *UserFileStorage) UpdateUser(updatedUser *models.User) error {
	users, err := s.LoadUsers()
	if err != nil {
		return err
	}

	for i, user := range users {
		if user.Email == updatedUser.Email {
			// Если email изменился, проверяем уникальность нового email
			if user.Email != updatedUser.Email {
				for j, u := range users {
					if i != j && u.Email == updatedUser.Email {
						return fmt.Errorf("email already exists")
					}
				}
			}

			users[i] = updatedUser
			return s.saveUsers(users)
		}
	}

	return fmt.Errorf("user not found")
}

func (s *UserFileStorage) DeleteUser(email string) error {
	users, err := s.LoadUsers()
	if err != nil {
		return err
	}

	for i, user := range users {
		if user.Email == email {
			// Удаляем пользователя из slice
			users = append(users[:i], users[i+1:]...)
			return s.saveUsers(users)
		}
	}

	return fmt.Errorf("user not found")
}

// Дополнительный метод для проверки существования пользователя
func (s *UserFileStorage) UserExists(email string) (bool, error) {
	users, err := s.LoadUsers()
	if err != nil {
		return false, err
	}

	for _, user := range users {
		if user.Email == email {
			return true, nil
		}
	}

	return false, nil
}
