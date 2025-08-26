package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/ypxd99/yandex-gophkeeper/internal/model"
	"github.com/ypxd99/yandex-gophkeeper/internal/repository"
	"github.com/ypxd99/yandex-gophkeeper/internal/repository/storage"
	"github.com/ypxd99/yandex-gophkeeper/internal/service"
	"github.com/ypxd99/yandex-gophkeeper/util"
)

// MockSecretRepository - мок репозитория секретов для тестирования
type MockSecretRepository struct {
	secrets map[string]*model.Secret
	users   map[string][]*model.Secret
}

func NewMockSecretRepository() *MockSecretRepository {
	return &MockSecretRepository{
		secrets: make(map[string]*model.Secret),
		users:   make(map[string][]*model.Secret),
	}
}

func (m *MockSecretRepository) Create(ctx context.Context, secret *model.Secret) error {
	m.secrets[secret.ID] = secret
	m.users[secret.UserID] = append(m.users[secret.UserID], secret)
	return nil
}

func (m *MockSecretRepository) GetByID(ctx context.Context, id string) (*model.Secret, error) {
	if secret, exists := m.secrets[id]; exists {
		return secret, nil
	}
	return nil, model.ErrSecretNotFound
}

func (m *MockSecretRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Secret, error) {
	if secrets, exists := m.users[userID]; exists {
		return secrets, nil
	}
	return []*model.Secret{}, nil
}

func (m *MockSecretRepository) GetByType(ctx context.Context, userID string, secretType model.SecretType) ([]*model.Secret, error) {
	if secrets, exists := m.users[userID]; exists {
		var filtered []*model.Secret
		for _, secret := range secrets {
			if secret.Type == secretType {
				filtered = append(filtered, secret)
			}
		}
		return filtered, nil
	}
	return []*model.Secret{}, nil
}

func (m *MockSecretRepository) Update(ctx context.Context, secret *model.Secret) error {
	if _, exists := m.secrets[secret.ID]; exists {
		m.secrets[secret.ID] = secret
		return nil
	}
	return model.ErrSecretNotFound
}

func (m *MockSecretRepository) Delete(ctx context.Context, id string) error {
	if _, exists := m.secrets[id]; exists {
		delete(m.secrets, id)
		return nil
	}
	return model.ErrSecretNotFound
}

func (m *MockSecretRepository) SearchByTitle(ctx context.Context, userID string, query string) ([]*model.Secret, error) {
	// Простая реализация поиска для тестов
	return m.GetByUserID(ctx, userID)
}

func (m *MockSecretRepository) SearchByTags(ctx context.Context, userID string, tags []string) ([]*model.Secret, error) {
	// Простая реализация поиска по тегам для тестов
	return m.GetByUserID(ctx, userID)
}

// GetVersions возвращает все версии секрета
func (m *MockSecretRepository) GetVersions(ctx context.Context, secretID string) ([]*model.SecretVersion, error) {
	// Для мока возвращаем пустой слайс
	return []*model.SecretVersion{}, nil
}

// CreateVersion создает новую версию секрета
func (m *MockSecretRepository) CreateVersion(ctx context.Context, version *model.SecretVersion) error {
	// Для мока просто возвращаем успех
	return nil
}

// GetVersionByNumber возвращает версию секрета по номеру
func (m *MockSecretRepository) GetVersionByNumber(ctx context.Context, secretID string, versionNumber int) (*model.SecretVersion, error) {
	// Для мока возвращаем ошибку "не найдено"
	return nil, fmt.Errorf("version not found")
}

// MockStorage - мок хранилища для тестирования
type MockStorage struct {
	secrets map[string]*model.Secret
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		secrets: make(map[string]*model.Secret),
	}
}

func (m *MockStorage) Close() error {
	return nil
}

func (m *MockStorage) GetUsers() []*model.User {
	return []*model.User{}
}

func (m *MockStorage) GetSecrets() []*model.Secret {
	secrets := make([]*model.Secret, 0, len(m.secrets))
	for _, secret := range m.secrets {
		secrets = append(secrets, secret)
	}
	return secrets
}

func (m *MockStorage) GetSessions() []*model.Session {
	return []*model.Session{}
}

func (m *MockStorage) Save() error {
	return nil
}

func (m *MockStorage) GenerateID() string {
	return "test_id"
}

