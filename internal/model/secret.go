package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// SecretType представляет тип секрета
type SecretType string

const (
	SecretTypeLoginPassword SecretType = "login_password"
	SecretTypeText          SecretType = "text"
	SecretTypeBinary        SecretType = "binary"
	SecretTypeCard          SecretType = "card"
	SecretTypeOTP           SecretType = "otp"
)

// unmarshalSecretData - общая функция для десериализации SecretData
// Принимает JSON данные и тип секрета, возвращает десериализованные данные
func unmarshalSecretData(data json.RawMessage, secretType SecretType) (SecretData, error) {
	if data == nil {
		return nil, nil
	}

	switch secretType {
	case SecretTypeLoginPassword:
		var loginData LoginPasswordData
		if err := json.Unmarshal(data, &loginData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal login data: %w", err)
		}
		return &loginData, nil
	case SecretTypeText:
		var textData TextData
		if err := json.Unmarshal(data, &textData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal text data: %w", err)
		}
		return &textData, nil
	case SecretTypeBinary:
		var binaryData BinaryData
		if err := json.Unmarshal(data, &binaryData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal binary data: %w", err)
		}
		return &binaryData, nil
	case SecretTypeCard:
		var cardData CardData
		if err := json.Unmarshal(data, &cardData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal card data: %w", err)
		}
		return &cardData, nil
	case SecretTypeOTP:
		var otpData OTPData
		if err := json.Unmarshal(data, &otpData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal OTP data: %w", err)
		}
		return &otpData, nil
	default:
		return nil, fmt.Errorf("unsupported secret type: %s", secretType)
	}
}

// Secret представляет секрет пользователя
type Secret struct {
	ID          string     `json:"id" db:"id"`
	UserID      string     `json:"user_id" db:"user_id"`
	Type        SecretType `json:"type" db:"type"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Tags        []string   `json:"tags" db:"tags"`
	Data        SecretData `json:"data" db:"data"`
	Version     int        `json:"version" db:"version"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UnmarshalJSON реализует custom unmarshaling для правильной десериализации SecretData
func (s *Secret) UnmarshalJSON(data []byte) error {
	// Создаем временную структуру для парсинга
	type tempSecret struct {
		ID          string          `json:"id"`
		UserID      string          `json:"user_id"`
		Type        SecretType      `json:"type"`
		Title       string          `json:"title"`
		Description string          `json:"description"`
		Tags        []string        `json:"tags"`
		Data        json.RawMessage `json:"data"`
		Version     int             `json:"version"`
		CreatedAt   time.Time       `json:"created_at"`
		UpdatedAt   time.Time       `json:"updated_at"`
		DeletedAt   *time.Time      `json:"deleted_at,omitempty"`
	}

	var temp tempSecret
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Копируем основные поля
	s.ID = temp.ID
	s.UserID = temp.UserID
	s.Type = temp.Type
	s.Title = temp.Title
	s.Description = temp.Description
	s.Tags = temp.Tags
	s.Version = temp.Version
	s.CreatedAt = temp.CreatedAt
	s.UpdatedAt = temp.UpdatedAt
	s.DeletedAt = temp.DeletedAt

	// Используем общую функцию для десериализации данных
	secretData, err := unmarshalSecretData(temp.Data, temp.Type)
	if err != nil {
		return err
	}
	s.Data = secretData

	return nil
}

// SecretData представляет интерфейс для данных секрета
type SecretData interface {
	GetType() SecretType
}

// LoginPasswordData представляет данные для логина и пароля
type LoginPasswordData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	URL      string `json:"url"`
	Notes    string `json:"notes"`
	// Зашифрованные поля
	LoginEncrypted    string `json:"login_encrypted"`
	PasswordEncrypted string `json:"password_encrypted"`
	URLEncrypted      string `json:"url_encrypted"`
	NotesEncrypted    string `json:"notes_encrypted"`
}

// GetType возвращает тип данных
func (d *LoginPasswordData) GetType() SecretType {
	return SecretTypeLoginPassword
}

// TextData представляет текстовые данные
type TextData struct {
	Text  string `json:"text"`
	Notes string `json:"notes"`
	// Зашифрованные поля
	TextEncrypted  string `json:"text_encrypted"`
	NotesEncrypted string `json:"notes_encrypted"`
}

// GetType возвращает тип данных
func (d *TextData) GetType() SecretType {
	return SecretTypeText
}

// BinaryData представляет бинарные данные
type BinaryData struct {
	Data  []byte `json:"data"`
	Notes string `json:"notes"`
	// Зашифрованные поля
	DataEncrypted  string `json:"data_encrypted"`
	NotesEncrypted string `json:"notes_encrypted"`
}

// GetType возвращает тип данных
func (d *BinaryData) GetType() SecretType {
	return SecretTypeBinary
}

