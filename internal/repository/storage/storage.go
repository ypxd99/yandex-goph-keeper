package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ypxd99/yandex-gophkeeper/internal/model"
	"github.com/ypxd99/yandex-gophkeeper/internal/service"
)

// Storage представляет файловое хранилище для данных GophKeeper.
// Обеспечивает персистентное хранение пользователей, секретов и сессий
// в JSON файле с поддержкой конкурентного доступа.
type Storage struct {
	filePath     string
	mutex        sync.RWMutex
	data         *StorageData
	simpleCrypto *service.SimpleCrypto
}

// StorageData представляет структуру данных в файле хранилища.
// Содержит все сущности системы: пользователей, секреты, сессии и метаданные.
type StorageData struct {
	Users          map[string]*model.User          `json:"users"`
	Secrets        map[string]*model.Secret        `json:"secrets"`
	Sessions       map[string]*model.Session       `json:"sessions"`
	SecretVersions map[string]*model.SecretVersion `json:"secret_versions"`
	Metadata       StorageMetadata                 `json:"metadata"`
}

// StorageMetadata содержит метаданные хранилища для отслеживания
// версии, времени создания и последнего обновления данных.
type StorageMetadata struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   string    `json:"version"`
}

// InitStorage создает новое файловое хранилище по указанному пути.
// Инициализирует структуру данных, создает директории при необходимости
// и загружает существующие данные из файла.
// Возвращает инициализированное хранилище или ошибку при неудачной инициализации.
func InitStorage(filePath string, simpleCrypto *service.SimpleCrypto) (*Storage, error) {
	storage := &Storage{
		filePath:     filePath,
		simpleCrypto: simpleCrypto,
	}

	// Создаем директорию если не существует
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Загружаем существующие данные или создаем новый файл
	if err := storage.load(); err != nil {
		return nil, fmt.Errorf("failed to load storage: %w", err)
	}

	return storage, nil
}

// load загружает данные из файла хранилища.
// Если файл не существует, создает новый файл с пустой структурой данных.
func (s *Storage) load() error {
	// Проверяем существование файла
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		// Файл не существует, создаем новый с пустой структурой
		s.data = &StorageData{
			Users:          make(map[string]*model.User),
			Secrets:        make(map[string]*model.Secret),
			Sessions:       make(map[string]*model.Session),
			SecretVersions: make(map[string]*model.SecretVersion),
			Metadata: StorageMetadata{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   "1.0.0",
			},
		}

		// Создаем файл
		return s.save()
	}

	// Читаем файл
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Парсим JSON
	if err := json.Unmarshal(data, &s.data); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// Инициализируем maps если они nil
	if s.data.Users == nil {
		s.data.Users = make(map[string]*model.User)
	}
	if s.data.Secrets == nil {
		s.data.Secrets = make(map[string]*model.Secret)
	}
	if s.data.Sessions == nil {
		s.data.Sessions = make(map[string]*model.Session)
	}
	if s.data.SecretVersions == nil {
		s.data.SecretVersions = make(map[string]*model.SecretVersion)
	}

	return nil
}

