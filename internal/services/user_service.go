package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
	"webdev-90-days/internal/models"
	"webdev-90-days/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userStorage  storage.UserStorage // отдельное хранилище для пользователей
	userSessions map[string]Session
	sessionMutex sync.RWMutex
}

type Session struct {
	Username  string
	IP        string
	Browser   string
	CreatedAt time.Time
}

func NewUserService(userStorage storage.UserStorage) *UserService {
	return &UserService{userStorage: userStorage,
		userSessions: make(map[string]Session),
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

		return fmt.Errorf("failed to register user: %w", err)
	}
	return nil
}

// LoginUser - вход пользователя
func (s *UserService) LoginUser(username, password string) error {
	User, err := s.userStorage.GetUserByName(username)
	if err != nil {
		return fmt.Errorf("We have not this acc")
	}

	err = bcrypt.CompareHashAndPassword([]byte(User.Password), []byte(password))
	return err
}

// остальные методы без изменений...
func (s *UserService) HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (s *UserService) CreateSession(username, ip, browser string) string {
	sessionID := s.generateSessionID()

	s.sessionMutex.Lock()
	s.userSessions[sessionID] = Session{
		Username:  username,
		IP:        ip,
		Browser:   browser,
		CreatedAt: time.Now(),
	}
	s.sessionMutex.Unlock()

	return sessionID
}
func (s *UserService) DeleteSession(cookie string) {
	delete(s.userSessions, cookie)

}

// Проверка сессии
func (s *UserService) ValidateSession(sessionID, currentIP, currentBrowser string) (string, bool) {
	s.sessionMutex.RLock()
	session, exists := s.userSessions[sessionID]
	s.sessionMutex.RUnlock()

	if !exists {
		return "", false
	}

	// Проверяем IP и браузер
	if session.IP == currentIP && session.Browser == currentBrowser {
		return session.Username, true
	}

	return "", false
}

// Генерация ID сессии
func (s *UserService) generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// // Очистка старых сессий (можно запускать в горутине)
// func (s *UserService) cleanupSessions() {
// 	s.sessionMutex.Lock()
// 	for id, session := range s.userSessions {
// 		if time.Since(session.CreatedAt) > 24*time.Hour {
// 			delete(s.userSessions, id)
// 		}
// 	}
// 	s.sessionMutex.Unlock()
// }