func (m *MockStorage) GetUserByID(id string) (*model.User, error) {
	return nil, model.ErrUserNotFound
}

func (m *MockStorage) GetUserByUsername(username string) (*model.User, error) {
	return nil, model.ErrUserNotFound
}

func (m *MockStorage) GetUserByEmail(email string) (*model.User, error) {
	return nil, model.ErrUserNotFound
}

func (m *MockStorage) CreateUser(user *model.User) error {
	return nil
}

func (m *MockStorage) UpdateUser(user *model.User) error {
	return nil
}

func (m *MockStorage) DeleteUser(id string) error {
	return nil
}

func (m *MockStorage) UserExists(username string) bool {
	return false
}

// Методы для работы с секретами
func (m *MockStorage) CreateSecret(secret *model.Secret) error {
	m.secrets[secret.ID] = secret
	return nil
}

func (m *MockStorage) UpdateSecret(secret *model.Secret) error {
	if _, exists := m.secrets[secret.ID]; exists {
		m.secrets[secret.ID] = secret
		return nil
	}
	return fmt.Errorf("secret not found")
}

func (m *MockStorage) DeleteSecret(id string) error {
	if _, exists := m.secrets[id]; exists {
		delete(m.secrets, id)
		return nil
	}
	return fmt.Errorf("secret not found")
}

// Методы для версионирования секретов
func (m *MockStorage) GetSecretVersions(secretID string) []*model.SecretVersion {
	return []*model.SecretVersion{}
}

func (m *MockStorage) CreateSecretVersion(version *model.SecretVersion) error {
	return nil
}

func (m *MockStorage) GetSecretVersionByNumber(secretID string, versionNumber int) (*model.SecretVersion, error) {
	return nil, fmt.Errorf("version not found")
}

