package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// PasswordHasher управляет хешированием паролей
type PasswordHasher struct{}

// NewPasswordHasher создает новый хешер паролей
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{}
}

// HashPassword хеширует пароль с солью
func (p *PasswordHasher) HashPassword(password string) (string, string, error) {
	// Генерируем случайную соль
	salt, err := GenerateRandomString(16)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Хешируем пароль с солью
	hash := sha256.Sum256([]byte(password + salt))
	hashString := hex.EncodeToString(hash[:])

	return hashString, salt, nil
}

// CheckPassword проверяет пароль
func (p *PasswordHasher) CheckPassword(password, hash, salt string) bool {
	// Хешируем введенный пароль с той же солью
	checkHash := sha256.Sum256([]byte(password + salt))
	checkHashString := hex.EncodeToString(checkHash[:])

	return hash == checkHashString
}
