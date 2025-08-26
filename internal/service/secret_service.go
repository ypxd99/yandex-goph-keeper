package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ypxd99/yandex-gophkeeper/internal/model"
	"github.com/ypxd99/yandex-gophkeeper/internal/repository"
)

// SecretServiceImpl представляет сервис управления секретами пользователей.
// Обеспечивает создание, чтение, обновление и удаление секретов с шифрованием данных.
type SecretServiceImpl struct {
	SecretRepo   repository.SecretRepository
	simpleCrypto *SimpleCrypto
}

// InitSecretService создает и возвращает новый экземпляр SecretServiceImpl.
// Принимает репозиторий секретов и сервис шифрования.
// Возвращает инициализированный сервис управления секретами.
func InitSecretService(storage repository.Storage) *SecretServiceImpl {
	// Создаем репозиторий
	secretRepo := repository.NewSecretFileRepository(storage)

	// Создаем простой сервис шифрования
	simpleCrypto := InitSimpleCrypto("keys")

	return &SecretServiceImpl{
		SecretRepo:   secretRepo,
		simpleCrypto: simpleCrypto,
	}
}

// Create создает новый секрет для пользователя.
// Шифрует данные секрета и сохраняет в хранилище.
// Принимает контекст запроса, ID пользователя и данные секрета.
// Возвращает созданный секрет или ошибку при неудачном создании.
func (s *SecretServiceImpl) Create(ctx context.Context, userID string, req *model.CreateSecretRequest) (*model.Secret, error) {
	// Десериализуем данные секрета на основе типа
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

	// Шифруем данные секрета
	if err := s.encryptSecretData(secretData, userID); err != nil {
		return nil, fmt.Errorf("failed to encrypt secret data: %w", err)
	}

	// Создаем новый секрет
	secret := &model.Secret{
		ID:          fmt.Sprintf("secret_%d_%d", time.Now().UnixNano(), time.Now().Unix()),
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

	// Сохраняем секрет в репозиторий
	if err := s.SecretRepo.Create(ctx, secret); err != nil {
		return nil, fmt.Errorf("failed to create secret: %w", err)
	}

	// Создаем первую версию секрета
	firstVersion := &model.SecretVersion{
		ID:          s.generateVersionID(),
		SecretID:    secret.ID,
		UserID:      userID,
		Type:        secret.Type,
		Title:       secret.Title,
		Description: secret.Description,
		Tags:        secret.Tags,
		Data:        secret.Data,
		Version:     secret.Version,
		CreatedAt:   secret.CreatedAt,
		UpdatedAt:   secret.UpdatedAt,
		IsActive:    true,
	}

	if err := s.SecretRepo.CreateVersion(ctx, firstVersion); err != nil {
		return nil, fmt.Errorf("failed to create first version: %w", err)
	}

	return secret, nil
}

// GetByID возвращает секрет по ID.
// Проверяет права доступа и расшифровывает данные.
// Принимает контекст запроса, ID пользователя и ID секрета.
// Возвращает секрет или ошибку при неудачном получении.
func (s *SecretServiceImpl) GetByID(ctx context.Context, userID string, secretID string) (*model.Secret, error) {
	// Получаем секрет из репозитория
	secret, err := s.SecretRepo.GetByID(ctx, secretID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// Проверяем права доступа
	if secret.UserID != userID {
		return nil, fmt.Errorf("access denied: secret does not belong to user")
	}

	// Проверяем, что секрет не удален
	if secret.DeletedAt != nil {
		return nil, fmt.Errorf("secret not found")
	}

	return secret, nil
}

// GetAll возвращает все секреты пользователя.
// Расшифровывает данные всех секретов.
// Принимает контекст запроса и ID пользователя.
// Возвращает слайс секретов или ошибку при неудачном получении.
func (s *SecretServiceImpl) GetAll(ctx context.Context, userID string) ([]*model.Secret, error) {
	// Получаем все секреты пользователя
	secrets, err := s.SecretRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets: %w", err)
	}

	// Фильтруем удаленные секреты
	var activeSecrets []*model.Secret
	for _, secret := range secrets {
		if secret.DeletedAt == nil {
			activeSecrets = append(activeSecrets, secret)
		}
	}

	return activeSecrets, nil
}

// GetByType возвращает секреты определенного типа.
// Фильтрует по типу и расшифровывает данные.
// Принимает контекст запроса, ID пользователя и тип секрета.
// Возвращает слайс секретов или ошибку при неудачном получении.
func (s *SecretServiceImpl) GetByType(ctx context.Context, userID string, secretType model.SecretType) ([]*model.Secret, error) {
	// Получаем секреты определенного типа из репозитория
	secrets, err := s.SecretRepo.GetByType(ctx, userID, secretType)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets by type: %w", err)
	}

	// Фильтруем удаленные секреты
	var activeSecrets []*model.Secret
	for _, secret := range secrets {
		if secret.DeletedAt == nil {
			activeSecrets = append(activeSecrets, secret)
		}
	}

	return activeSecrets, nil
}