// CardData представляет данные банковской карты
type CardData struct {
	Number     string `json:"number"`
	ExpiryDate string `json:"expiry_date"`
	CVV        string `json:"cvv"`
	Cardholder string `json:"cardholder"`
	Notes      string `json:"notes"`
	// Зашифрованные поля
	NumberEncrypted     string `json:"number_encrypted"`
	ExpiryDateEncrypted string `json:"expiry_date_encrypted"`
	CVVEncrypted        string `json:"cvv_encrypted"`
	CardholderEncrypted string `json:"cardholder_encrypted"`
	NotesEncrypted      string `json:"notes_encrypted"`
}

// GetType возвращает тип данных
func (d *CardData) GetType() SecretType {
	return SecretTypeCard
}

// OTPData представляет данные для OTP
type OTPData struct {
	Secret string `json:"secret"`
	Type   string `json:"type"` // "TOTP" или "HOTP"
	Digits int    `json:"digits"`
	Period int    `json:"period"` // для TOTP
	Notes  string `json:"notes"`
	// Зашифрованные поля
	SecretEncrypted string `json:"secret_encrypted"`
	TypeEncrypted   string `json:"type_encrypted"`
	NotesEncrypted  string `json:"notes_encrypted"`
}

// GetType возвращает тип данных
func (d *OTPData) GetType() SecretType {
	return SecretTypeOTP
}

// CreateSecretRequest представляет запрос на создание секрета
type CreateSecretRequest struct {
	Type        SecretType      `json:"type" binding:"required"`
	Title       string          `json:"title" binding:"required"`
	Description string          `json:"description"`
	Tags        []string        `json:"tags"`
	Data        json.RawMessage `json:"data" binding:"required"`
}

// UpdateSecretRequest представляет запрос на обновление секрета
type UpdateSecretRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Tags        []string   `json:"tags"`
	Data        SecretData `json:"data"`
}

// UnmarshalJSON реализует custom unmarshaling для правильной десериализации SecretData
func (r *UpdateSecretRequest) UnmarshalJSON(data []byte) error {
	// Создаем временную структуру для парсинга
	type tempRequest struct {
		Title       *string         `json:"title"`
		Description *string         `json:"description"`
		Tags        []string        `json:"tags"`
		Data        json.RawMessage `json:"data"`
		Type        SecretType      `json:"type"` // Нужен для определения типа данных
	}

	var temp tempRequest
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Копируем основные поля
	r.Title = temp.Title
	r.Description = temp.Description
	r.Tags = temp.Tags

	// Используем общую функцию для десериализации данных
	secretData, err := unmarshalSecretData(temp.Data, temp.Type)
	if err != nil {
		return err
	}
	r.Data = secretData

	return nil
}

// SecretVersion представляет версию секрета
type SecretVersion struct {
	ID          string     `json:"id" db:"id"`
	SecretID    string     `json:"secret_id" db:"secret_id"`
	UserID      string     `json:"user_id" db:"user_id"`
	Type        SecretType `json:"type" db:"type"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Tags        []string   `json:"tags" db:"tags"`
	Data        SecretData `json:"data" db:"data"`
	Version     int        `json:"version" db:"version"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	IsActive    bool       `json:"is_active" db:"is_active"`
}

// UnmarshalJSON реализует custom unmarshaling для правильной десериализации SecretData
func (sv *SecretVersion) UnmarshalJSON(data []byte) error {
	// Создаем временную структуру для парсинга
	type tempSecretVersion struct {
		ID          string          `json:"id"`
		SecretID    string          `json:"secret_id"`
		UserID      string          `json:"user_id"`
		Type        SecretType      `json:"type"`
		Title       string          `json:"title"`
		Description string          `json:"description"`
		Tags        []string        `json:"tags"`
		Data        json.RawMessage `json:"data"`
		Version     int             `json:"version"`
		CreatedAt   time.Time       `json:"created_at"`
		UpdatedAt   time.Time       `json:"updated_at"`
		IsActive    bool            `json:"is_active"`
	}

	var temp tempSecretVersion
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Копируем основные поля
	sv.ID = temp.ID
	sv.SecretID = temp.SecretID
	sv.UserID = temp.UserID
	sv.Type = temp.Type
	sv.Title = temp.Title
	sv.Description = temp.Description
	sv.Tags = temp.Tags
	sv.Version = temp.Version
	sv.CreatedAt = temp.CreatedAt
	sv.UpdatedAt = temp.UpdatedAt
	sv.IsActive = temp.IsActive

	// Используем общую функцию для десериализации данных
	secretData, err := unmarshalSecretData(temp.Data, temp.Type)
	if err != nil {
		return err
	}
	sv.Data = secretData

	return nil
}
