package model

import (
	"time"
)

// User представляет пользователя системы
type User struct {
	ID             string    `json:"id" db:"id"`
	Username       string    `json:"username" db:"username"`
	Email          string    `json:"email" db:"email"`
	EmailEncrypted string    `json:"email_encrypted" db:"email_encrypted"` // Зашифрованный email
	PasswordHash   string    `json:"password_hash" db:"password_hash"`
	Salt           string    `json:"salt" db:"salt"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	LastLoginAt    time.Time `json:"last_login_at" db:"last_login_at"`
	IsActive       bool      `json:"is_active" db:"is_active"`
}

// UserRegistration представляет данные для регистрации пользователя
type UserRegistration struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// UserLogin представляет данные для входа пользователя
type UserLogin struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	UserAgent string `json:"user_agent"`
	IPAddress string `json:"ip_address"`
}

// UserResponse представляет ответ с данными пользователя (без секретной информации)
type UserResponse struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastLoginAt time.Time `json:"last_login_at"`
	IsActive    bool      `json:"is_active"`
}