// Update обновляет существующий секрет.
// Проверяет права доступа, шифрует новые данные и обновляет метаданные.
// Принимает контекст запроса, ID пользователя, ID секрета и новые данные.
// Возвращает обновленный секрет или ошибку при неудачном обновлении.
func (s *SecretServiceImpl) Update(ctx context.Context, userID string, secretID string, req *model.UpdateSecretRequest) (*model.Secret, error) {
	// Получаем секрет для проверки прав доступа
	secret, err := s.SecretRepo.GetByID(ctx, secretID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// Проверяем права доступа
	if secret.UserID != userID {
		return nil, fmt.Errorf("access denied: secret does not belong to user")
	}

	// Проверяем, что секрет не удален
	if secret.DeletedAt != nil {
		return nil, fmt.Errorf("secret not found")
	}

	// Обновляем поля секрета
	if req.Title != nil {
		secret.Title = *req.Title
	}
	if req.Description != nil {
		secret.Description = *req.Description
	}
	if req.Tags != nil {
		secret.Tags = req.Tags
	}
	if req.Data != nil {
		secret.Data = req.Data
	}

	// Обновляем метаданные
	secret.UpdatedAt = time.Now()
	secret.Version++

	// Сохраняем обновленный секрет
	if err := s.SecretRepo.Update(ctx, secret); err != nil {
		return nil, fmt.Errorf("failed to update secret: %w", err)
	}

	// Создаем запись о новой версии
	newVersion := &model.SecretVersion{
		ID:          s.generateVersionID(),
		SecretID:    secretID,
		UserID:      userID,
		Type:        secret.Type,
		Title:       secret.Title,
		Description: secret.Description,
		Tags:        secret.Tags,
		Data:        secret.Data,
		Version:     secret.Version,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
	}

	if err := s.SecretRepo.CreateVersion(ctx, newVersion); err != nil {
		return nil, fmt.Errorf("failed to create version record: %w", err)
	}

	return secret, nil
}

// Delete удаляет секрет по ID.
// Проверяет права доступа и выполняет soft delete.
// Принимает контекст запроса, ID пользователя и ID секрета.
// Возвращает ошибку при неудачном удалении.
func (s *SecretServiceImpl) Delete(ctx context.Context, userID string, secretID string) error {
	// Получаем секрет для проверки прав доступа
	secret, err := s.SecretRepo.GetByID(ctx, secretID)
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	// Проверяем права доступа
	if secret.UserID != userID {
		return fmt.Errorf("access denied: secret does not belong to user")
	}

	// Выполняем soft delete
	now := time.Now()
	secret.DeletedAt = &now
	secret.UpdatedAt = now

	// Обновляем секрет в репозитории
	if err := s.SecretRepo.Update(ctx, secret); err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}

// SearchByTitle ищет секреты по названию.
// Выполняет поиск по частичному совпадению названия.
// Принимает контекст запроса, ID пользователя и поисковый запрос.
// Возвращает слайс найденных секретов или ошибку при неудачном поиске.
func (s *SecretServiceImpl) SearchByTitle(ctx context.Context, userID string, query string) ([]*model.Secret, error) {
	// Получаем все секреты пользователя
	secrets, err := s.SecretRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets: %w", err)
	}

	// Фильтруем по названию и удаленным секретам
	var foundSecrets []*model.Secret
	for _, secret := range secrets {
		if secret.DeletedAt == nil && contains(secret.Title, query) {
			foundSecrets = append(foundSecrets, secret)
		}
	}

	return foundSecrets, nil
}

