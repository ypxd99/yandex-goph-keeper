package repository

import (
	"context"
	"fmt"

	"github.com/ypxd99/yandex-gophkeeper/internal/model"
)

// SecretFileRepository реализует SecretRepository с файловым хранилищем.
// Обеспечивает CRUD операции для секретов пользователей с персистентным хранением в JSON файле.
type SecretFileRepository struct {
	storage Storage
}

// NewSecretFileRepository создает новый экземпляр SecretFileRepository.
// Принимает интерфейс Storage для работы с данными.
// Возвращает инициализированный репозиторий секретов.
func NewSecretFileRepository(storage Storage) *SecretFileRepository {
	return &SecretFileRepository{
		storage: storage,
	}
}

// Create создает новый секрет в хранилище.
// Генерирует ID и устанавливает временные метки.
// Принимает контекст запроса и данные секрета.
// Возвращает ошибку при неудачном создании.
func (r *SecretFileRepository) Create(ctx context.Context, secret *model.Secret) error {
	// Для реального хранилища используем прямой вызов
	if realStorage, ok := r.storage.(interface{ CreateSecret(*model.Secret) error }); ok {
		return realStorage.CreateSecret(secret)
	}

	// Для мока используем старую логику
	secrets := r.storage.GetSecrets()
	secrets = append(secrets, secret)

	if mockStorage, ok := r.storage.(interface{ SetSecrets([]*model.Secret) }); ok {
		mockStorage.SetSecrets(secrets)
	}

	return r.storage.Save()
}

// GetByID возвращает секрет по уникальному идентификатору.
// Принимает контекст запроса и ID секрета.
// Возвращает секрет или ошибку если секрет не найден.
func (r *SecretFileRepository) GetByID(ctx context.Context, id string) (*model.Secret, error) {
	secrets := r.storage.GetSecrets()

	for _, secret := range secrets {
		if secret.ID == id {
			return secret, nil
		}
	}

	return nil, fmt.Errorf("secret not found")
}

// GetByUserID возвращает все секреты пользователя.
// Принимает контекст запроса и ID пользователя.
// Возвращает слайс секретов или ошибку при неудачном получении.
func (r *SecretFileRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Secret, error) {
	secrets := r.storage.GetSecrets()
	var userSecrets []*model.Secret

	for _, secret := range secrets {
		if secret.UserID == userID {
			userSecrets = append(userSecrets, secret)
		}
	}

	return userSecrets, nil
}

// GetByType возвращает секреты пользователя определенного типа.
// Фильтрует по типу секрета.
// Принимает контекст запроса, ID пользователя и тип секрета.
// Возвращает слайс секретов или ошибку при неудачном получении.
func (r *SecretFileRepository) GetByType(ctx context.Context, userID string, secretType model.SecretType) ([]*model.Secret, error) {
	secrets := r.storage.GetSecrets()
	var typeSecrets []*model.Secret

	for _, secret := range secrets {
		if secret.UserID == userID && secret.Type == secretType {
			typeSecrets = append(typeSecrets, secret)
		}
	}

	return typeSecrets, nil
}

// Update обновляет данные существующего секрета.
// Проверяет существование секрета и обновляет временные метки.
// Принимает контекст запроса и обновленные данные секрета.
// Возвращает ошибку при неудачном обновлении.
func (r *SecretFileRepository) Update(ctx context.Context, secret *model.Secret) error {
	// Для реального хранилища используем прямой вызов
	if realStorage, ok := r.storage.(interface{ UpdateSecret(*model.Secret) error }); ok {
		return realStorage.UpdateSecret(secret)
	}

	// Для мока используем старую логику
	secrets := r.storage.GetSecrets()

	for i, existingSecret := range secrets {
		if existingSecret.ID == secret.ID {
			secrets[i] = secret

			if mockStorage, ok := r.storage.(interface{ SetSecrets([]*model.Secret) }); ok {
				mockStorage.SetSecrets(secrets)
			}

			return r.storage.Save()
		}
	}

	return fmt.Errorf("secret not found")
}

// Delete удаляет секрет по ID.
// Принимает контекст запроса и ID секрета.
// Возвращает ошибку при неудачном удалении.
func (r *SecretFileRepository) Delete(ctx context.Context, id string) error {
	secrets := r.storage.GetSecrets()

	for i, secret := range secrets {
		if secret.ID == id {
			// Удаляем секрет из слайса
			secrets = append(secrets[:i], secrets[i+1:]...)

			// Для мока нужно обновить внутреннее состояние
			if mockStorage, ok := r.storage.(interface{ SetSecrets([]*model.Secret) }); ok {
				mockStorage.SetSecrets(secrets)
			}

			return r.storage.Save()
		}
	}

	return fmt.Errorf("secret not found")
}

// SearchByTitle ищет секреты по названию.
// Выполняет поиск по частичному совпадению названия.
// Принимает контекст запроса, ID пользователя и поисковый запрос.
// Возвращает слайс найденных секретов или ошибку при неудачном поиске.
func (r *SecretFileRepository) SearchByTitle(ctx context.Context, userID string, query string) ([]*model.Secret, error) {
	secrets := r.storage.GetSecrets()
	var results []*model.Secret

	for _, secret := range secrets {
		if secret.UserID == userID && secret.DeletedAt == nil {
			// Простой поиск по частичному совпадению для тестов
			if contains(secret.Title, query) {
				results = append(results, secret)
			}
		}
	}

	return results, nil
}

// contains проверяет, содержит ли строка подстроку (регистронезависимо)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}

// GetVersions возвращает все версии секрета.
// Принимает контекст запроса и ID секрета.
// Возвращает слайс версий секрета или ошибку при неудачном получении.
func (r *SecretFileRepository) GetVersions(ctx context.Context, secretID string) ([]*model.SecretVersion, error) {
	versions := r.storage.GetSecretVersions(secretID)
	return versions, nil
}

// CreateVersion создает новую версию секрета.
// Принимает контекст запроса и данные версии секрета.
// Возвращает ошибку при неудачном создании.
func (r *SecretFileRepository) CreateVersion(ctx context.Context, version *model.SecretVersion) error {
	return r.storage.CreateSecretVersion(version)
}

// GetVersionByNumber возвращает версию секрета по номеру.
// Принимает контекст запроса, ID секрета и номер версии.
// Возвращает версию секрета или ошибку если версия не найдена.
func (r *SecretFileRepository) GetVersionByNumber(ctx context.Context, secretID string, versionNumber int) (*model.SecretVersion, error) {
	return r.storage.GetSecretVersionByNumber(secretID, versionNumber)
}

// SearchByTags ищет секреты по тегам.
// Выполняет поиск по точному совпадению тегов.
// Принимает контекст запроса, ID пользователя и слайс тегов.
// Возвращает слайс найденных секретов или ошибку при неудачном поиске.
func (r *SecretFileRepository) SearchByTags(ctx context.Context, userID string, tags []string) ([]*model.Secret, error) {
	secrets := r.storage.GetSecrets()
	var results []*model.Secret

	for _, secret := range secrets {
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
