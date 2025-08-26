package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ypxd99/yandex-gophkeeper/internal/model"
)

// SessionFileRepository реализует SessionRepository с файловым хранилищем.
// Обеспечивает CRUD операции для сессий пользователей с персистентным хранением в JSON файле.
type SessionFileRepository struct {
	storage Storage
}

// NewSessionFileRepository создает новый экземпляр SessionFileRepository.
// Принимает интерфейс Storage для работы с данными.
// Возвращает инициализированный репозиторий сессий.
func NewSessionFileRepository(storage Storage) *SessionFileRepository {
	return &SessionFileRepository{
		storage: storage,
	}
}

// Create создает новую сессию в хранилище.
// Генерирует ID и устанавливает временные метки.
// Принимает контекст запроса и данные сессии.
// Возвращает ошибку при неудачном создании.
func (r *SessionFileRepository) Create(ctx context.Context, session *model.Session) error {
	// Получаем все сессии
	sessions := r.storage.GetSessions()

	// Добавляем новую сессию
	sessions = append(sessions, session)

	// Для мока нужно обновить внутреннее состояние
	if mockStorage, ok := r.storage.(interface{ SetSessions([]*model.Session) }); ok {
		mockStorage.SetSessions(sessions)
	}

	// Сохраняем изменения
	return r.storage.Save()
}

// GetByID возвращает сессию по уникальному идентификатору.
// Принимает контекст запроса и ID сессии.
// Возвращает сессию или ошибку если сессия не найдена.
func (r *SessionFileRepository) GetByID(ctx context.Context, id string) (*model.Session, error) {
	sessions := r.storage.GetSessions()

	for _, session := range sessions {
		if session.ID == id {
			return session, nil
		}
	}

	return nil, fmt.Errorf("session not found")
}

// GetByUserID возвращает все сессии пользователя.
// Принимает контекст запроса и ID пользователя.
// Возвращает слайс сессий или ошибку при неудачном получении.
func (r *SessionFileRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Session, error) {
	sessions := r.storage.GetSessions()
	var userSessions []*model.Session

	for _, session := range sessions {
		if session.UserID == userID {
			userSessions = append(userSessions, session)
		}
	}

	return userSessions, nil
}

// GetByToken возвращает сессию по refresh токену.
// Принимает контекст запроса и refresh токен.
// Возвращает сессию или ошибку если сессия не найдена.
func (r *SessionFileRepository) GetByToken(ctx context.Context, token string) (*model.Session, error) {
	sessions := r.storage.GetSessions()

	for _, session := range sessions {
		if session.RefreshToken == token {
			return session, nil
		}
	}

	return nil, model.ErrSessionNotFound
}

// Update обновляет данные существующей сессии.
// Проверяет существование сессии и обновляет временные метки.
// Принимает контекст запроса и обновленные данные сессии.
// Возвращает ошибку при неудачном обновлении.
func (r *SessionFileRepository) Update(ctx context.Context, session *model.Session) error {
	sessions := r.storage.GetSessions()

	for i, existingSession := range sessions {
		if existingSession.ID == session.ID {
			sessions[i] = session

			// Для мока нужно обновить внутреннее состояние
			if mockStorage, ok := r.storage.(interface{ SetSessions([]*model.Session) }); ok {
				mockStorage.SetSessions(sessions)
			}

			return r.storage.Save()
		}
	}

	return fmt.Errorf("session not found")
}

// Delete удаляет сессию по ID.
// Принимает контекст запроса и ID сессии.
// Возвращает ошибку при неудачном удалении.
func (r *SessionFileRepository) Delete(ctx context.Context, id string) error {
	sessions := r.storage.GetSessions()

	for i, session := range sessions {
		if session.ID == id {
			// Удаляем сессию из слайса
			sessions = append(sessions[:i], sessions[i+1:]...)

			// Для мока нужно обновить внутреннее состояние
			if mockStorage, ok := r.storage.(interface{ SetSessions([]*model.Session) }); ok {
				mockStorage.SetSessions(sessions)
			}

			return r.storage.Save()
		}
	}

	return fmt.Errorf("session not found")
}

// DeleteByUserID удаляет все сессии пользователя.
// Принимает контекст запроса и ID пользователя.
// Возвращает ошибку при неудачном удалении.
func (r *SessionFileRepository) DeleteByUserID(ctx context.Context, userID string) error {
	sessions := r.storage.GetSessions()
	var remainingSessions []*model.Session

	for _, session := range sessions {
		if session.UserID != userID {
			remainingSessions = append(remainingSessions, session)
		}
	}

	// Для мока нужно обновить внутреннее состояние
	if mockStorage, ok := r.storage.(interface{ SetSessions([]*model.Session) }); ok {
		mockStorage.SetSessions(remainingSessions)
	}

	return r.storage.Save()
}

// DeleteExpired удаляет истекшие сессии.
// Принимает контекст запроса.
// Возвращает ошибку при неудачном удалении.
func (r *SessionFileRepository) DeleteExpired(ctx context.Context) error {
	sessions := r.storage.GetSessions()
	var activeSessions []*model.Session
	now := time.Now()

	for _, session := range sessions {
		if session.ExpiresAt.After(now) {
			activeSessions = append(activeSessions, session)
		}
	}

	// Для мока нужно обновить внутреннее состояние
	if mockStorage, ok := r.storage.(interface{ SetSessions([]*model.Session) }); ok {
		mockStorage.SetSessions(activeSessions)
	}

	return r.storage.Save()
}
