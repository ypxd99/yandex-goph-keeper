package repository

import (
	"context"

	"github.com/ypxd99/yandex-gophkeeper/internal/model"
)

// UserRepository определяет интерфейс для работы с пользователями.
// Предоставляет методы для создания, чтения, обновления и удаления пользователей.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, username string) bool
	UpdateLastLogin(ctx context.Context, id string) error
}

// SecretRepository определяет интерфейс для работы с секретами.
// Предоставляет методы для управления секретами пользователей.
type SecretRepository interface {
	Create(ctx context.Context, secret *model.Secret) error
	GetByID(ctx context.Context, id string) (*model.Secret, error)
	GetByUserID(ctx context.Context, userID string) ([]*model.Secret, error)
	GetByType(ctx context.Context, userID string, secretType model.SecretType) ([]*model.Secret, error)
	Update(ctx context.Context, secret *model.Secret) error
	Delete(ctx context.Context, id string) error
	SearchByTitle(ctx context.Context, userID string, query string) ([]*model.Secret, error)
	SearchByTags(ctx context.Context, userID string, tags []string) ([]*model.Secret, error)
	// Методы для версионирования
	GetVersions(ctx context.Context, secretID string) ([]*model.SecretVersion, error)
	CreateVersion(ctx context.Context, version *model.SecretVersion) error
	GetVersionByNumber(ctx context.Context, secretID string, versionNumber int) (*model.SecretVersion, error)
}

// SessionRepository определяет интерфейс для работы с сессиями пользователей.
// Предоставляет методы для управления активными сессиями.
type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	GetByID(ctx context.Context, id string) (*model.Session, error)
	GetByUserID(ctx context.Context, userID string) ([]*model.Session, error)
	GetByToken(ctx context.Context, token string) (*model.Session, error)
	Update(ctx context.Context, session *model.Session) error
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}

// Storage определяет интерфейс для файлового хранилища.
// Предоставляет методы для работы с данными и метаданными хранилища.
type Storage interface {
	Close() error
	GetUsers() []*model.User
	GetSecrets() []*model.Secret
	GetSessions() []*model.Session
	Save() error
	GenerateID() string
	GetUserByID(id string) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	GetUserByEmail(email string) (*model.User, error)
	CreateUser(user *model.User) error
	UpdateUser(user *model.User) error
	DeleteUser(id string) error
	UserExists(username string) bool
	// Методы для работы с секретами
	CreateSecret(secret *model.Secret) error
	UpdateSecret(secret *model.Secret) error
	DeleteSecret(id string) error
	// Методы для версионирования секретов
	GetSecretVersions(secretID string) []*model.SecretVersion
	CreateSecretVersion(version *model.SecretVersion) error
	GetSecretVersionByNumber(secretID string, versionNumber int) (*model.SecretVersion, error)
}
