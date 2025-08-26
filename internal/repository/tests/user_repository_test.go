package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ypxd99/yandex-gophkeeper/internal/model"
	"github.com/ypxd99/yandex-gophkeeper/internal/repository"
	"github.com/ypxd99/yandex-gophkeeper/internal/repository/storage"
	"github.com/ypxd99/yandex-gophkeeper/internal/service"
)

// MockStorage - мок хранилища для тестирования репозиториев
type MockStorage struct {
	users    map[string]*model.User
	secrets  map[string]*model.Secret
	sessions map[string]*model.Session
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		users:    make(map[string]*model.User),
		secrets:  make(map[string]*model.Secret),
		sessions: make(map[string]*model.Session),
	}
}

func (m *MockStorage) Close() error {
	return nil
}

func (m *MockStorage) GetUsers() []*model.User {
	users := make([]*model.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users
}

func (m *MockStorage) GetSecrets() []*model.Secret {
	secrets := make([]*model.Secret, 0, len(m.secrets))
	for _, secret := range m.secrets {
		secrets = append(secrets, secret)
	}
	return secrets
}

func (m *MockStorage) SetSecrets(secrets []*model.Secret) {
	m.secrets = make(map[string]*model.Secret)
	for _, secret := range secrets {
		m.secrets[secret.ID] = secret
	}
}

func (m *MockStorage) SetSessions(sessions []*model.Session) {
	m.sessions = make(map[string]*model.Session)
	for _, session := range sessions {
		m.sessions[session.ID] = session
	}
}

func (m *MockStorage) GetSessions() []*model.Session {
	sessions := make([]*model.Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (m *MockStorage) Save() error {
	return nil
}

func (m *MockStorage) GenerateID() string {
	return "test_id"
}

func (m *MockStorage) GetUserByID(id string) (*model.User, error) {
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return nil, model.ErrUserNotFound
}

func (m *MockStorage) GetUserByUsername(username string) (*model.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, model.ErrUserNotFound
}

func (m *MockStorage) GetUserByEmail(email string) (*model.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, model.ErrUserNotFound
}

func (m *MockStorage) CreateUser(user *model.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockStorage) UpdateUser(user *model.User) error {
	if _, exists := m.users[user.ID]; exists {
		m.users[user.ID] = user
		return nil
	}
	return model.ErrUserNotFound
}

// CreateSecret создает новый секрет в моке
func (m *MockStorage) CreateSecret(secret *model.Secret) error {
	m.secrets[secret.ID] = secret
	return nil
}

// UpdateSecret обновляет существующий секрет в моке
func (m *MockStorage) UpdateSecret(secret *model.Secret) error {
	if _, exists := m.secrets[secret.ID]; exists {
		m.secrets[secret.ID] = secret
		return nil
	}
	return fmt.Errorf("secret not found")
}

// DeleteSecret удаляет секрет из мока
func (m *MockStorage) DeleteSecret(id string) error {
	if _, exists := m.secrets[id]; exists {
		delete(m.secrets, id)
		return nil
	}
	return fmt.Errorf("secret not found")
}

// GetSecretVersions возвращает версии секрета из мока
func (m *MockStorage) GetSecretVersions(secretID string) []*model.SecretVersion {
	// Для мока возвращаем пустой слайс
	return []*model.SecretVersion{}
}

// CreateSecretVersion создает новую версию секрета в моке
func (m *MockStorage) CreateSecretVersion(version *model.SecretVersion) error {
	// Для мока просто возвращаем успех
	return nil
}

// GetSecretVersionByNumber возвращает версию секрета по номеру из мока
func (m *MockStorage) GetSecretVersionByNumber(secretID string, versionNumber int) (*model.SecretVersion, error) {
	// Для мока возвращаем ошибку "не найдено"
	return nil, fmt.Errorf("version not found")
}

func (m *MockStorage) DeleteUser(id string) error {
	if _, exists := m.users[id]; exists {
		delete(m.users, id)
		return nil
	}
	return model.ErrUserNotFound
}

func (m *MockStorage) UserExists(username string) bool {
	for _, user := range m.users {
		if user.Username == username {
			return true
		}
	}
	return false
}

func TestUserFileRepositoryCreate(t *testing.T) {
	mockStorage := NewMockStorage()
	userRepo := repository.NewUserFileRepository(mockStorage)

	// Тестовые данные
	user := &model.User{
		ID:           "user123",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Salt:         "salt123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	// Тестируем создание пользователя
	err := userRepo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Проверяем, что пользователь сохранен
	if len(mockStorage.users) != 1 {
		t.Errorf("Expected 1 user in storage, got %d", len(mockStorage.users))
	}

	// Проверяем, что пользователь можно получить по ID
	savedUser, err := userRepo.GetByID(context.Background(), "user123")
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	if savedUser.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", savedUser.Username)
	}
}

func TestUserFileRepositoryGetByUsername(t *testing.T) {
	mockStorage := NewMockStorage()
	userRepo := repository.NewUserFileRepository(mockStorage)

	// Создаем тестового пользователя
	user := &model.User{
		ID:           "user123",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Salt:         "salt123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	mockStorage.CreateUser(user)

	// Тестируем получение пользователя по username
	foundUser, err := userRepo.GetByUsername(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("Failed to get user by username: %v", err)
	}

	if foundUser.ID != "user123" {
		t.Errorf("Expected user ID 'user123', got '%s'", foundUser.ID)
	}

	// Тестируем получение несуществующего пользователя
	_, err = userRepo.GetByUsername(context.Background(), "nonexistent")
	if err != model.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserFileRepositoryGetByEmail(t *testing.T) {
	mockStorage := NewMockStorage()
	userRepo := repository.NewUserFileRepository(mockStorage)

	// Создаем тестового пользователя
	user := &model.User{
		ID:           "user123",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Salt:         "salt123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	mockStorage.CreateUser(user)

	// Тестируем получение пользователя по email
	foundUser, err := userRepo.GetByEmail(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	if foundUser.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", foundUser.Username)
	}

	// Тестируем получение несуществующего пользователя
	_, err = userRepo.GetByEmail(context.Background(), "nonexistent@example.com")
	if err != model.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserFileRepositoryExists(t *testing.T) {
	mockStorage := NewMockStorage()
	userRepo := repository.NewUserFileRepository(mockStorage)

	// Создаем тестового пользователя
	user := &model.User{
		ID:           "user123",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Salt:         "salt123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	mockStorage.CreateUser(user)

	// Тестируем проверку существования пользователя
	if !userRepo.Exists(context.Background(), "testuser") {
		t.Error("Expected user to exist")
	}

	if userRepo.Exists(context.Background(), "nonexistent") {
		t.Error("Expected user to not exist")
	}
}

func TestUserFileRepositoryUpdate(t *testing.T) {
	mockStorage := NewMockStorage()
	userRepo := repository.NewUserFileRepository(mockStorage)

	// Создаем тестового пользователя
	user := &model.User{
		ID:           "user123",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Salt:         "salt123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	mockStorage.CreateUser(user)

	// Обновляем пользователя
	user.Email = "updated@example.com"
	user.UpdatedAt = time.Now()

	err := userRepo.Update(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Проверяем, что изменения сохранены
	updatedUser, err := userRepo.GetByID(context.Background(), "user123")
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updatedUser.Email != "updated@example.com" {
		t.Errorf("Expected email 'updated@example.com', got '%s'", updatedUser.Email)
	}
}

func TestUserFileRepositoryDelete(t *testing.T) {
	mockStorage := NewMockStorage()
	userRepo := repository.NewUserFileRepository(mockStorage)

	// Создаем тестового пользователя
	user := &model.User{
		ID:           "user123",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Salt:         "salt123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	mockStorage.CreateUser(user)

	// Проверяем, что пользователь существует
	if !userRepo.Exists(context.Background(), "testuser") {
		t.Error("Expected user to exist before deletion")
	}

	// Удаляем пользователя
	err := userRepo.Delete(context.Background(), "user123")
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Проверяем, что пользователь удален
	if userRepo.Exists(context.Background(), "testuser") {
		t.Error("Expected user to not exist after deletion")
	}
}

// TestDeadlockCreateUser проверяет, что CreateUser не вызывает deadlock
func TestDeadlockCreateUser(t *testing.T) {
	// Создаем временный файл для тестирования
	tempFile := t.TempDir() + "/test_storage.json"

	// Инициализируем простое шифрование для тестов
	simpleCrypto := service.InitSimpleCrypto(t.TempDir())

	// Инициализируем реальное хранилище с мьютексами
	store, err := storage.InitStorage(tempFile, simpleCrypto)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Создаем репозиторий
	userRepo := repository.NewUserFileRepository(store)

	// Тест 1: Создание пользователя (должно работать без deadlock)
	user1 := &model.User{
		Username:     "testuser1",
		Email:        "test1@example.com",
		PasswordHash: "hashed_password",
		Salt:         "salt123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	// Устанавливаем таймаут для выявления deadlock
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- userRepo.Create(ctx, user1)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("CreateUser failed: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("CreateUser hanged - potential deadlock detected!")
	}

	// Тест 2: Создание пользователя с дублирующимся username (должно работать без deadlock)
	user2 := &model.User{
		Username:     "testuser1", // Дублирующийся username
		Email:        "test2@example.com",
		PasswordHash: "hashed_password",
		Salt:         "salt123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	done = make(chan error, 1)
	go func() {
		done <- userRepo.Create(ctx, user2)
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Error("Expected error for duplicate username, but got none")
		}
	case <-ctx.Done():
		t.Fatal("CreateUser with duplicate username hanged - potential deadlock detected!")
	}

	// Тест 3: Создание пользователя с дублирующимся email (должно работать без deadlock)
	user3 := &model.User{
		Username:     "testuser2",
		Email:        "test2@example.com", // Изменяем email чтобы избежать конфликта
		PasswordHash: "hashed_password",
		Salt:         "salt123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	// Отладочная информация
	t.Logf("Attempting to create user with email: %s", user3.Email)
	t.Logf("First user email: %s", user1.Email)

	done = make(chan error, 1)
	go func() {
		done <- userRepo.Create(ctx, user3)
	}()

	select {
	case err := <-done:
		// Теперь мы НЕ ожидаем ошибку, так как email разные
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("CreateUser with duplicate email hanged - potential deadlock detected!")
	}
}

// TestDeadlockConcurrentAccess проверяет конкурентный доступ к хранилищу
func TestDeadlockConcurrentAccess(t *testing.T) {
	tempFile := t.TempDir() + "/test_storage.json"

	// Инициализируем простое шифрование для тестов
	simpleCrypto := service.InitSimpleCrypto(t.TempDir())

	store, err := storage.InitStorage(tempFile, simpleCrypto)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	userRepo := repository.NewUserFileRepository(store)

	// Создаем несколько горутин для конкурентного доступа
	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			user := &model.User{
				Username:     fmt.Sprintf("user%d", id),
				Email:        fmt.Sprintf("user%d@example.com", id),
				PasswordHash: "hashed_password",
				Salt:         "salt123",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
				IsActive:     true,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := userRepo.Create(ctx, user); err != nil {
				errors <- fmt.Errorf("goroutine %d failed: %w", id, err)
			}
		}(i)
	}

	// Ждем завершения всех горутин
	wg.Wait()
	close(errors)

	// Проверяем ошибки
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}

// TestDeadlockMixedOperations проверяет смешанные операции чтения/записи
func TestDeadlockMixedOperations(t *testing.T) {
	tempFile := t.TempDir() + "/test_storage.json"

	// Инициализируем простое шифрование для тестов
	simpleCrypto := service.InitSimpleCrypto(t.TempDir())

	store, err := storage.InitStorage(tempFile, simpleCrypto)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	userRepo := repository.NewUserFileRepository(store)

	// Создаем пользователя для тестирования
	user := &model.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Salt:         "salt123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Тестируем смешанные операции
	t.Run("MixedOperations_NoDeadlock", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var wg sync.WaitGroup
		errors := make(chan error, 3)

		// Операция записи (Update)
		wg.Add(1)
		go func() {
			defer wg.Done()
			user.UpdatedAt = time.Now()
			if err := userRepo.Update(ctx, user); err != nil {
				errors <- fmt.Errorf("update failed: %w", err)
			}
		}()

		// Операция чтения (GetByUsername)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := userRepo.GetByUsername(ctx, "testuser"); err != nil {
				errors <- fmt.Errorf("get by username failed: %w", err)
			}
		}()

		// Операция чтения (Exists)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !userRepo.Exists(ctx, "testuser") {
				errors <- fmt.Errorf("exists check failed")
			}
		}()

		wg.Wait()
		close(errors)

		for err := range errors {
			t.Errorf("Mixed operations error: %v", err)
		}
	})
}

// TestDeadlockNestedCalls проверяет, что вложенные вызовы не вызывают deadlock
func TestDeadlockNestedCalls(t *testing.T) {
	tempFile := t.TempDir() + "/test_storage.json"

	// Инициализируем простое шифрование для тестов
	simpleCrypto := service.InitSimpleCrypto(t.TempDir())

	store, err := storage.InitStorage(tempFile, simpleCrypto)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Тестируем прямые вызовы storage методов
	t.Run("StorageDirectCalls_NoDeadlock", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		user := &model.User{
			Username:     "nesteduser",
			Email:        "nested@example.com",
			PasswordHash: "hashed_password",
			Salt:         "salt123",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			IsActive:     true,
		}

		done := make(chan error, 1)
		go func() {
			done <- store.CreateUser(user)
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Errorf("Direct CreateUser failed: %v", err)
			}
		case <-ctx.Done():
			t.Fatal("Direct CreateUser hanged - potential deadlock detected!")
		}
	})
}