func TestSecretServiceCreate(t *testing.T) {
	// Создаем мок хранилища
	mockStorage := NewMockStorage()

	// Создаем сервис с моком, используя InitSecretService
	secretService := service.InitSecretService(mockStorage)

	// Тестовые данные
	textData := &model.TextData{
		Text:  "Secret content",
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

	// Тестируем создание секрета
	secret, err := secretService.Create(context.Background(), "user123", req)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	if secret == nil {
		t.Fatal("Expected secret to be created, got nil")
	}

	if secret.UserID != "user123" {
		t.Errorf("Expected user ID to be 'user123', got '%s'", secret.UserID)
	}

	if secret.Title != "Test Secret" {
		t.Errorf("Expected title to be 'Test Secret', got '%s'", secret.Title)
	}

	if secret.Type != model.SecretTypeText {
		t.Errorf("Expected type to be SecretTypeText, got %v", secret.Type)
	}

	// Проверяем, что секрет сохранен в репозитории
	if len(mockStorage.secrets) != 1 {
		t.Errorf("Expected 1 secret in repository, got %d", len(mockStorage.secrets))
	}
}

func TestSecretServiceGetByID(t *testing.T) {
	mockRepo := NewMockSecretRepository()
	secretService := &service.SecretServiceImpl{
		SecretRepo: mockRepo,
	}

	// Создаем тестовый секрет
	testSecret := &model.Secret{
		ID:          "secret123",
		UserID:      "user123",
		Type:        model.SecretTypeLoginPassword,
		Title:       "Test Secret",
		Description: "Test Description",
		Data: &model.LoginPasswordData{
			Login:    "testuser",
			Password: "testpass",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	// Сохраняем секрет в мок репозитории
	mockRepo.Create(context.Background(), testSecret)

	// Тестируем получение секрета
	secret, err := secretService.GetByID(context.Background(), "user123", "secret123")
	if err != nil {
		t.Fatalf("Failed to get secret: %v", err)
	}

	if secret.ID != "secret123" {
		t.Errorf("Expected secret ID to be 'secret123', got '%s'", secret.ID)
	}

	// Тестируем доступ запрещен для другого пользователя
	_, err = secretService.GetByID(context.Background(), "user456", "secret123")
	if err == nil {
		t.Error("Expected access denied error, got nil")
	}
}

func TestSecretServiceGetAll(t *testing.T) {
	mockRepo := NewMockSecretRepository()
	secretService := &service.SecretServiceImpl{
		SecretRepo: mockRepo,
	}

	// Создаем несколько тестовых секретов
	secrets := []*model.Secret{
		{
			ID:        "secret1",
			UserID:    "user123",
			Type:      model.SecretTypeText,
			Title:     "Secret 1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
		{
			ID:        "secret2",
			UserID:    "user123",
			Type:      model.SecretTypeCard,
			Title:     "Secret 2",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
	}

	// Сохраняем секреты
	for _, secret := range secrets {
		mockRepo.Create(context.Background(), secret)
	}

	// Тестируем получение всех секретов
	allSecrets, err := secretService.GetAll(context.Background(), "user123")
	if err != nil {
		t.Fatalf("Failed to get all secrets: %v", err)
	}

	if len(allSecrets) != 2 {
		t.Errorf("Expected 2 secrets, got %d", len(allSecrets))
	}

	// Тестируем получение секретов для другого пользователя
	otherSecrets, err := secretService.GetAll(context.Background(), "user456")
	if err != nil {
		t.Fatalf("Failed to get secrets for other user: %v", err)
	}

	if len(otherSecrets) != 0 {
		t.Errorf("Expected 0 secrets for other user, got %d", len(otherSecrets))
	}
}

func TestSecretServiceSearchByTitle(t *testing.T) {
	mockRepo := NewMockSecretRepository()
	secretService := &service.SecretServiceImpl{
		SecretRepo: mockRepo,
	}

	// Создаем тестовые секреты
	secrets := []*model.Secret{
		{
			ID:        "secret1",
			UserID:    "user123",
			Type:      model.SecretTypeText,
			Title:     "GitHub Account",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
		{
			ID:        "secret2",
			UserID:    "user123",
			Type:      model.SecretTypeText,
			Title:     "Email Password",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
	}

	// Сохраняем секреты
	for _, secret := range secrets {
		mockRepo.Create(context.Background(), secret)
	}

	// Тестируем поиск по названию
	results, err := secretService.SearchByTitle(context.Background(), "user123", "GitHub")
	if err != nil {
		t.Fatalf("Failed to search secrets: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 search result, got %d", len(results))
	}

	if results[0].Title != "GitHub Account" {
		t.Errorf("Expected title 'GitHub Account', got '%s'", results[0].Title)
	}
}

func TestSecretServiceDelete(t *testing.T) {
	mockRepo := NewMockSecretRepository()
	secretService := &service.SecretServiceImpl{
		SecretRepo: mockRepo,
	}

	// Создаем тестовый секрет
	testSecret := &model.Secret{
		ID:        "secret123",
		UserID:    "user123",
		Type:      model.SecretTypeText,
		Title:     "Test Secret",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	// Сохраняем секрет
	mockRepo.Create(context.Background(), testSecret)

	// Тестируем удаление секрета
	err := secretService.Delete(context.Background(), "user123", "secret123")
	if err != nil {
		t.Fatalf("Failed to delete secret: %v", err)
	}

	// Проверяем, что секрет помечен как удаленный
	secret, err := mockRepo.GetByID(context.Background(), "secret123")
	if err != nil {
		t.Fatalf("Failed to get deleted secret: %v", err)
	}

	if secret.DeletedAt == nil {
		t.Error("Expected secret to be marked as deleted")
	}

	// Тестируем удаление секрета другого пользователя
	err = secretService.Delete(context.Background(), "user456", "secret123")
	if err == nil {
		t.Error("Expected access denied error, got nil")
	}
}

func TestSecretServiceGetByType(t *testing.T) {
	mockRepo := NewMockSecretRepository()
	secretService := &service.SecretServiceImpl{
		SecretRepo: mockRepo,
	}

	// Создаем тестовые секреты разных типов
	secrets := []*model.Secret{
		{
			ID:        "secret1",
			UserID:    "user123",
			Type:      model.SecretTypeText,
			Title:     "Text Secret",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
		{
			ID:        "secret2",
			UserID:    "user123",
			Type:      model.SecretTypeCard,
			Title:     "Card Secret",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
		{
			ID:        "secret3",
			UserID:    "user123",
			Type:      model.SecretTypeText,
			Title:     "Another Text Secret",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
	}

	// Сохраняем секреты
	for _, secret := range secrets {
		mockRepo.Create(context.Background(), secret)
	}

	// Тестируем получение секретов определенного типа
	textSecrets, err := secretService.GetByType(context.Background(), "user123", model.SecretTypeText)
	if err != nil {
		t.Fatalf("Failed to get secrets by type: %v", err)
	}

	if len(textSecrets) != 2 {
		t.Errorf("Expected 2 text secrets, got %d", len(textSecrets))
	}

	// Проверяем, что все секреты правильного типа
	for _, secret := range textSecrets {
		if secret.Type != model.SecretTypeText {
			t.Errorf("Expected type SecretTypeText, got %v", secret.Type)
		}
	}
}

func TestSecretServiceUpdate(t *testing.T) {
	mockRepo := NewMockSecretRepository()
	secretService := &service.SecretServiceImpl{
		SecretRepo: mockRepo,
	}

	// Создаем тестовый секрет
	testSecret := &model.Secret{
		ID:          "secret123",
		UserID:      "user123",
		Type:        model.SecretTypeText,
		Title:       "Original Title",
		Description: "Original Description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Version:     1,
	}

	// Сохраняем секрет
	mockRepo.Create(context.Background(), testSecret)

	// Тестируем обновление секрета
	updateReq := &model.UpdateSecretRequest{
		Title:       stringPtr("Updated Title"),
		Description: stringPtr("Updated Description"),
		Tags:        []string{"updated", "test"},
	}

	updatedSecret, err := secretService.Update(context.Background(), "user123", "secret123", updateReq)
	if err != nil {
		t.Fatalf("Failed to update secret: %v", err)
	}

	if updatedSecret.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", updatedSecret.Title)
	}

	if updatedSecret.Version != 2 {
		t.Errorf("Expected version 2, got %d", updatedSecret.Version)
	}

	// Тестируем обновление секрета другого пользователя
	_, err = secretService.Update(context.Background(), "user456", "secret123", updateReq)
	if err == nil {
		t.Error("Expected access denied error, got nil")
	}
}

func TestSecretServiceSearchByTags(t *testing.T) {
	mockRepo := NewMockSecretRepository()
	secretService := &service.SecretServiceImpl{
		SecretRepo: mockRepo,
	}

	// Создаем тестовые секреты с тегами
	secrets := []*model.Secret{
		{
			ID:        "secret1",
			UserID:    "user123",
			Type:      model.SecretTypeText,
			Title:     "Secret 1",
			Tags:      []string{"important", "work"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
		{
			ID:        "secret2",
			UserID:    "user123",
			Type:      model.SecretTypeText,
			Title:     "Secret 2",
			Tags:      []string{"personal", "home"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
		{
			ID:        "secret3",
			UserID:    "user123",
			Type:      model.SecretTypeText,
			Title:     "Secret 3",
			Tags:      []string{"important", "personal"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
	}

	// Сохраняем секреты
	for _, secret := range secrets {
		mockRepo.Create(context.Background(), secret)
	}

	// Тестируем поиск по тегам
	results, err := secretService.SearchByTags(context.Background(), "user123", []string{"important"})
	if err != nil {
		t.Fatalf("Failed to search secrets by tags: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 search results, got %d", len(results))
	}

	// Проверяем, что найдены правильные секреты
	titles := make(map[string]bool)
	for _, secret := range results {
		titles[secret.Title] = true
	}

	if !titles["Secret 1"] {
		t.Error("Expected to find 'Secret 1'")
	}
	if !titles["Secret 3"] {
		t.Error("Expected to find 'Secret 3'")
	}
}

func TestSecretServiceGetVersions(t *testing.T) {
	mockRepo := NewMockSecretRepository()
	secretService := &service.SecretServiceImpl{
		SecretRepo: mockRepo,
	}

	// Создаем тестовый секрет
	testSecret := &model.Secret{
		ID:        "secret123",
		UserID:    "user123",
		Type:      model.SecretTypeText,
		Title:     "Test Secret",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	// Сохраняем секрет
	mockRepo.Create(context.Background(), testSecret)

	// Тестируем получение версий секрета
	versions, err := secretService.GetVersions(context.Background(), "user123", "secret123")
	if err != nil {
		t.Fatalf("Failed to get secret versions: %v", err)
	}

	if len(versions) != 1 {
		t.Errorf("Expected 1 version, got %d", len(versions))
	}

	// Тестируем получение версий секрета другого пользователя
	_, err = secretService.GetVersions(context.Background(), "user456", "secret123")
	if err == nil {
		t.Error("Expected access denied error, got nil")
	}
}

func TestSecretServiceRestoreVersion(t *testing.T) {
	mockRepo := NewMockSecretRepository()
	secretService := &service.SecretServiceImpl{
		SecretRepo: mockRepo,
	}

	// Создаем тестовый секрет
	testSecret := &model.Secret{
		ID:        "secret123",
		UserID:    "user123",
		Type:      model.SecretTypeText,
		Title:     "Test Secret",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	// Сохраняем секрет
	mockRepo.Create(context.Background(), testSecret)

	// Тестируем восстановление версии секрета
	_, err := secretService.RestoreVersion(context.Background(), "user123", "secret123", 1)
	if err == nil {
		t.Error("Expected error about versioning not implemented, got nil")
	}

	// Тестируем восстановление версии секрета другого пользователя
	_, err = secretService.RestoreVersion(context.Background(), "user456", "secret123", 1)
	if err == nil {
		t.Error("Expected access denied error, got nil")
	}
}

// stringPtr возвращает указатель на строку
func stringPtr(s string) *string {
	return &s
}

// createTestAuthService создает тестовый сервис аутентификации без зависимости от конфигурации
func createTestAuthService(storage repository.Storage) *service.AuthServiceImpl {
	// Создаем репозитории
	userRepo := repository.NewUserFileRepository(storage)
	sessionRepo := repository.NewSessionFileRepository(storage)

	// Создаем утилиты с тестовыми параметрами
	jwtManager := util.NewJWTManager("test-secret-key", 15*time.Minute, 30*24*time.Hour)
	hasher := util.NewPasswordHasher()

	// Создаем простое шифрование для тестов
	simpleCrypto := service.InitSimpleCrypto("test_keys")

	return &service.AuthServiceImpl{
		UserRepo:     userRepo,
		SessionRepo:  sessionRepo,
		JWTManager:   jwtManager,
		Hasher:       hasher,
		SimpleCrypto: simpleCrypto,
	}
}

// TestDeadlockServiceIntegration проверяет, что сервис не вызывает deadlock
func TestDeadlockServiceIntegration(t *testing.T) {
	// Инициализируем логгер для тестов
	util.InitLogger(util.LoggerCfg{
		Level: log.DebugLevel,
	})

	// Создаем временный файл для тестирования
	tempFile := t.TempDir() + "/test_storage.json"

	// Инициализируем простое шифрование для тестов
	simpleCrypto := service.InitSimpleCrypto(t.TempDir())

	// Инициализируем реальное хранилище
	store, err := storage.InitStorage(tempFile, simpleCrypto)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Инициализируем сервисы с реальным хранилищем
	authService := createTestAuthService(store)

	// Тест 1: Регистрация пользователя через сервис
	t.Run("ServiceRegister_NoDeadlock", func(t *testing.T) {
		req := &model.UserRegistration{
			Username: "serviceuser1",
			Email:    "service1@example.com",
			Password: "password123",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		done := make(chan interface{}, 1)
		go func() {
			result, err := authService.Register(ctx, req)
			if err != nil {
				done <- err
			} else {
				done <- result
			}
		}()

		select {
		case result := <-done:
			if err, ok := result.(error); ok {
				t.Errorf("Service.Register failed: %v", err)
			}
		case <-ctx.Done():
			t.Fatal("Service.Register hanged - potential deadlock detected!")
		}
	})

	// Тест 2: Попытка регистрации с дублирующимся username
	t.Run("ServiceRegister_DuplicateUsername_NoDeadlock", func(t *testing.T) {
		req := &model.UserRegistration{
			Username: "serviceuser1", // Дублирующийся username
			Email:    "service2@example.com",
			Password: "password123",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		done := make(chan interface{}, 1)
		go func() {
			result, err := authService.Register(ctx, req)
			if err != nil {
				done <- err
			} else {
				done <- result
			}
		}()

		select {
		case result := <-done:
			if err, ok := result.(error); ok {
				// Ожидаем ошибку для дублирующегося username
				if !strings.Contains(err.Error(), "username already exists") {
					t.Errorf("Expected 'username already exists' error, got: %v", err)
				}
			} else {
				t.Error("Expected error for duplicate username, but got success")
			}
		case <-ctx.Done():
			t.Fatal("Service.Register with duplicate username hanged - potential deadlock detected!")
		}
	})

	// Тест 3: Попытка регистрации с дублирующимся email
	t.Run("ServiceRegister_DuplicateEmail_NoDeadlock", func(t *testing.T) {
		req := &model.UserRegistration{
			Username: "serviceuser2",
			Email:    "service1@example.com", // Дублирующийся email
			Password: "password123",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		done := make(chan interface{}, 1)
		go func() {
			result, err := authService.Register(ctx, req)
			if err != nil {
				done <- err
			} else {
				done <- result
			}
		}()

		select {
		case result := <-done:
			if err, ok := result.(error); ok {
				// Ожидаем ошибку для дублирующегося email
				if !strings.Contains(err.Error(), "email already exists") {
					t.Errorf("Expected 'email already exists' error, got: %v", err)
				}
			} else {
				t.Error("Expected error for duplicate email, but got success")
			}
		case <-ctx.Done():
			t.Fatal("Service.Register with duplicate email hanged - potential deadlock detected!")
		}
	})

	// Тест 4: Логин пользователя
	t.Run("ServiceLogin_NoDeadlock", func(t *testing.T) {
		req := &model.UserLogin{
			Username: "serviceuser1",
			Password: "password123",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		done := make(chan interface{}, 1)
		go func() {
			result, err := authService.Login(ctx, req, "test-agent", "127.0.0.1")
			if err != nil {
				done <- err
			} else {
				done <- result
			}
		}()

		select {
		case result := <-done:
			if err, ok := result.(error); ok {
				t.Errorf("Service.Login failed: %v", err)
			}
		case <-ctx.Done():
			t.Fatal("Service.Login hanged - potential deadlock detected!")
		}
	})
}

// TestDeadlockConcurrentServiceCalls проверяет конкурентные вызовы сервиса
func TestDeadlockConcurrentServiceCalls(t *testing.T) {
	// Инициализируем логгер для тестов
	util.InitLogger(util.LoggerCfg{
		Level: log.DebugLevel,
	})

	tempFile := t.TempDir() + "/test_storage.json"

	// Инициализируем простое шифрование для тестов
	simpleCrypto := service.InitSimpleCrypto(t.TempDir())

	store, err := storage.InitStorage(tempFile, simpleCrypto)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	authService := createTestAuthService(store)

	// Создаем несколько горутин для конкурентной регистрации
	const numGoroutines = 5
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			req := &model.UserRegistration{
				Username: fmt.Sprintf("concurrentuser%d", id),
				Email:    fmt.Sprintf("concurrent%d@example.com", id),
				Password: "password123",
			}

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			if _, err := authService.Register(ctx, req); err != nil {
				// Ошибки "email already exists" ожидаемы в конкурентных тестах
				if !strings.Contains(err.Error(), "email already exists") {
					errors <- fmt.Errorf("goroutine %d failed with unexpected error: %w", id, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Проверяем только неожиданные ошибки
	for err := range errors {
		t.Errorf("Concurrent service call error: %v", err)
	}
}

// TestDeadlockServiceWithRealStorage проверяет, что сервис работает с реальным хранилищем
func TestDeadlockServiceWithRealStorage(t *testing.T) {
	tempFile := t.TempDir() + "/test_storage.json"

	// Инициализируем простое шифрование для тестов
	simpleCrypto := service.InitSimpleCrypto(t.TempDir())

	store, err := storage.InitStorage(tempFile, simpleCrypto)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	authService := createTestAuthService(store)

	// Тест полного цикла: регистрация -> логин -> создание секрета
	t.Run("FullCycle_NoDeadlock", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		// Шаг 1: Регистрация
		registerReq := &model.UserRegistration{
			Username: "cycleuser",
			Email:    "cycle@example.com",
			Password: "password123",
		}

		registerResult, err := authService.Register(ctx, registerReq)
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		// Шаг 2: Логин
		loginReq := &model.UserLogin{
			Username: "cycleuser",
			Password: "password123",
		}

		loginResult, err := authService.Login(ctx, loginReq, "test-agent", "127.0.0.1")
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}

		// Проверяем, что получили токены
		if registerResult.TokenPair.AccessToken == "" {
			t.Error("Registration result missing access token")
		}
		if loginResult.TokenPair.AccessToken == "" {
			t.Error("Login result missing access token")
		}

		// Шаг 3: Проверка валидации токена
		claims, err := authService.ValidateToken(ctx, loginResult.TokenPair.AccessToken)
		if err != nil {
			t.Fatalf("Token validation failed: %v", err)
		}

		if claims.UserID != registerResult.User.ID {
			t.Errorf("Expected user ID %s, got %s", registerResult.User.ID, claims.UserID)
		}
		if claims.Username != "cycleuser" {
			t.Errorf("Expected username 'cycleuser', got %s", claims.Username)
		}
	})
}

// createTestSecretService создает тестовый сервис секретов без зависимости от конфигурации
func createTestSecretService(storage repository.Storage) *service.SecretServiceImpl {
	return service.InitSecretService(storage)
}

// TestSecretVersioning проверяет функционал версионирования секретов
func TestSecretVersioning(t *testing.T) {
	// Создаем временный файл для тестирования
	tempFile := t.TempDir() + "/test_storage.json"

	// Инициализируем простое шифрование для тестов
	simpleCrypto := service.InitSimpleCrypto(t.TempDir())

	// Инициализируем реальное хранилище
	store, err := storage.InitStorage(tempFile, simpleCrypto)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Инициализируем сервисы
	secretService := createTestSecretService(store)

	// Тест 1: Создание секрета и получение версий
	t.Run("CreateSecretAndGetVersions", func(t *testing.T) {
		ctx := context.Background()
		userID := "testuser123"

		// Создаем секрет
		textData := &model.TextData{
			Text:  "This is a test secret",
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
			Tags:        []string{"test", "demo"},
			Data:        dataJSON,
		}

		t.Logf("Creating secret with title: %s", req.Title)
		secret, err := secretService.Create(ctx, userID, req)
		if err != nil {
			t.Fatalf("Failed to create secret: %v", err)
		}

		t.Logf("Created secret with ID: %s", secret.ID)

		// Получаем версии секрета
		versions, err := secretService.GetVersions(ctx, userID, secret.ID)
		if err != nil {
			t.Fatalf("Failed to get versions: %v", err)
		}

		if len(versions) != 1 {
			t.Errorf("Expected 1 version, got %d", len(versions))
		}

		version := versions[0]
		if version.SecretID != secret.ID {
			t.Errorf("Expected SecretID %s, got %s", secret.ID, version.SecretID)
		}
		if version.Version != 1 {
			t.Errorf("Expected Version 1, got %d", version.Version)
		}
		if version.Title != "Test Secret" {
			t.Errorf("Expected Title 'Test Secret', got %s", version.Title)
		}
	})

	// Тест 2: Обновление секрета и создание новой версии
	t.Run("UpdateSecretAndCreateVersion", func(t *testing.T) {
		ctx := context.Background()
		userID := "testuser123"

		// Создаем секрет
		textData := &model.TextData{
			Text:  "Update Test Secret",
			Notes: "Original notes",
		}

		// Сериализуем данные в JSON
		dataJSON, err := json.Marshal(textData)
		if err != nil {
			t.Fatalf("Failed to marshal text data: %v", err)
		}

		req := &model.CreateSecretRequest{
			Type:        model.SecretTypeText,
			Title:       "Update Test Secret",
			Description: "Original Description",
			Tags:        []string{"test", "update"},
			Data:        dataJSON,
		}

		secret, err := secretService.Create(ctx, userID, req)
		if err != nil {
			t.Fatalf("Failed to create secret: %v", err)
		}

		// Обновляем секрет
		updateTextData := &model.TextData{
			Text:  "Updated text",
			Notes: "Updated notes",
		}

		updateReq := &model.UpdateSecretRequest{
			Title: stringPtr("Updated Test Secret"),
			Data:  updateTextData,
		}

		updatedSecret, err := secretService.Update(ctx, userID, secret.ID, updateReq)
		if err != nil {
			t.Fatalf("Failed to update secret: %v", err)
		}

		if updatedSecret.Version != 2 {
			t.Errorf("Expected Version 2, got %d", updatedSecret.Version)
		}

		// Получаем версии секрета
		t.Logf("Getting versions after update for secret ID: %s", secret.ID)
		versions, err := secretService.GetVersions(ctx, userID, secret.ID)
		if err != nil {
			t.Fatalf("Failed to get versions: %v", err)
		}

		t.Logf("Found %d versions", len(versions))
		for i, v := range versions {
			t.Logf("Version %d: ID=%s, Version=%d, Title=%s", i, v.ID, v.Version, v.Title)
		}

		if len(versions) != 2 {
			t.Errorf("Expected 2 versions, got %d", len(versions))
		}

		// Проверяем, что есть версии 1 и 2
		hasVersion1 := false
		hasVersion2 := false
		for _, version := range versions {
			if version.Version == 1 {
				hasVersion1 = true
			}
			if version.Version == 2 {
				hasVersion2 = true
			}
		}

		if !hasVersion1 {
			t.Error("Expected to find version 1")
		}
		if !hasVersion2 {
			t.Error("Expected to find version 2")
		}
	})

	// Тест 3: Восстановление версии секрета
	t.Run("RestoreSecretVersion", func(t *testing.T) {
		ctx := context.Background()
		userID := "testuser123"

		// Создаем секрет
		textData := &model.TextData{
			Text:  "Original text",
			Notes: "Original notes",
		}

		// Сериализуем данные в JSON
		dataJSON, err := json.Marshal(textData)
		if err != nil {
			t.Fatalf("Failed to marshal text data: %v", err)
		}

		req := &model.CreateSecretRequest{
			Type:        model.SecretTypeText,
			Title:       "Restore Test Secret",
			Description: "Original Description",
			Tags:        []string{"test", "restore"},
			Data:        dataJSON,
		}

		secret, err := secretService.Create(ctx, userID, req)
		if err != nil {
			t.Fatalf("Failed to create secret: %v", err)
		}

		// Обновляем секрет
		updateReq := &model.UpdateSecretRequest{
			Title: stringPtr("Changed Title"),
			Data: &model.TextData{
				Text:  "Changed text",
				Notes: "Changed notes",
			},
		}

		_, err = secretService.Update(ctx, userID, secret.ID, updateReq)
		if err != nil {
			t.Fatalf("Failed to update secret: %v", err)
		}

		// Восстанавливаем первую версию
		restoredSecret, err := secretService.RestoreVersion(ctx, userID, secret.ID, 1)
		if err != nil {
			t.Fatalf("Failed to restore version: %v", err)
		}

		if restoredSecret.Version != 3 {
			t.Errorf("Expected restored version 3, got %d", restoredSecret.Version)
		}
		if restoredSecret.Title != "Restore Test Secret" {
			t.Errorf("Expected restored title 'Restore Test Secret', got %s", restoredSecret.Title)
		}

		// Проверяем, что создалась новая версия
		versions, err := secretService.GetVersions(ctx, userID, secret.ID)
		if err != nil {
			t.Fatalf("Failed to get versions: %v", err)
		}

		if len(versions) != 3 {
			t.Errorf("Expected 3 versions after restore, got %d", len(versions))
		}
	})

	// Тест 4: Проверка прав доступа к версиям
	t.Run("VersionAccessControl", func(t *testing.T) {
		ctx := context.Background()
		userID1 := "user1"
		userID2 := "user2"

		// Создаем секрет для первого пользователя
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
			Title:       "Access Control Test",
			Description: "Test Description",
			Tags:        []string{"test", "access"},
			Data:        dataJSON,
		}

		secret, err := secretService.Create(ctx, userID1, req)
		if err != nil {
			t.Fatalf("Failed to create secret: %v", err)
		}

		// Второй пользователь пытается получить версии
		_, err = secretService.GetVersions(ctx, userID2, secret.ID)
		if err == nil {
			t.Error("Expected error when accessing another user's secret versions")
		}
		if err.Error() != "access denied: secret does not belong to user" {
			t.Errorf("Expected access denied error, got: %v", err)
		}

		// Второй пользователь пытается восстановить версию
		_, err = secretService.RestoreVersion(ctx, userID2, secret.ID, 1)
		if err == nil {
			t.Error("Expected error when restoring another user's secret version")
		}
		if err.Error() != "access denied: secret does not belong to user" {
			t.Errorf("Expected access denied error, got: %v", err)
		}
	})
}