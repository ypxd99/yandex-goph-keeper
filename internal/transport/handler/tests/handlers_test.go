package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ypxd99/yandex-gophkeeper/internal/model"
	"github.com/ypxd99/yandex-gophkeeper/internal/transport/handler"
)

// MockAuthService - мок сервиса аутентификации для тестирования хэндлеров
type MockAuthService struct {
	users map[string]*model.User
}

func NewMockAuthService() *MockAuthService {
	return &MockAuthService{
		users: make(map[string]*model.User),
	}
}

func (m *MockAuthService) Register(ctx context.Context, req *model.UserRegistration) (*model.AuthResponse, error) {
	// Создаем тестового пользователя
	user := &model.User{
		ID:           "user123",
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: "hashed_password",
		Salt:         "salt123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	m.users[user.ID] = user

	// Создаем токены
	tokenPair := &model.TokenPair{
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
	}

	return &model.AuthResponse{
		User: model.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			IsActive:  user.IsActive,
		},
		TokenPair: *tokenPair,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

func (m *MockAuthService) Login(ctx context.Context, req *model.UserLogin, userAgent, ipAddress string) (*model.AuthResponse, error) {
	// Простая проверка для тестов
	if req.Username == "testuser" && req.Password == "testpass123" {
		user := &model.User{
			ID:           "user123",
			Username:     req.Username,
			Email:        "test@example.com",
			PasswordHash: "hashed_password",
			Salt:         "salt123",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			IsActive:     true,
		}

		tokenPair := &model.TokenPair{
			AccessToken:  "access_token_123",
			RefreshToken: "refresh_token_123",
		}

		return &model.AuthResponse{
			User: model.UserResponse{
				ID:        user.ID,
				Username:  user.Username,
				Email:     user.Email,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				IsActive:  user.IsActive,
			},
			TokenPair: *tokenPair,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}, nil
	}

	return nil, model.ErrInvalidCredentials
}

func (m *MockAuthService) Refresh(ctx context.Context, req *model.RefreshTokenRequest) (*model.TokenPair, error) {
	if req.RefreshToken == "valid_refresh_token" {
		return &model.TokenPair{
			AccessToken:  "new_access_token",
			RefreshToken: "new_refresh_token",
		}, nil
	}
	return nil, model.ErrInvalidToken
}

func (m *MockAuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "valid_refresh_token" {
		return nil
	}
	return model.ErrInvalidToken
}

func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (*model.TokenClaims, error) {
	if token == "valid_access_token" {
		return &model.TokenClaims{
			UserID:   "user123",
			Username: "testuser",
			Exp:      time.Now().Add(time.Hour).Unix(),
		}, nil
	}
	return nil, model.ErrInvalidToken
}

func (m *MockAuthService) GetUserSessions(ctx context.Context, userID string) ([]*model.SessionResponse, error) {
	return []*model.SessionResponse{
		{
			ID:         "session123",
			UserAgent:  "Test Browser",
			IPAddress:  "127.0.0.1",
			CreatedAt:  time.Now(),
			LastUsedAt: time.Now(),
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		},
	}, nil
}

func (m *MockAuthService) RevokeSession(ctx context.Context, userID, sessionID string) error {
	if sessionID == "session123" {
		return nil
	}
	return model.ErrSessionNotFound
}

func (m *MockAuthService) LogoutAll(ctx context.Context, userID string) error {
	return nil
}

func (m *MockAuthService) ChangePassword(ctx context.Context, userID string, req *model.PasswordChangeRequest) error {
	if req.CurrentPassword == "currentpass" && req.NewPassword == "newpass123" {
		return nil
	}
	return model.ErrInvalidCurrentPassword
}

// MockSecretService - мок сервиса секретов для тестирования хэндлеров
type MockSecretService struct {
	secrets map[string]*model.Secret
}

func NewMockSecretService() *MockSecretService {
	return &MockSecretService{
		secrets: make(map[string]*model.Secret),
	}
}

func (m *MockSecretService) Create(ctx context.Context, userID string, req *model.CreateSecretRequest) (*model.Secret, error) {
	// Десериализуем данные секрета на основе типа (как в реальном сервисе)
	var secretData model.SecretData

	switch req.Type {
	case model.SecretTypeLoginPassword:
		var loginData model.LoginPasswordData
		if err := json.Unmarshal(req.Data, &loginData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal login data: %w", err)
		}
		secretData = &loginData
	case model.SecretTypeText:
		var textData model.TextData
		if err := json.Unmarshal(req.Data, &textData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal text data: %w", err)
		}
		secretData = &textData
	case model.SecretTypeBinary:
		var binaryData model.BinaryData
		if err := json.Unmarshal(req.Data, &binaryData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal binary data: %w", err)
		}
		secretData = &binaryData
	case model.SecretTypeCard:
		var cardData model.CardData
		if err := json.Unmarshal(req.Data, &cardData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal card data: %w", err)
		}
		secretData = &cardData
	case model.SecretTypeOTP:
		var otpData model.OTPData
		if err := json.Unmarshal(req.Data, &otpData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal OTP data: %w", err)
		}
		secretData = &otpData
	default:
		return nil, fmt.Errorf("unsupported secret type: %s", req.Type)
	}

	secret := &model.Secret{
		ID:          "secret123",
		UserID:      userID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Tags:        req.Tags,
		Data:        secretData,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Version:     1,
	}

	m.secrets[secret.ID] = secret
	return secret, nil
}

func (m *MockSecretService) GetByID(ctx context.Context, userID string, secretID string) (*model.Secret, error) {
	if secret, exists := m.secrets[secretID]; exists {
		if secret.UserID == userID {
			return secret, nil
		}
		return nil, model.ErrSecretNotFound
	}
	return nil, model.ErrSecretNotFound
}

func (m *MockSecretService) GetAll(ctx context.Context, userID string) ([]*model.Secret, error) {
	var userSecrets []*model.Secret
	for _, secret := range m.secrets {
		if secret.UserID == userID {
			userSecrets = append(userSecrets, secret)
		}
	}
	return userSecrets, nil
}

func (m *MockSecretService) GetByType(ctx context.Context, userID string, secretType model.SecretType) ([]*model.Secret, error) {
	var typeSecrets []*model.Secret
	for _, secret := range m.secrets {
		if secret.UserID == userID && secret.Type == secretType {
			typeSecrets = append(typeSecrets, secret)
		}
	}
	return typeSecrets, nil
}

func (m *MockSecretService) Update(ctx context.Context, userID string, secretID string, req *model.UpdateSecretRequest) (*model.Secret, error) {
	if secret, exists := m.secrets[secretID]; exists {
		if secret.UserID == userID {
			if req.Title != nil {
				secret.Title = *req.Title
			}
			if req.Description != nil {
				secret.Description = *req.Description
			}
			secret.Tags = req.Tags
			if req.Data != nil {
				secret.Data = req.Data
			}
			secret.UpdatedAt = time.Now()
			secret.Version++
			return secret, nil
		}
		return nil, model.ErrSecretNotFound
	}
	return nil, model.ErrSecretNotFound
}

func (m *MockSecretService) Delete(ctx context.Context, userID string, secretID string) error {
	if secret, exists := m.secrets[secretID]; exists {
		if secret.UserID == userID {
			now := time.Now()
			secret.DeletedAt = &now
			secret.UpdatedAt = now
			return nil
		}
		return model.ErrSecretNotFound
	}
	return model.ErrSecretNotFound
}

func (m *MockSecretService) SearchByTitle(ctx context.Context, userID string, query string) ([]*model.Secret, error) {
	var results []*model.Secret
	for _, secret := range m.secrets {
		if secret.UserID == userID && secret.DeletedAt == nil {
			// Простой поиск для тестов
			if secret.Title == query {
				results = append(results, secret)
			}
		}
	}
	return results, nil
}

func (m *MockSecretService) SearchByTags(ctx context.Context, userID string, tags []string) ([]*model.Secret, error) {
	var results []*model.Secret
	for _, secret := range m.secrets {
		if secret.UserID == userID && secret.DeletedAt == nil {
			// Простой поиск по тегам для тестов
			for _, tag := range tags {
				for _, secretTag := range secret.Tags {
					if tag == secretTag {
						results = append(results, secret)
						break
					}
				}
			}
		}
	}
	return results, nil
}

func (m *MockSecretService) GetVersions(ctx context.Context, userID, secretID string) ([]*model.SecretVersion, error) {
	return []*model.SecretVersion{}, nil
}

func (m *MockSecretService) RestoreVersion(ctx context.Context, userID, secretID string, version int) (*model.Secret, error) {
	return nil, nil
}

func (m *MockSecretService) Sync(ctx context.Context, userID string, lastSyncTime time.Time) ([]*model.Secret, error) {
	return []*model.Secret{}, nil
}

func TestHandlerRegister(t *testing.T) {
	// Настраиваем Gin для тестов
	gin.SetMode(gin.TestMode)

	// Создаем моки
	mockAuthService := NewMockAuthService()
	mockSecretService := NewMockSecretService()

	// Создаем хэндлер
	h := handler.InitHandler(mockAuthService, mockSecretService)

	// Создаем роутер
	router := gin.New()
	h.InitRoutes(router)

	// Тестовые данные
	reqBody := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "testpass123",
	}

	jsonData, _ := json.Marshal(reqBody)

	// Создаем HTTP запрос
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Создаем ResponseRecorder
	w := httptest.NewRecorder()

	// Выполняем запрос
	router.ServeHTTP(w, req)

	// Проверяем статус код
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	// Проверяем ответ
	var response model.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.User.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", response.User.Username)
	}

	if response.TokenPair.AccessToken == "" {
		t.Error("Expected access token to be present")
	}
}

