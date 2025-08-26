package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenPair представляет пару токенов доступа и обновления
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// TokenClaims представляет claims JWT токена
type TokenClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
	Type     string `json:"type"` // "access" или "refresh"
}

// GetExpirationTime возвращает время истечения токена
func (t *TokenClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	if t.Exp == 0 {
		return nil, nil
	}
	exp := jwt.NewNumericDate(time.Unix(t.Exp, 0))
	return exp, nil
}

// GetNotBefore возвращает время начала действия токена
func (t *TokenClaims) GetNotBefore() (*jwt.NumericDate, error) {
	if t.Iat == 0 {
		return nil, nil
	}
	iat := jwt.NewNumericDate(time.Unix(t.Iat, 0))
	return iat, nil
}

// GetIssuedAt возвращает время выдачи токена
func (t *TokenClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	if t.Iat == 0 {
		return nil, nil
	}
	iat := jwt.NewNumericDate(time.Unix(t.Iat, 0))
	return iat, nil
}

// GetIssuer возвращает издателя токена
func (t *TokenClaims) GetIssuer() (string, error) {
	return "", nil
}

// GetSubject возвращает субъект токена
func (t *TokenClaims) GetSubject() (string, error) {
	return t.UserID, nil
}

// GetAudience возвращает аудиторию токена
func (t *TokenClaims) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

// RefreshTokenRequest представляет запрос на обновление токена
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse представляет ответ на успешную аутентификацию
type AuthResponse struct {
	User      UserResponse `json:"user"`
	TokenPair TokenPair    `json:"tokens"`
	ExpiresAt time.Time    `json:"expires_at"`
}

// PasswordChangeRequest представляет запрос на смену пароля
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// Session представляет активную сессию пользователя
type Session struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	LastUsedAt   time.Time `json:"last_used_at" db:"last_used_at"`
	IsActive     bool      `json:"is_active" db:"is_active"`
}

// SessionResponse представляет ответ с данными сессии
type SessionResponse struct {
	ID         string    `json:"id"`
	UserAgent  string    `json:"user_agent"`
	IPAddress  string    `json:"ip_address"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
}
