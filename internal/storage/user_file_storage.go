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

func (s *UserFileStorage) loadUsers() ([]*models.User, error) {
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
	users, err := s.loadUsers()
	if err != nil {
		return err
	}

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
	users, err := s.loadUsers()
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

func (s *UserFileStorage) GetAllFavoriteCoins(nameU string) ([]string, error) {
	users, err := s.loadUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Username == nameU {
			return user.FavoriteCoins, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

func (s *UserFileStorage) NewFavoriteCoin(nameU string, nameCoin string) error {
	users, err := s.loadUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Username == nameU {
			for _, coin := range user.FavoriteCoins {
				if coin == nameCoin {
					return fmt.Errorf("coin already in list of favorite coins")
				}
			}
			user.FavoriteCoins = append(user.FavoriteCoins, nameCoin)
			if err := s.saveUsers(users); err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("user not found")
}
func (s *UserFileStorage) RemoveFavoriteCoin(nameU string, nameCoin string) error {
	users, err := s.loadUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Username == nameU {
			for i, coin := range user.FavoriteCoins {
				if coin == nameCoin {
					user.FavoriteCoins = removeElement(user.FavoriteCoins, i)
					if err := s.saveUsers(users); err != nil {
						return err
					}
					return nil
				}
			}
			return fmt.Errorf("coin does not in list")
		}
	}

	return fmt.Errorf("user not found")
}

func removeElement(slice []string, i int) []string {
	if i < 0 || i >= len(slice) {
		return slice // возвращаем исходный срез если индекс вне диапазона
	}
	return append(slice[:i], slice[i+1:]...)
}