// SearchByTags ищет секреты по тегам.
// Выполняет поиск по точному совпадению тегов.
// Принимает контекст запроса, ID пользователя и слайс тегов.
// Возвращает слайс найденных секретов или ошибку при неудачном поиске.
func (s *SecretServiceImpl) SearchByTags(ctx context.Context, userID string, tags []string) ([]*model.Secret, error) {
	// Получаем все секреты пользователя
	secrets, err := s.SecretRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets: %w", err)
	}

	// Фильтруем по тегам и удаленным секретам
	var foundSecrets []*model.Secret
	for _, secret := range secrets {
		if secret.DeletedAt == nil && hasMatchingTags(secret.Tags, tags) {
			foundSecrets = append(foundSecrets, secret)
		}
	}

	return foundSecrets, nil
}

// GetVersions возвращает все версии секрета.
// Принимает контекст запроса, ID пользователя и ID секрета.
// Возвращает слайс версий секрета или ошибку при неудачном получении.
func (s *SecretServiceImpl) GetVersions(ctx context.Context, userID, secretID string) ([]*model.SecretVersion, error) {
	// Получаем секрет для проверки прав доступа
	secret, err := s.SecretRepo.GetByID(ctx, secretID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// Проверяем права доступа
	if secret.UserID != userID {
		return nil, fmt.Errorf("access denied: secret does not belong to user")
	}

	// Проверяем, что секрет не удален
	if secret.DeletedAt != nil {
		return nil, fmt.Errorf("secret not found")
	}

	// Получаем все версии секрета
	versions, err := s.SecretRepo.GetVersions(ctx, secretID)
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}

	// Если версий нет, создаем текущую версию
	if len(versions) == 0 {
		currentVersion := &model.SecretVersion{
			ID:          s.generateVersionID(),
			SecretID:    secretID,
			UserID:      userID,
			Type:        secret.Type,
			Title:       secret.Title,
			Description: secret.Description,
			Tags:        secret.Tags,
			Data:        secret.Data,
			Version:     secret.Version,
			CreatedAt:   secret.CreatedAt,
			UpdatedAt:   secret.UpdatedAt,
			IsActive:    true,
		}

		if err := s.SecretRepo.CreateVersion(ctx, currentVersion); err != nil {
			return nil, fmt.Errorf("failed to create current version: %w", err)
		}

		return []*model.SecretVersion{currentVersion}, nil
	}

	return versions, nil
}

