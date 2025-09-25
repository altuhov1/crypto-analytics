package models

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"name"`
}
