package tests

import (
	"context"
	"testing"
	"time"

	"github.com/ypxd99/yandex-gophkeeper/internal/model"
	"github.com/ypxd99/yandex-gophkeeper/internal/repository"
)

func TestSessionFileRepositoryCreate(t *testing.T) {
	mockStorage := NewMockStorage()
	sessionRepo := repository.NewSessionFileRepository(mockStorage)

	// Тестовые данные
	session := &model.Session{
		ID:           "session123",
		UserID:       "user123",
		RefreshToken: "refresh_token_123",
		UserAgent:    "Test Browser",
		IPAddress:    "127.0.0.1",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		IsActive:     true,
	}

	// Тестируем создание сессии
	err := sessionRepo.Create(context.Background(), session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Проверяем, что сессия сохранена
	if len(mockStorage.sessions) != 1 {
		t.Errorf("Expected 1 session in storage, got %d", len(mockStorage.sessions))
	}

	// Проверяем, что сессию можно получить по ID
	savedSession, err := sessionRepo.GetByID(context.Background(), "session123")
	if err != nil {
		t.Fatalf("Failed to get session by ID: %v", err)
	}

	if savedSession.UserID != "user123" {
		t.Errorf("Expected user ID 'user123', got '%s'", savedSession.UserID)
	}
}

func TestSessionFileRepositoryGetByUserID(t *testing.T) {
	mockStorage := NewMockStorage()
	sessionRepo := repository.NewSessionFileRepository(mockStorage)

	// Создаем несколько тестовых сессий для одного пользователя
	sessions := []*model.Session{
		{
			ID:           "session1",
			UserID:       "user123",
			RefreshToken: "token1",
			UserAgent:    "Browser 1",
			IPAddress:    "127.0.0.1",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			IsActive:     true,
		},
		{
			ID:           "session2",
			UserID:       "user123",
			RefreshToken: "token2",
			UserAgent:    "Browser 2",
			IPAddress:    "127.0.0.2",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			IsActive:     true,
		},
		{
			ID:           "session3",
			UserID:       "user456",
			RefreshToken: "token3",
			UserAgent:    "Browser 3",
			IPAddress:    "127.0.0.3",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			IsActive:     true,
		},
	}

	// Сохраняем сессии
	for _, session := range sessions {
		mockStorage.sessions[session.ID] = session
	}

	// Тестируем получение сессий пользователя
	userSessions, err := sessionRepo.GetByUserID(context.Background(), "user123")
	if err != nil {
		t.Fatalf("Failed to get sessions by user ID: %v", err)
	}

	if len(userSessions) != 2 {
		t.Errorf("Expected 2 sessions for user123, got %d", len(userSessions))
	}

	// Проверяем, что все сессии принадлежат правильному пользователю
	for _, session := range userSessions {
		if session.UserID != "user123" {
			t.Errorf("Expected user ID 'user123', got '%s'", session.UserID)
		}
	}
}

func TestSessionFileRepositoryGetByToken(t *testing.T) {
	mockStorage := NewMockStorage()
	sessionRepo := repository.NewSessionFileRepository(mockStorage)

	// Создаем тестовую сессию
	session := &model.Session{
		ID:           "session123",
		UserID:       "user123",
		RefreshToken: "refresh_token_123",
		UserAgent:    "Test Browser",
		IPAddress:    "127.0.0.1",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		IsActive:     true,
	}

	mockStorage.sessions[session.ID] = session

	// Тестируем получение сессии по токену
	foundSession, err := sessionRepo.GetByToken(context.Background(), "refresh_token_123")
	if err != nil {
		t.Fatalf("Failed to get session by token: %v", err)
	}

	if foundSession.ID != "session123" {
		t.Errorf("Expected session ID 'session123', got '%s'", foundSession.ID)
	}

	// Тестируем получение сессии по несуществующему токену
	_, err = sessionRepo.GetByToken(context.Background(), "nonexistent_token")
	if err != model.ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}
}

