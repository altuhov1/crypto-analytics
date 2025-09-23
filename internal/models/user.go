package models

type User struct {
	Email    string `json:"email"`
	Password string `json:"-"` // не сериализуем в JSON
	Username string `json:"name"`
}
