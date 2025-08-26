package tests

import (
	"context"
	"testing"
	"time"

	"github.com/ypxd99/yandex-gophkeeper/internal/model"
	"github.com/ypxd99/yandex-gophkeeper/internal/repository"
)

func TestSecretFileRepositoryCreate(t *testing.T) {
	mockStorage := NewMockStorage()
	secretRepo := repository.NewSecretFileRepository(mockStorage)

	// Тестовые данные
	secret := &model.Secret{
		ID:          "secret123",
		UserID:      "user123",
		Type:        model.SecretTypeText,
		Title:       "Test Secret",
		Description: "Test Description",
		Tags:        []string{"test"},
		Data: &model.TextData{
			Text:  "Secret content",
			Notes: "Test notes",
		},
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Тестируем создание секрета
	err := secretRepo.Create(context.Background(), secret)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	// Проверяем, что секрет сохранен
	if len(mockStorage.secrets) != 1 {
		t.Errorf("Expected 1 secret in storage, got %d", len(mockStorage.secrets))
	}

	// Проверяем, что секрет можно получить по ID
	savedSecret, err := secretRepo.GetByID(context.Background(), "secret123")
	if err != nil {
		t.Fatalf("Failed to get secret by ID: %v", err)
	}

	if savedSecret.Title != "Test Secret" {
		t.Errorf("Expected title 'Test Secret', got '%s'", savedSecret.Title)
	}
}

func TestSecretFileRepositoryGetByUserID(t *testing.T) {
	mockStorage := NewMockStorage()
	secretRepo := repository.NewSecretFileRepository(mockStorage)

	// Создаем несколько тестовых секретов для одного пользователя
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
		{
			ID:        "secret3",
			UserID:    "user456",
			Type:      model.SecretTypeText,
			Title:     "Secret 3",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
	}

	// Сохраняем секреты
	for _, secret := range secrets {
		mockStorage.secrets[secret.ID] = secret
	}

	// Тестируем получение секретов пользователя
	userSecrets, err := secretRepo.GetByUserID(context.Background(), "user123")
	if err != nil {
		t.Fatalf("Failed to get secrets by user ID: %v", err)
	}

	if len(userSecrets) != 2 {
		t.Errorf("Expected 2 secrets for user123, got %d", len(userSecrets))
	}

	// Проверяем, что все секреты принадлежат правильному пользователю
	for _, secret := range userSecrets {
		if secret.UserID != "user123" {
			t.Errorf("Expected user ID 'user123', got '%s'", secret.UserID)
		}
	}
}

func TestSecretFileRepositoryGetByType(t *testing.T) {
	mockStorage := NewMockStorage()
	secretRepo := repository.NewSecretFileRepository(mockStorage)

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
		mockStorage.secrets[secret.ID] = secret
	}

	// Тестируем получение секретов определенного типа
	textSecrets, err := secretRepo.GetByType(context.Background(), "user123", model.SecretTypeText)
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

func TestSecretFileRepositorySearchByTitle(t *testing.T) {
	mockStorage := NewMockStorage()
	secretRepo := repository.NewSecretFileRepository(mockStorage)

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
		{
			ID:        "secret3",
			UserID:    "user123",
			Type:      model.SecretTypeText,
			Title:     "Bank Account",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
	}

	// Сохраняем секреты
	for _, secret := range secrets {
		mockStorage.secrets[secret.ID] = secret
	}

	// Тестируем поиск по названию
	results, err := secretRepo.SearchByTitle(context.Background(), "user123", "Account")
	if err != nil {
		t.Fatalf("Failed to search secrets by title: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 search results, got %d", len(results))
	}

	// Проверяем, что найдены правильные секреты
	titles := make(map[string]bool)
	for _, secret := range results {
		titles[secret.Title] = true
	}

	if !titles["GitHub Account"] {
		t.Error("Expected to find 'GitHub Account'")
	}
	if !titles["Bank Account"] {
		t.Error("Expected to find 'Bank Account'")
	}
}

func TestSecretFileRepositoryUpdate(t *testing.T) {
	mockStorage := NewMockStorage()
	secretRepo := repository.NewSecretFileRepository(mockStorage)

	// Создаем тестовый секрет
	secret := &model.Secret{
		ID:          "secret123",
		UserID:      "user123",
		Type:        model.SecretTypeText,
		Title:       "Original Title",
		Description: "Original Description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Version:     1,
	}

	mockStorage.secrets[secret.ID] = secret

	// Обновляем секрет
	secret.Title = "Updated Title"
	secret.Description = "Updated Description"
	secret.UpdatedAt = time.Now()
	secret.Version = 2

	err := secretRepo.Update(context.Background(), secret)
	if err != nil {
		t.Fatalf("Failed to update secret: %v", err)
	}

	// Проверяем, что изменения сохранены
	updatedSecret, err := secretRepo.GetByID(context.Background(), "secret123")
	if err != nil {
		t.Fatalf("Failed to get updated secret: %v", err)
	}

	if updatedSecret.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", updatedSecret.Title)
	}

	if updatedSecret.Version != 2 {
		t.Errorf("Expected version 2, got %d", updatedSecret.Version)
	}
}

func TestSecretFileRepositoryDelete(t *testing.T) {
	mockStorage := NewMockStorage()
	secretRepo := repository.NewSecretFileRepository(mockStorage)

	// Создаем тестовый секрет
	secret := &model.Secret{
		ID:        "secret123",
		UserID:    "user123",
		Type:      model.SecretTypeText,
		Title:     "Test Secret",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	mockStorage.secrets[secret.ID] = secret

	// Проверяем, что секрет существует
	if len(mockStorage.secrets) != 1 {
		t.Error("Expected 1 secret before deletion")
	}

	// Удаляем секрет
	err := secretRepo.Delete(context.Background(), "secret123")
	if err != nil {
		t.Fatalf("Failed to delete secret: %v", err)
	}

	// Проверяем, что секрет удален
	if len(mockStorage.secrets) != 0 {
		t.Error("Expected 0 secrets after deletion")
	}
}
