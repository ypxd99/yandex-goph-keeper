package util

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ypxd99/yandex-gophkeeper/internal/model"
)

// JWTManager управляет JWT токенами
type JWTManager struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewJWTManager создает новый менеджер JWT
func NewJWTManager(secretKey string, accessExpiry, refreshExpiry time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateTokenPair генерирует пару токенов доступа и обновления
func (j *JWTManager) GenerateTokenPair(userID, username string) (*model.TokenPair, error) {
	now := time.Now()

	// Создаем access token
	accessClaims := &model.TokenClaims{
		UserID:   userID,
		Username: username,
		Exp:      now.Add(j.accessExpiry).Unix(),
		Iat:      now.Unix(),
		Type:     "access",
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(j.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Создаем refresh token
	refreshClaims := &model.TokenClaims{
		UserID:   userID,
		Username: username,
		Exp:      now.Add(j.refreshExpiry).Unix(),
		Iat:      now.Unix(),
		Type:     "refresh",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &model.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(j.accessExpiry.Seconds()),
	}, nil
}

// ValidateToken проверяет и парсит JWT токен
func (j *JWTManager) ValidateToken(tokenString string) (*model.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*model.TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Проверяем тип токена
	if claims.Type != "access" && claims.Type != "refresh" {
		return nil, fmt.Errorf("invalid token type")
	}

	return claims, nil
}

// RefreshToken обновляет access token используя refresh token
func (j *JWTManager) RefreshToken(refreshTokenString string) (*model.TokenPair, error) {
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Проверяем что это refresh token
	if claims.Type != "refresh" {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	// Генерируем новую пару токенов
	return j.GenerateTokenPair(claims.UserID, claims.Username)
}

// GenerateRandomString генерирует случайную строку заданной длины
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
