package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// SimpleCrypto простой сервис шифрования с одним ключом
type SimpleCrypto struct {
	key []byte
}

// InitSimpleCrypto создает или загружает ключ шифрования
func InitSimpleCrypto(keysDir string) *SimpleCrypto {
	keyPath := filepath.Join(keysDir, "master.key")

	// Создаем директорию если не существует
	if err := os.MkdirAll(keysDir, 0700); err != nil {
		panic(fmt.Sprintf("failed to create keys directory: %v", err))
	}

	var key []byte
	var err error

	// Пытаемся загрузить существующий ключ
	if _, statErr := os.Stat(keyPath); statErr == nil {
		key, err = os.ReadFile(keyPath)
		if err != nil {
			panic(fmt.Sprintf("failed to read master key: %v", err))
		}
	} else {
		// Генерируем новый ключ
		key = make([]byte, 32) // AES-256
		if _, err := io.ReadFull(rand.Reader, key); err != nil {
			panic(fmt.Sprintf("failed to generate master key: %v", err))
		}

		// Сохраняем ключ
		if err := os.WriteFile(keyPath, key, 0600); err != nil {
			panic(fmt.Sprintf("failed to save master key: %v", err))
		}
	}

	return &SimpleCrypto{key: key}
}

// Encrypt шифрует данные
func (sc *SimpleCrypto) Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(sc.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Создаем GCM режим
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Создаем nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to create nonce: %w", err)
	}

	// Шифруем
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// Decrypt расшифровывает данные
func (sc *SimpleCrypto) Decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(sc.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Создаем GCM режим
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Получаем размер nonce
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Извлекаем nonce и расшифровываем
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// EncryptString шифрует строку и возвращает base64
func (sc *SimpleCrypto) EncryptString(text string) (string, error) {
	if text == "" {
		return "", nil
	}

	encrypted, err := sc.Encrypt([]byte(text))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptString расшифровывает base64 строку
func (sc *SimpleCrypto) DecryptString(encryptedText string) (string, error) {
	if encryptedText == "" {
		return "", nil
	}

	// Декодируем base64
	data, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Расшифровываем
	decrypted, err := sc.Decrypt(data)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// GetDecryptedEmail расшифровывает email пользователя для проверки уникальности
func (sc *SimpleCrypto) GetDecryptedEmail(encryptedEmail string) string {
	if encryptedEmail == "" {
		return ""
	}

	decrypted, err := sc.DecryptString(encryptedEmail)
	if err != nil {
		// Если не удалось расшифровать, возвращаем пустую строку
		return ""
	}

	return decrypted
}
