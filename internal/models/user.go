package models

import "time"

// User представляет пользователя нашего приложения
type User struct {
	ID        int       // Уникальный номер пользователя
	Email     string    // Email пользователя
	Username  string    // Имя пользователя (например, "crypto_trader")
	CreatedAt time.Time // Когда создан аккаунт
}
