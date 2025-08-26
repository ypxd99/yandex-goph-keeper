package repository

import (
	"context"
	"time"

	"github.com/ypxd99/yandex-gophkeeper/internal/model"
)

// UserFileRepository реализует UserRepository с файловым хранилищем
type UserFileRepository struct {
	storage Storage
}

// NewUserFileRepository создает новый репозиторий пользователей
func NewUserFileRepository(storage Storage) *UserFileRepository {
	return &UserFileRepository{storage: storage}
}

// Create создает нового пользователя
func (r *UserFileRepository) Create(ctx context.Context, user *model.User) error {
	return r.storage.CreateUser(user)
}

// GetByID возвращает пользователя по ID
func (r *UserFileRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	return r.storage.GetUserByID(id)
}

// GetByUsername возвращает пользователя по username
func (r *UserFileRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return r.storage.GetUserByUsername(username)
}

// GetByEmail возвращает пользователя по email
func (r *UserFileRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.storage.GetUserByEmail(email)
}

// Update обновляет данные пользователя
func (r *UserFileRepository) Update(ctx context.Context, user *model.User) error {
	return r.storage.UpdateUser(user)
}

// Delete удаляет пользователя
func (r *UserFileRepository) Delete(ctx context.Context, id string) error {
	return r.storage.DeleteUser(id)
}

// UpdateLastLogin обновляет время последнего входа
func (r *UserFileRepository) UpdateLastLogin(ctx context.Context, id string) error {
	user, err := r.storage.GetUserByID(id)
	if err != nil {
		return err
	}

	user.LastLoginAt = time.Now()
	user.UpdatedAt = time.Now()

	return r.storage.UpdateUser(user)
}

// Exists проверяет существование пользователя с указанным username
func (r *UserFileRepository) Exists(ctx context.Context, username string) bool {
	return r.storage.UserExists(username)
}
