package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ypxd99/yandex-gophkeeper/internal/model"
)

// FileStorage реализует файловое хранилище для данных
type FileStorage struct {
	filePath string
	mutex    sync.RWMutex
	data     *StorageData
}

// StorageData представляет структуру данных в файле
type StorageData struct {
	Users    map[string]*model.User    `json:"users"`
	Secrets  map[string]*model.Secret  `json:"secrets"`
	Sessions map[string]*model.Session `json:"sessions"`
	Metadata StorageMetadata           `json:"metadata"`
}

// StorageMetadata содержит метаданные хранилища
type StorageMetadata struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   string    `json:"version"`
}

// NewFileStorage создает новое файловое хранилище
func NewFileStorage(filePath string) (*FileStorage, error) {
	storage := &FileStorage{
		filePath: filePath,
		data: &StorageData{
			Users:    make(map[string]*model.User),
			Secrets:  make(map[string]*model.Secret),
			Sessions: make(map[string]*model.Session),
			Metadata: StorageMetadata{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   "1.0.0",
			},
		},
	}

	// Создаем директорию если не существует
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Загружаем существующие данные
	if err := storage.load(); err != nil {
		return nil, fmt.Errorf("failed to load storage: %w", err)
	}

	return storage, nil
}

// load загружает данные из файла
func (f *FileStorage) load() error {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	// Проверяем существование файла
	if _, err := os.Stat(f.filePath); os.IsNotExist(err) {
		// Файл не существует, создаем новый
		return f.save()
	}

	// Читаем файл
	data, err := os.ReadFile(f.filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Парсим JSON
	if err := json.Unmarshal(data, &f.data); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// Инициализируем maps если они nil
	if f.data.Users == nil {
		f.data.Users = make(map[string]*model.User)
	}
	if f.data.Secrets == nil {
		f.data.Secrets = make(map[string]*model.Secret)
	}
	if f.data.Sessions == nil {
		f.data.Sessions = make(map[string]*model.Session)
	}

	return nil
}

// save сохраняет данные в файл
func (f *FileStorage) save() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Обновляем метаданные
	f.data.Metadata.UpdatedAt = time.Now()

	// Сериализуем в JSON
	data, err := json.MarshalIndent(f.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Записываем в файл
	if err := os.WriteFile(f.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// generateID генерирует уникальный ID
func (f *FileStorage) generateID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// Close закрывает хранилище
func (f *FileStorage) Close() error {
	return f.save()
}

// GetUsers возвращает всех пользователей
func (f *FileStorage) GetUsers() []*model.User {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	users := make([]*model.User, 0, len(f.data.Users))
	for _, user := range f.data.Users {
		users = append(users, user)
	}
	return users
}

// GetSecrets возвращает все секреты
func (f *FileStorage) GetSecrets() []*model.Secret {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	secrets := make([]*model.Secret, 0, len(f.data.Secrets))
	for _, secret := range f.data.Secrets {
		secrets = append(secrets, secret)
	}
	return secrets
}

// GetSessions возвращает все сессии
func (f *FileStorage) GetSessions() []*model.Session {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	sessions := make([]*model.Session, 0, len(f.data.Sessions))
	for _, session := range f.data.Sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// Save сохраняет данные в файл
func (f *FileStorage) Save() error {
	return f.save()
}

// GenerateID генерирует уникальный ID
func (f *FileStorage) GenerateID() string {
	return f.generateID()
}

// GetUserByID возвращает пользователя по ID
func (f *FileStorage) GetUserByID(id string) (*model.User, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	user, exists := f.data.Users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// GetUserByUsername возвращает пользователя по username
func (f *FileStorage) GetUserByUsername(username string) (*model.User, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	for _, user := range f.data.Users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

// GetUserByEmail возвращает пользователя по email
func (f *FileStorage) GetUserByEmail(email string) (*model.User, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	for _, user := range f.data.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

// CreateUser создает нового пользователя
// ВАЖНО: Все проверки уникальности выполняются внутри одного блокирования мьютекса
// чтобы избежать deadlock при вызове UserExists или GetByEmail
func (f *FileStorage) CreateUser(user *model.User) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Проверяем уникальность username и email прямо здесь
	for _, existingUser := range f.data.Users {
		if existingUser.Username == user.Username {
			return fmt.Errorf("username already exists")
		}
		if existingUser.Email == user.Email {
			return fmt.Errorf("email already exists")
		}
	}

	// Добавляем пользователя в map
	f.data.Users[user.ID] = user

	// Сохраняем изменения
	return f.save()
}

// UpdateUser обновляет пользователя
func (f *FileStorage) UpdateUser(user *model.User) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Проверяем существование пользователя
	if _, exists := f.data.Users[user.ID]; !exists {
		return fmt.Errorf("user not found")
	}

	// Обновляем пользователя
	f.data.Users[user.ID] = user

	// Сохраняем изменения
	return f.save()
}

// DeleteUser удаляет пользователя
func (f *FileStorage) DeleteUser(id string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Проверяем существование пользователя
	if _, exists := f.data.Users[id]; !exists {
		return fmt.Errorf("user not found")
	}

	// Удаляем пользователя
	delete(f.data.Users, id)

	// Сохраняем изменения
	return f.save()
}

// UserExists проверяет существование пользователя с указанным username
func (f *FileStorage) UserExists(username string) bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	for _, user := range f.data.Users {
		if user.Username == username {
			return true
		}
	}
	return false
}