// RestoreVersion восстанавливает определенную версию секрета.
// Принимает контекст запроса, ID пользователя, ID секрета и номер версии.
// Возвращает восстановленный секрет или ошибку при неудачном восстановлении.
func (s *SecretServiceImpl) RestoreVersion(ctx context.Context, userID, secretID string, version int) (*model.Secret, error) {
	// Получаем секрет для проверки прав доступа
	secret, err := s.SecretRepo.GetByID(ctx, secretID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// Проверяем права доступа
	if secret.UserID != userID {
		return nil, fmt.Errorf("access denied: secret does not belong to user")
	}

	// Проверяем, что секрет не удален
	if secret.DeletedAt != nil {
		return nil, fmt.Errorf("secret not found")
	}

	// Получаем версию для восстановления
	versionData, err := s.SecretRepo.GetVersionByNumber(ctx, secretID, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get version %d: %w", version, err)
	}

	// Создаем новую версию секрета с восстановленными данными
	restoredSecret := &model.Secret{
		ID:          secret.ID,
		UserID:      secret.UserID,
		Type:        versionData.Type,
		Title:       versionData.Title,
		Description: versionData.Description,
		Tags:        versionData.Tags,
		Data:        versionData.Data,
		Version:     secret.Version + 1, // Увеличиваем версию
		CreatedAt:   secret.CreatedAt,   // Сохраняем оригинальное время создания
		UpdatedAt:   time.Now(),
		DeletedAt:   nil,
	}

	// Сохраняем восстановленный секрет
	if err := s.SecretRepo.Update(ctx, restoredSecret); err != nil {
		return nil, fmt.Errorf("failed to update secret: %w", err)
	}

	// Создаем запись о новой версии
	newVersion := &model.SecretVersion{
		ID:          s.generateVersionID(),
		SecretID:    secretID,
		UserID:      userID,
		Type:        restoredSecret.Type,
		Title:       restoredSecret.Title,
		Description: restoredSecret.Description,
		Tags:        restoredSecret.Tags,
		Data:        restoredSecret.Data,
		Version:     restoredSecret.Version,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
	}

	if err := s.SecretRepo.CreateVersion(ctx, newVersion); err != nil {
		return nil, fmt.Errorf("failed to create version record: %w", err)
	}

	return restoredSecret, nil
}

// contains проверяет, содержит ли строка подстроку (регистронезависимо)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// hasMatchingTags проверяет, есть ли совпадающие теги между секретом и поисковым запросом
func hasMatchingTags(secretTags, searchTags []string) bool {
	for _, searchTag := range searchTags {
		for _, secretTag := range secretTags {
			if strings.EqualFold(searchTag, secretTag) {
				return true
			}
		}
	}
	return false
}

// generateVersionID генерирует уникальный идентификатор для версии секрета.
// Использует комбинацию времени и случайных данных для обеспечения уникальности.
// Возвращает строковый идентификатор.
func (s *SecretServiceImpl) generateVersionID() string {
	return fmt.Sprintf("version_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// encryptSecretData шифрует данные секрета
func (s *SecretServiceImpl) encryptSecretData(secretData model.SecretData, userID string) error {
	// Шифруем данные в зависимости от типа секрета
	switch data := secretData.(type) {
	case *model.LoginPasswordData:
		if err := s.encryptLoginPasswordData(data); err != nil {
			return fmt.Errorf("failed to encrypt login password data: %w", err)
		}
	case *model.TextData:
		if err := s.encryptTextData(data); err != nil {
			return fmt.Errorf("failed to encrypt text data: %w", err)
		}
	case *model.BinaryData:
		if err := s.encryptBinaryData(data); err != nil {
			return fmt.Errorf("failed to encrypt binary data: %w", err)
		}
	case *model.CardData:
		if err := s.encryptCardData(data); err != nil {
			return fmt.Errorf("failed to encrypt card data: %w", err)
		}
	case *model.OTPData:
		if err := s.encryptOTPData(data); err != nil {
			return fmt.Errorf("failed to encrypt OTP data: %w", err)
		}
	}

	return nil
}

// encryptLoginPasswordData шифрует данные логина и пароля
func (s *SecretServiceImpl) encryptLoginPasswordData(data *model.LoginPasswordData) error {
	var err error

	if data.Login != "" {
		data.LoginEncrypted, err = s.simpleCrypto.EncryptString(data.Login)
		if err != nil {
			return fmt.Errorf("failed to encrypt login: %w", err)
		}
		// Очищаем открытые данные после шифрования
		data.Login = ""
	}

	if data.Password != "" {
		data.PasswordEncrypted, err = s.simpleCrypto.EncryptString(data.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		// Очищаем открытые данные после шифрования
		data.Password = ""
	}

	if data.URL != "" {
		data.URLEncrypted, err = s.simpleCrypto.EncryptString(data.URL)
		if err != nil {
			return fmt.Errorf("failed to encrypt URL: %w", err)
		}
		// Очищаем открытые данные после шифрования
		data.URL = ""
	}

	if data.Notes != "" {
		data.NotesEncrypted, err = s.simpleCrypto.EncryptString(data.Notes)
		if err != nil {
			return fmt.Errorf("failed to encrypt notes: %w", err)
		}
		// Очищаем открытые данные после шифрования
		data.Notes = ""
	}

	return nil
}

// encryptTextData шифрует текстовые данные
func (s *SecretServiceImpl) encryptTextData(data *model.TextData) error {
	var err error

	if data.Text != "" {
		data.TextEncrypted, err = s.simpleCrypto.EncryptString(data.Text)
		if err != nil {
			return fmt.Errorf("failed to encrypt text: %w", err)
		}
		// Очищаем открытые данные после шифрования
		data.Text = ""
	}

	if data.Notes != "" {
		data.NotesEncrypted, err = s.simpleCrypto.EncryptString(data.Notes)
		if err != nil {
			return fmt.Errorf("failed to encrypt notes: %w", err)
		}
		// Очищаем открытые данные после шифрования
		data.Notes = ""
	}

	return nil
}

// encryptBinaryData шифрует бинарные данные
func (s *SecretServiceImpl) encryptBinaryData(data *model.BinaryData) error {
	var err error

	if len(data.Data) > 0 {
		encrypted, err := s.simpleCrypto.Encrypt(data.Data)
		if err != nil {
			return fmt.Errorf("failed to encrypt binary data: %w", err)
		}
		data.DataEncrypted = base64.StdEncoding.EncodeToString(encrypted)
	}

	if data.Notes != "" {
		data.NotesEncrypted, err = s.simpleCrypto.EncryptString(data.Notes)
		if err != nil {
			return fmt.Errorf("failed to encrypt notes: %w", err)
		}
	}

	return nil
}

// encryptCardData шифрует данные банковской карты
func (s *SecretServiceImpl) encryptCardData(data *model.CardData) error {
	var err error

	if data.Number != "" {
		data.NumberEncrypted, err = s.simpleCrypto.EncryptString(data.Number)
		if err != nil {
			return fmt.Errorf("failed to encrypt card number: %w", err)
		}
	}

	if data.ExpiryDate != "" {
		data.ExpiryDateEncrypted, err = s.simpleCrypto.EncryptString(data.ExpiryDate)
		if err != nil {
			return fmt.Errorf("failed to encrypt expiry date: %w", err)
		}
	}

	if data.CVV != "" {
		data.CVVEncrypted, err = s.simpleCrypto.EncryptString(data.CVV)
		if err != nil {
			return fmt.Errorf("failed to encrypt CVV: %w", err)
		}
	}

	if data.Cardholder != "" {
		data.CardholderEncrypted, err = s.simpleCrypto.EncryptString(data.Cardholder)
		if err != nil {
			return fmt.Errorf("failed to encrypt cardholder: %w", err)
		}
	}

	if data.Notes != "" {
		data.NotesEncrypted, err = s.simpleCrypto.EncryptString(data.Notes)
		if err != nil {
			return fmt.Errorf("failed to encrypt notes: %w", err)
		}
	}

	return nil
}

// encryptOTPData шифрует данные OTP
func (s *SecretServiceImpl) encryptOTPData(data *model.OTPData) error {
	var err error

	if data.Secret != "" {
		data.SecretEncrypted, err = s.simpleCrypto.EncryptString(data.Secret)
		if err != nil {
			return fmt.Errorf("failed to encrypt OTP secret: %w", err)
		}
	}

	if data.Type != "" {
		data.TypeEncrypted, err = s.simpleCrypto.EncryptString(data.Type)
		if err != nil {
			return fmt.Errorf("failed to encrypt OTP type: %w", err)
		}
	}

	if data.Notes != "" {
		data.NotesEncrypted, err = s.simpleCrypto.EncryptString(data.Notes)
		if err != nil {
			return fmt.Errorf("failed to encrypt notes: %w", err)
		}
	}

	return nil
}
