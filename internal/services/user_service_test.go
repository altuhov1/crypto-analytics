package services

import (
	"errors"
	"testing"

	"webdev-90-days/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// Mock хранилища
type MockUserStorage struct {
	mock.Mock
}

func (m *MockUserStorage) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserStorage) GetUserByName(name string) (*models.User, error) {
	args := m.Called(name)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserStorage) GetAllFavoriteCoins(name string) ([]string, error) {
	args := m.Called(name)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockUserStorage) NewFavoriteCoin(name, coin string) error {
	args := m.Called(name, coin)
	return args.Error(0)
}

func (m *MockUserStorage) RemoveFavoriteCoin(name, coin string) error {
	args := m.Called(name, coin)
	return args.Error(0)
}

func (m *MockUserStorage) ExportUsersToJSON(filename string) error {
	args := m.Called(filename)
	return args.Error(0)
}

func TestUserService_RegisterUser_Success(t *testing.T) {
	mockStorage := new(MockUserStorage)
	service := NewUserService(mockStorage)

	user := &models.User{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	}

	// Ожидаем, что CreateUser будет вызван с любым пользователем
	// и вернет nil (успех)
	mockStorage.On("CreateUser", mock.AnythingOfType("*models.User")).Return(nil)

	err := service.RegisterUser(user)

	assert.NoError(t, err)
	assert.NotEqual(t, "password123", user.Password) // пароль должен быть захэширован
	mockStorage.AssertCalled(t, "CreateUser", mock.AnythingOfType("*models.User"))
}

func TestUserService_RegisterUser_StorageError(t *testing.T) {
	mockStorage := new(MockUserStorage)
	service := NewUserService(mockStorage)

	user := &models.User{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	}

	// Симулируем ошибку от хранилища
	expectedErr := errors.New("user already exists")
	mockStorage.On("CreateUser", mock.AnythingOfType("*models.User")).Return(expectedErr)

	err := service.RegisterUser(user)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestUserService_LoginUser_Success(t *testing.T) {
	mockStorage := new(MockUserStorage)
	service := NewUserService(mockStorage)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	expectedUser := &models.User{
		Username: "testuser",
		Password: string(hashedPassword),
		Email:    "test@example.com",
	}

	mockStorage.On("GetUserByName", "testuser").Return(expectedUser, nil)

	err := service.LoginUser("testuser", "password123")

	assert.NoError(t, err)
}

func TestUserService_LoginUser_InvalidPassword(t *testing.T) {
	mockStorage := new(MockUserStorage)
	service := NewUserService(mockStorage)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	expectedUser := &models.User{
		Username: "testuser",
		Password: string(hashedPassword),
		Email:    "test@example.com",
	}

	mockStorage.On("GetUserByName", "testuser").Return(expectedUser, nil)

	err := service.LoginUser("testuser", "wrongpassword")

	assert.Error(t, err)
}