func TestHandlerLogin(t *testing.T) {
	// Настраиваем Gin для тестов
	gin.SetMode(gin.TestMode)

	// Создаем моки
	mockAuthService := NewMockAuthService()
	mockSecretService := NewMockSecretService()

	// Создаем хэндлер
	h := handler.InitHandler(mockAuthService, mockSecretService)

	// Создаем роутер
	router := gin.New()
	h.InitRoutes(router)

	// Тестовые данные
	reqBody := map[string]interface{}{
		"username": "testuser",
		"password": "testpass123",
	}

	jsonData, _ := json.Marshal(reqBody)

	// Создаем HTTP запрос
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Test Browser")

	// Создаем ResponseRecorder
	w := httptest.NewRecorder()

	// Выполняем запрос
	router.ServeHTTP(w, req)

	// Проверяем статус код
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Проверяем ответ
	var response model.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.User.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", response.User.Username)
	}
}

func TestHandlerCreateSecret(t *testing.T) {
	// Настраиваем Gin для тестов
	gin.SetMode(gin.TestMode)

	// Создаем моки
	mockAuthService := NewMockAuthService()
	mockSecretService := NewMockSecretService()

	// Создаем хэндлер
	h := handler.InitHandler(mockAuthService, mockSecretService)

	// Создаем роутер
	router := gin.New()
	h.InitRoutes(router)

	// Тестовые данные
	reqBody := map[string]interface{}{
		"type":        "text",
		"title":       "Test Secret",
		"description": "Test Description",
		"tags":        []string{"test"},
		"data": map[string]interface{}{
			"text":  "Secret content",
			"notes": "Test notes",
		},
	}

	jsonData, _ := json.Marshal(reqBody)

	// Создаем HTTP запрос
	req, _ := http.NewRequest("POST", "/api/v1/secrets", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Создаем ResponseRecorder
	w := httptest.NewRecorder()

	// Выполняем запрос
	router.ServeHTTP(w, req)

	// Проверяем статус код (должен быть 401 без аутентификации)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestHandlerHealthCheck(t *testing.T) {
	// Настраиваем Gin для тестов
	gin.SetMode(gin.TestMode)

	// Создаем моки
	mockAuthService := NewMockAuthService()
	mockSecretService := NewMockSecretService()

	// Создаем хэндлер
	h := handler.InitHandler(mockAuthService, mockSecretService)

	// Создаем роутер
	router := gin.New()
	h.InitRoutes(router)

	// Создаем HTTP запрос
	req, _ := http.NewRequest("GET", "/health", nil)

	// Создаем ResponseRecorder
	w := httptest.NewRecorder()

	// Выполняем запрос
	router.ServeHTTP(w, req)

	// Проверяем статус код
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Проверяем ответ
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", response["status"])
	}
}