func TestSessionFileRepositoryUpdate(t *testing.T) {
	mockStorage := NewMockStorage()
	sessionRepo := repository.NewSessionFileRepository(mockStorage)

	// Создаем тестовую сессию
	session := &model.Session{
		ID:           "session123",
		UserID:       "user123",
		RefreshToken: "refresh_token_123",
		UserAgent:    "Original Browser",
		IPAddress:    "127.0.0.1",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		IsActive:     true,
	}

	mockStorage.sessions[session.ID] = session

	// Обновляем сессию
	session.UserAgent = "Updated Browser"
	session.UpdatedAt = time.Now()

	err := sessionRepo.Update(context.Background(), session)
	if err != nil {
		t.Fatalf("Failed to update session: %v", err)
	}

	// Проверяем, что изменения сохранены
	updatedSession, err := sessionRepo.GetByID(context.Background(), "session123")
	if err != nil {
		t.Fatalf("Failed to get updated session: %v", err)
	}

	if updatedSession.UserAgent != "Updated Browser" {
		t.Errorf("Expected user agent 'Updated Browser', got '%s'", updatedSession.UserAgent)
	}
}

func TestSessionFileRepositoryDelete(t *testing.T) {
	mockStorage := NewMockStorage()
	sessionRepo := repository.NewSessionFileRepository(mockStorage)

	// Создаем тестовую сессию
	session := &model.Session{
		ID:           "session123",
		UserID:       "user123",
		RefreshToken: "refresh_token_123",
		UserAgent:    "Test Browser",
		IPAddress:    "127.0.0.1",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		IsActive:     true,
	}

	mockStorage.sessions[session.ID] = session

	// Проверяем, что сессия существует
	if len(mockStorage.sessions) != 1 {
		t.Error("Expected 1 session before deletion")
	}

	// Удаляем сессию
	err := sessionRepo.Delete(context.Background(), "session123")
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Проверяем, что сессия удалена
	if len(mockStorage.sessions) != 0 {
		t.Error("Expected 0 sessions after deletion")
	}
}

func TestSessionFileRepositoryDeleteByUserID(t *testing.T) {
	mockStorage := NewMockStorage()
	sessionRepo := repository.NewSessionFileRepository(mockStorage)

	// Создаем несколько тестовых сессий для одного пользователя
	sessions := []*model.Session{
		{
			ID:           "session1",
			UserID:       "user123",
			RefreshToken: "token1",
			UserAgent:    "Browser 1",
			IPAddress:    "127.0.0.1",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			IsActive:     true,
		},
		{
			ID:           "session2",
			UserID:       "user123",
			RefreshToken: "token2",
			UserAgent:    "Browser 2",
			IPAddress:    "127.0.0.2",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			IsActive:     true,
		},
		{
			ID:           "session3",
			UserID:       "user456",
			RefreshToken: "token3",
			UserAgent:    "Browser 3",
			IPAddress:    "127.0.0.3",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			IsActive:     true,
		},
	}

	// Сохраняем сессии
	for _, session := range sessions {
		mockStorage.sessions[session.ID] = session
	}

	// Проверяем, что у нас есть 3 сессии
	if len(mockStorage.sessions) != 3 {
		t.Error("Expected 3 sessions before deletion")
	}

	// Удаляем все сессии пользователя
	err := sessionRepo.DeleteByUserID(context.Background(), "user123")
	if err != nil {
		t.Fatalf("Failed to delete sessions by user ID: %v", err)
	}

	// Проверяем, что сессии пользователя удалены, а других пользователей остались
	if len(mockStorage.sessions) != 1 {
		t.Errorf("Expected 1 session after deletion, got %d", len(mockStorage.sessions))
	}

	// Проверяем, что осталась только сессия user456
	remainingSession, err := sessionRepo.GetByID(context.Background(), "session3")
	if err != nil {
		t.Fatalf("Failed to get remaining session: %v", err)
	}

	if remainingSession.UserID != "user456" {
		t.Errorf("Expected remaining session to belong to user456, got '%s'", remainingSession.UserID)
	}
}
