package tests

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ypxd99/yandex-gophkeeper/internal/model"
)

func TestUserCreation(t *testing.T) {
	user := &model.User{
		ID:           "test_user_123",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Salt:         "salt123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	if user.ID != "test_user_123" {
		t.Errorf("Expected user ID to be 'test_user_123', got '%s'", user.ID)
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username to be 'testuser', got '%s'", user.Username)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email to be 'test@example.com', got '%s'", user.Email)
	}

	if !user.IsActive {
		t.Error("Expected user to be active")
	}
}

func TestSecretCreation(t *testing.T) {
	secret := &model.Secret{
		ID:          "secret_123",
		UserID:      "user_123",
		Type:        model.SecretTypeLoginPassword,
		Title:       "Test Secret",
		Description: "Test Description",
		Tags:        []string{"test", "demo"},
		Data: &model.LoginPasswordData{
			Login:    "testuser",
			Password: "testpass",
			URL:      "https://example.com",
			Notes:    "Test notes",
		},
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if secret.ID != "secret_123" {
		t.Errorf("Expected secret ID to be 'secret_123', got '%s'", secret.ID)
	}

	if secret.Type != model.SecretTypeLoginPassword {
		t.Errorf("Expected secret type to be SecretTypeLoginPassword, got %v", secret.Type)
	}

	if secret.Title != "Test Secret" {
		t.Errorf("Expected title to be 'Test Secret', got '%s'", secret.Title)
	}

	if len(secret.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(secret.Tags))
	}

	// Проверяем данные секрета
	if loginData, ok := secret.Data.(*model.LoginPasswordData); ok {
		if loginData.Login != "testuser" {
			t.Errorf("Expected login to be 'testuser', got '%s'", loginData.Login)
		}
		if loginData.Password != "testpass" {
			t.Errorf("Expected password to be 'testpass', got '%s'", loginData.Password)
		}
	} else {
		t.Error("Expected secret data to be LoginPasswordData")
	}
}

func TestSecretDataTypes(t *testing.T) {
	// Тест для текстового секрета
	textData := &model.TextData{
		Text:  "Secret text content",
		Notes: "Text secret notes",
	}

	if textData.GetType() != model.SecretTypeText {
		t.Errorf("Expected TextData type to be SecretTypeText, got %v", textData.GetType())
	}

	// Тест для банковской карты
	cardData := &model.CardData{
		Number:     "1234-5678-9012-3456",
		ExpiryDate: "12/25",
		CVV:        "123",
		Cardholder: "John Doe",
		Notes:      "Main credit card",
	}

	if cardData.GetType() != model.SecretTypeCard {
		t.Errorf("Expected CardData type to be SecretTypeCard, got %v", cardData.GetType())
	}

	// Тест для OTP
	otpData := &model.OTPData{
		Secret: "JBSWY3DPEHPK3PXP",
		Type:   "TOTP",
		Digits: 6,
		Period: 30,
		Notes:  "Google Authenticator",
	}

	if otpData.GetType() != model.SecretTypeOTP {
		t.Errorf("Expected OTPData type to be SecretTypeOTP, got %v", otpData.GetType())
	}
}

func TestCreateSecretRequest(t *testing.T) {
	// Создаем тестовые данные
	textData := &model.TextData{
		Text:  "Test text",
		Notes: "Test notes",
	}

	// Сериализуем данные в JSON
	dataJSON, err := json.Marshal(textData)
	if err != nil {
		t.Fatalf("Failed to marshal text data: %v", err)
	}

	req := &model.CreateSecretRequest{
		Type:        model.SecretTypeText,
		Title:       "Test Secret",
		Description: "Test Description",
		Tags:        []string{"test"},
		Data:        dataJSON,
	}

	// Проверяем, что все поля установлены правильно
	if req.Type != model.SecretTypeText {
		t.Errorf("Expected type to be SecretTypeText, got %v", req.Type)
	}

	if req.Title != "Test Secret" {
		t.Errorf("Expected title to be 'Test Secret', got '%s'", req.Title)
	}

	if req.Description != "Test Description" {
		t.Errorf("Expected description to be 'Test Description', got '%s'", req.Description)
	}

	if len(req.Tags) != 1 || req.Tags[0] != "test" {
		t.Errorf("Expected tags to be ['test'], got %v", req.Tags)
	}

	// Проверяем, что Data содержит правильные JSON данные
	var decodedData model.TextData
	if err := json.Unmarshal(req.Data, &decodedData); err != nil {
		t.Errorf("Failed to unmarshal data: %v", err)
	}

	if decodedData.Text != "Test text" {
		t.Errorf("Expected text to be 'Test text', got '%s'", decodedData.Text)
	}

	if decodedData.Notes != "Test notes" {
		t.Errorf("Expected notes to be 'Test notes', got '%s'", decodedData.Notes)
	}
}

func TestUserRegistration(t *testing.T) {
	reg := &model.UserRegistration{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "securepassword123",
	}

	if reg.Username != "newuser" {
		t.Errorf("Expected username to be 'newuser', got '%s'", reg.Username)
	}

	if reg.Email != "newuser@example.com" {
		t.Errorf("Expected email to be 'newuser@example.com', got '%s'", reg.Email)
	}

	if reg.Password != "securepassword123" {
		t.Errorf("Expected password to be 'securepassword123', got '%s'", reg.Password)
	}
}

func TestUserLogin(t *testing.T) {
	login := &model.UserLogin{
		Username: "existinguser",
		Password: "userpassword",
	}

	if login.Username != "existinguser" {
		t.Errorf("Expected username to be 'existinguser', got '%s'", login.Username)
	}

	if login.Password != "userpassword" {
		t.Errorf("Expected password to be 'userpassword', got '%s'", login.Password)
	}
}