// save сохраняет данные в файл хранилища.
// Обновляет метаданные и сериализует структуру данных в JSON.
// ВАЖНО: Эта функция НЕ блокирует мьютекс, так как она вызывается
// из функций, которые уже заблокировали мьютекс записи
func (s *Storage) save() error {
	// Обновляем метаданные
	s.data.Metadata.UpdatedAt = time.Now()

	// Сериализуем в JSON
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Записываем в файл
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GenerateID генерирует уникальный идентификатор для сущностей.
// Использует комбинацию текущего времени и наносекунд для обеспечения уникальности.
// Возвращает строковый идентификатор для использования в системе.
func (s *Storage) GenerateID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// Save сохраняет данные в файл хранилища.
// Обновляет метаданные и сериализует структуру данных в JSON.
// Возвращает ошибку при неудачном сохранении данных.
func (s *Storage) Save() error {
	return s.save()
}

// Close закрывает хранилище и сохраняет все несохраненные изменения.
// Должен вызываться при завершении работы приложения для корректного сохранения данных.
// Возвращает ошибку при неудачном сохранении данных.
func (s *Storage) Close() error {
	return s.save()
}

// GetUsers возвращает копию всех пользователей из хранилища.
// Использует блокировку чтения для безопасного доступа к данным.
// Возвращает слайс указателей на пользователей.
func (s *Storage) GetUsers() []*model.User {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	users := make([]*model.User, 0, len(s.data.Users))
	for _, user := range s.data.Users {
		users = append(users, user)
	}
	return users
}

// GetUserByID возвращает пользователя по ID.
// Использует блокировку чтения для безопасного доступа к данным.
// Возвращает пользователя или ошибку если пользователь не найден.
func (s *Storage) GetUserByID(id string) (*model.User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	user, exists := s.data.Users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// GetUserByUsername возвращает пользователя по username.
// Использует блокировку чтения для безопасного доступа к данным.
// Возвращает пользователя или ошибку если пользователь не найден.
func (s *Storage) GetUserByUsername(username string) (*model.User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, user := range s.data.Users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

// GetUserByEmail возвращает пользователя по email.
// Использует блокировку чтения для безопасного доступа к данным.
// Возвращает пользователя или ошибку если пользователь не найден.
func (s *Storage) GetUserByEmail(email string) (*model.User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, user := range s.data.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

// CreateUser создает нового пользователя.
// Использует блокировку записи для безопасного сохранения данных.
// ВАЖНО: Все проверки уникальности выполняются внутри одного блокирования мьютекса
// чтобы избежать deadlock при вызове UserExists или GetUserByEmail
// Возвращает ошибку при неудачном создании.
func (s *Storage) CreateUser(user *model.User) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Проверяем уникальность username и email прямо здесь
	for _, existingUser := range s.data.Users {
		if existingUser.Username == user.Username {
			return fmt.Errorf("username already exists")
		}
		// Расшифровываем email для сравнения
		decryptedEmail := s.simpleCrypto.GetDecryptedEmail(existingUser.EmailEncrypted)
		if decryptedEmail == user.Email {
			return fmt.Errorf("email already exists")
		}
	}

	// Генерируем ID если не задан
	if user.ID == "" {
		user.ID = s.GenerateID()
	}

	// Устанавливаем временные метки
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	// Сохраняем пользователя
	s.data.Users[user.ID] = user

	// Сохраняем в файл
	return s.save()
}

// UpdateUser обновляет существующего пользователя.
// Использует блокировку записи для безопасного сохранения данных.
// Возвращает ошибку при неудачном обновлении.
func (s *Storage) UpdateUser(user *model.User) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Проверяем существование пользователя
	existingUser, exists := s.data.Users[user.ID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Проверяем уникальность username и email
	for id, u := range s.data.Users {
		if id == user.ID {
			continue
		}
		if u.Username == user.Username {
			return fmt.Errorf("username already exists")
		}
		// Расшифровываем email для сравнения
		decryptedEmail := s.simpleCrypto.GetDecryptedEmail(u.EmailEncrypted)
		if decryptedEmail == user.Email {
			return fmt.Errorf("email already exists")
		}
	}

	// Обновляем временные метки
	user.UpdatedAt = time.Now()
	user.CreatedAt = existingUser.CreatedAt // Сохраняем оригинальное время создания

	// Сохраняем пользователя
	s.data.Users[user.ID] = user

	// Сохраняем в файл
	return s.save()
}

// DeleteUser удаляет пользователя по ID.
// Использует блокировку записи для безопасного удаления данных.
// Возвращает ошибку при неудачном удалении.
func (s *Storage) DeleteUser(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.data.Users[id]; !exists {
		return fmt.Errorf("user not found")
	}

	delete(s.data.Users, id)

	// Сохраняем в файл
	return s.save()
}

// UserExists проверяет существование пользователя с указанным username.
// Использует блокировку чтения для безопасного доступа к данным.
// Возвращает true если пользователь существует, false в противном случае.
func (s *Storage) UserExists(username string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, user := range s.data.Users {
		if user.Username == username {
			return true
		}
	}
	return false
}

// GetSecrets возвращает копию всех секретов из хранилища.
// Использует блокировку чтения для безопасного доступа к данным.
// Возвращает слайс указателей на секреты.
func (s *Storage) GetSecrets() []*model.Secret {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	secrets := make([]*model.Secret, 0, len(s.data.Secrets))
	for _, secret := range s.data.Secrets {
		secrets = append(secrets, secret)
	}
	return secrets
}

// GetSessions возвращает копию всех сессий из хранилища.
// Использует блокировку чтения для безопасного доступа к данным.
// Возвращает слайс указателей на сессии.
func (s *Storage) GetSessions() []*model.Session {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	sessions := make([]*model.Session, 0, len(s.data.Sessions))
	for _, session := range s.data.Sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// CreateSecret создает новый секрет в хранилище.
// Использует блокировку записи для безопасного сохранения данных.
// Возвращает ошибку при неудачном создании.
func (s *Storage) CreateSecret(secret *model.Secret) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Генерируем ID если не задан
	if secret.ID == "" {
		secret.ID = s.GenerateID()
	}

	// Устанавливаем временные метки
	now := time.Now()
	if secret.CreatedAt.IsZero() {
		secret.CreatedAt = now
	}
	if secret.UpdatedAt.IsZero() {
		secret.UpdatedAt = now
	}

	// Сохраняем секрет
	s.data.Secrets[secret.ID] = secret

	// Сохраняем в файл
	return s.save()
}

// UpdateSecret обновляет существующий секрет в хранилище.
// Использует блокировку записи для безопасного сохранения данных.
// Возвращает ошибку при неудачном обновлении.
func (s *Storage) UpdateSecret(secret *model.Secret) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Проверяем существование секрета
	if _, exists := s.data.Secrets[secret.ID]; !exists {
		return fmt.Errorf("secret not found")
	}

	// Обновляем временные метки
	secret.UpdatedAt = time.Now()

	// Сохраняем секрет
	s.data.Secrets[secret.ID] = secret

	// Сохраняем в файл
	return s.save()
}

// DeleteSecret удаляет секрет по ID.
// Использует блокировку записи для безопасного удаления данных.
// Возвращает ошибку при неудачном удалении.
func (s *Storage) DeleteSecret(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.data.Secrets[id]; !exists {
		return fmt.Errorf("secret not found")
	}

	delete(s.data.Secrets, id)

	// Сохраняем в файл
	return s.save()
}

// GetSecretVersions возвращает все версии секрета.
// Использует блокировку чтения для безопасного доступа к данным.
// Возвращает слайс указателей на версии секрета.
func (s *Storage) GetSecretVersions(secretID string) []*model.SecretVersion {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	versions := make([]*model.SecretVersion, 0)
	for _, version := range s.data.SecretVersions {
		if version.SecretID == secretID && version.IsActive {
			versions = append(versions, version)
		}
	}
	return versions
}

// CreateSecretVersion создает новую версию секрета.
// Использует блокировку записи для безопасного сохранения данных.
// Возвращает ошибку при неудачном создании.
func (s *Storage) CreateSecretVersion(version *model.SecretVersion) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Генерируем ID если не задан
	if version.ID == "" {
		version.ID = s.GenerateID()
	}

	// Устанавливаем временные метки
	now := time.Now()
	if version.CreatedAt.IsZero() {
		version.CreatedAt = now
	}
	if version.UpdatedAt.IsZero() {
		version.UpdatedAt = now
	}

	// Сохраняем версию
	s.data.SecretVersions[version.ID] = version

	// Сохраняем в файл
	return s.save()
}

// GetSecretVersionByNumber возвращает версию секрета по номеру.
// Использует блокировку чтения для безопасного доступа к данным.
// Возвращает версию секрета или ошибку если версия не найдена.
func (s *Storage) GetSecretVersionByNumber(secretID string, versionNumber int) (*model.SecretVersion, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, version := range s.data.SecretVersions {
		if version.SecretID == secretID && version.Version == versionNumber && version.IsActive {
			return version, nil
		}
	}
	return nil, fmt.Errorf("version not found")
}
