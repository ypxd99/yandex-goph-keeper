package service

import (
	"context"

	"github.com/ypxd99/yandex-gophkeeper/internal/model"
)

// AuthService определяет интерфейс для аутентификации
type AuthService interface {
	// Register регистрирует нового пользователя
	Register(ctx context.Context, req *model.UserRegistration) (*model.AuthResponse, error)

	// Login выполняет вход пользователя
	Login(ctx context.Context, req *model.UserLogin, userAgent, ipAddress string) (*model.AuthResponse, error)

	// Refresh обновляет токены доступа
	Refresh(ctx context.Context, req *model.RefreshTokenRequest) (*model.TokenPair, error)

	// Logout выполняет выход пользователя
	Logout(ctx context.Context, refreshToken string) error

	// LogoutAll выполняет выход из всех устройств
	LogoutAll(ctx context.Context, userID string) error

	// ChangePassword меняет пароль пользователя
	ChangePassword(ctx context.Context, userID string, req *model.PasswordChangeRequest) error

	// ValidateToken проверяет валидность токена
	ValidateToken(ctx context.Context, token string) (*model.TokenClaims, error)

	// GetUserSessions возвращает активные сессии пользователя
	GetUserSessions(ctx context.Context, userID string) ([]*model.SessionResponse, error)

	// RevokeSession отзывает конкретную сессию
	RevokeSession(ctx context.Context, userID, sessionID string) error
}

// SecretService определяет интерфейс для управления секретами
type SecretService interface {
	// Create создает новый секрет
	Create(ctx context.Context, userID string, req *model.CreateSecretRequest) (*model.Secret, error)

	// GetByID возвращает секрет по ID
	GetByID(ctx context.Context, userID, secretID string) (*model.Secret, error)

	// GetAll возвращает все секреты пользователя
	GetAll(ctx context.Context, userID string) ([]*model.Secret, error)

	// GetByType возвращает секреты определенного типа
	GetByType(ctx context.Context, userID string, secretType model.SecretType) ([]*model.Secret, error)

	// Update обновляет секрет
	Update(ctx context.Context, userID, secretID string, req *model.UpdateSecretRequest) (*model.Secret, error)

	// Delete удаляет секрет
	Delete(ctx context.Context, userID, secretID string) error

	// SearchByTags ищет секреты по тегам
	SearchByTags(ctx context.Context, userID string, tags []string) ([]*model.Secret, error)

	// SearchByTitle ищет секреты по названию
	SearchByTitle(ctx context.Context, userID string, title string) ([]*model.Secret, error)

	// GetVersions возвращает все версии секрета
	GetVersions(ctx context.Context, userID, secretID string) ([]*model.SecretVersion, error)

	// RestoreVersion восстанавливает определенную версию секрета
	RestoreVersion(ctx context.Context, userID, secretID string, version int) (*model.Secret, error)
}
