package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ypxd99/yandex-gophkeeper/internal/model"
	"github.com/ypxd99/yandex-gophkeeper/internal/repository"
)

// JWTManagerInterface определяет интерфейс для JWT менеджера
type JWTManagerInterface interface {
	GenerateTokenPair(userID, username string) (*model.TokenPair, error)
	ValidateToken(token string) (*model.TokenClaims, error)
	RefreshToken(refreshToken string) (*model.TokenPair, error)
}

// PasswordHasherInterface определяет интерфейс для хешера паролей
type PasswordHasherInterface interface {
	HashPassword(password string) (string, string, error)
	CheckPassword(password, hash, salt string) bool
}

// MockUserRepository представляет мок для UserRepository
type MockUserRepository struct {
	users map[string]*model.User
}

// NewMockUserRepository создает новый мок репозитория пользователей
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*model.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, model.ErrUserNotFound
	}
	return user, nil
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, model.ErrUserNotFound
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, model.ErrUserNotFound
}

func (m *MockUserRepository) Exists(ctx context.Context, username string) bool {
	for _, user := range m.users {
		if user.Username == username {
			return true
		}
	}
	return false
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	if _, exists := m.users[user.ID]; !exists {
		return model.ErrUserNotFound
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	if _, exists := m.users[id]; !exists {
		return model.ErrUserNotFound
	}
	delete(m.users, id)
	return nil
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	if _, exists := m.users[id]; !exists {
		return model.ErrUserNotFound
	}
	return nil
}

// MockSessionRepository представляет мок для SessionRepository
type MockSessionRepository struct {
	sessions map[string]*model.Session
}

// NewMockSessionRepository создает новый мок репозитория сессий
func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{
		sessions: make(map[string]*model.Session),
	}
}

func (m *MockSessionRepository) Create(ctx context.Context, session *model.Session) error {
	m.sessions[session.ID] = session
	return nil
}

func (m *MockSessionRepository) GetByID(ctx context.Context, id string) (*model.Session, error) {
	session, exists := m.sessions[id]
	if !exists {
		return nil, model.ErrSessionNotFound
	}
	return session, nil
}

func (m *MockSessionRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Session, error) {
	var userSessions []*model.Session
	for _, session := range m.sessions {
		if session.UserID == userID {
			userSessions = append(userSessions, session)
		}
	}
	return userSessions, nil
}

func (m *MockSessionRepository) GetByToken(ctx context.Context, token string) (*model.Session, error) {
	for _, session := range m.sessions {
		if session.RefreshToken == token {
			return session, nil
		}
	}
	return nil, model.ErrSessionNotFound
}

func (m *MockSessionRepository) Update(ctx context.Context, session *model.Session) error {
	if _, exists := m.sessions[session.ID]; !exists {
		return model.ErrSessionNotFound
	}
	m.sessions[session.ID] = session
	return nil
}

func (m *MockSessionRepository) Delete(ctx context.Context, id string) error {
	if _, exists := m.sessions[id]; !exists {
		return model.ErrSessionNotFound
	}
	delete(m.sessions, id)
	return nil
}

func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	for id, session := range m.sessions {
		if session.UserID == userID {
			delete(m.sessions, id)
		}
	}
	return nil
}

func (m *MockSessionRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	for id, session := range m.sessions {
		if session.ExpiresAt.Before(now) {
			delete(m.sessions, id)
		}
	}
	return nil
}

// MockJWTManager представляет мок для JWTManager
type MockJWTManager struct {
	tokens map[string]*model.TokenPair
}

// NewMockJWTManager создает новый мок JWT менеджера
func NewMockJWTManager() *MockJWTManager {
	return &MockJWTManager{
		tokens: make(map[string]*model.TokenPair),
	}
}

func (m *MockJWTManager) GenerateTokenPair(userID, username string) (*model.TokenPair, error) {
	tokenPair := &model.TokenPair{
		AccessToken:  "mock_access_token_" + userID,
		RefreshToken: "mock_refresh_token_" + userID,
		ExpiresIn:    900,
	}
	m.tokens[userID] = tokenPair
	return tokenPair, nil
}

func (m *MockJWTManager) ValidateToken(token string) (*model.TokenClaims, error) {
	// Простая валидация для тестов
	if len(token) > 0 && token != "invalid_token" {
		return &model.TokenClaims{
			UserID:   "user123",
			Username: "testuser",
			Exp:      time.Now().Add(time.Hour).Unix(),
		}, nil
	}
	return nil, fmt.Errorf("invalid token")
}

func (m *MockJWTManager) RefreshToken(refreshToken string) (*model.TokenPair, error) {
	// Простая логика для тестов
	if refreshToken == "mock_refresh_token_user123" {
		return &model.TokenPair{
			AccessToken:  "new_access_token_user123",
			RefreshToken: "new_refresh_token_user123",
			ExpiresIn:    900,
		}, nil
	}
	return nil, fmt.Errorf("invalid token")
}

// MockPasswordHasher представляет мок для PasswordHasher
type MockPasswordHasher struct {
	hashes map[string]string
	salts  map[string]string
}

// NewMockPasswordHasher создает новый мок хешера паролей
func NewMockPasswordHasher() *MockPasswordHasher {
	return &MockPasswordHasher{
		hashes: make(map[string]string),
		salts:  make(map[string]string),
	}
}

func (m *MockPasswordHasher) HashPassword(password string) (string, string, error) {
	hash := "hashed_" + password
	salt := "salt_" + password
	m.hashes[password] = hash
	m.salts[password] = salt
	return hash, salt, nil
}

func (m *MockPasswordHasher) CheckPassword(password, hash, salt string) bool {
	expectedHash, exists := m.hashes[password]
	if !exists {
		return false
	}
	return expectedHash == hash
}

// Тесты для AuthService

func TestAuthServiceRegister(t *testing.T) {
	mockUserRepo := NewMockUserRepository()
	mockSessionRepo := NewMockSessionRepository()
	mockJWTManager := NewMockJWTManager()
	mockHasher := NewMockPasswordHasher()

	authService := &TestAuthServiceImpl{
		UserRepo:    mockUserRepo,
		SessionRepo: mockSessionRepo,
		JWTManager:  mockJWTManager,
		Hasher:      mockHasher,
	}

	// Тестируем успешную регистрацию
	req := &model.UserRegistration{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	response, err := authService.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	if response.TokenPair.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}

	if response.TokenPair.RefreshToken == "" {
		t.Error("Expected non-empty refresh token")
	}

	// Проверяем, что пользователь создан в репозитории
	user, err := mockUserRepo.GetByUsername(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}

	// Тестируем повторную регистрацию с тем же username
	_, err = authService.Register(context.Background(), req)
	if err != model.ErrUsernameAlreadyExists {
		t.Errorf("Expected ErrUsernameAlreadyExists, got %v", err)
	}
}

func TestAuthServiceLogin(t *testing.T) {
	mockUserRepo := NewMockUserRepository()
	mockSessionRepo := NewMockSessionRepository()
	mockJWTManager := NewMockJWTManager()
	mockHasher := NewMockPasswordHasher()

	authService := &TestAuthServiceImpl{
		UserRepo:    mockUserRepo,
		SessionRepo: mockSessionRepo,
		JWTManager:  mockJWTManager,
		Hasher:      mockHasher,
	}

	// Создаем пользователя для тестирования входа
	password := "password123"
	hash, salt, _ := mockHasher.HashPassword(password)
	user := &model.User{
		ID:           "user123",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hash,
		Salt:         salt,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	mockUserRepo.Create(context.Background(), user)

	// Тестируем успешный вход
	req := &model.UserLogin{
		Username:  "testuser",
		Password:  "password123",
		UserAgent: "test-agent",
		IPAddress: "127.0.0.1",
	}

	response, err := authService.Login(context.Background(), req, "test-agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	if response.TokenPair.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}

	// Тестируем вход с неправильным паролем
	req.Password = "wrongpassword"
	_, err = authService.Login(context.Background(), req, "test-agent", "127.0.0.1")
	if err == nil {
		t.Error("Expected error for wrong password")
	}
}

func TestAuthServiceRefresh(t *testing.T) {
	mockUserRepo := NewMockUserRepository()
	mockSessionRepo := NewMockSessionRepository()
	mockJWTManager := NewMockJWTManager()
	mockHasher := NewMockPasswordHasher()

	authService := &TestAuthServiceImpl{
		UserRepo:    mockUserRepo,
		SessionRepo: mockSessionRepo,
		JWTManager:  mockJWTManager,
		Hasher:      mockHasher,
	}

	// Создаем пользователя для тестирования
	user := &model.User{
		ID:           "user123",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Salt:         "salt",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	mockUserRepo.Create(context.Background(), user)

	// Создаем сессию для тестирования
	session := &model.Session{
		ID:           "session123",
		UserID:       "user123",
		RefreshToken: "mock_refresh_token_user123",
		UserAgent:    "test-agent",
		IPAddress:    "127.0.0.1",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastUsedAt:   time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		IsActive:     true,
	}
	mockSessionRepo.Create(context.Background(), session)

	// Тестируем обновление токена
	req := &model.RefreshTokenRequest{
		RefreshToken: "mock_refresh_token_user123",
	}

	tokenPair, err := authService.Refresh(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}

	if tokenPair == nil {
		t.Fatal("Expected non-nil token pair")
	}

	if tokenPair.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}
}

func TestAuthServiceLogout(t *testing.T) {
	mockUserRepo := NewMockUserRepository()
	mockSessionRepo := NewMockSessionRepository()
	mockJWTManager := NewMockJWTManager()
	mockHasher := NewMockPasswordHasher()

	authService := &TestAuthServiceImpl{
		UserRepo:    mockUserRepo,
		SessionRepo: mockSessionRepo,
		JWTManager:  mockJWTManager,
		Hasher:      mockHasher,
	}

	// Создаем сессию для тестирования
	session := &model.Session{
		ID:           "session123",
		UserID:       "user123",
		RefreshToken: "refresh_token_123",
		UserAgent:    "test-agent",
		IPAddress:    "127.0.0.1",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastUsedAt:   time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		IsActive:     true,
	}
	mockSessionRepo.Create(context.Background(), session)

	// Тестируем выход
	err := authService.Logout(context.Background(), "refresh_token_123")
	if err != nil {
		t.Fatalf("Failed to logout: %v", err)
	}

	// Проверяем, что сессия деактивирована
	updatedSession, err := mockSessionRepo.GetByID(context.Background(), "session123")
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if updatedSession.IsActive {
		t.Error("Expected session to be inactive after logout")
	}
}

func TestAuthServiceGetUserSessions(t *testing.T) {
	mockUserRepo := NewMockUserRepository()
	mockSessionRepo := NewMockSessionRepository()
	mockJWTManager := NewMockJWTManager()
	mockHasher := NewMockPasswordHasher()

	authService := &TestAuthServiceImpl{
		UserRepo:    mockUserRepo,
		SessionRepo: mockSessionRepo,
		JWTManager:  mockJWTManager,
		Hasher:      mockHasher,
	}

	// Создаем сессии для тестирования
	sessions := []*model.Session{
		{
			ID:           "session1",
			UserID:       "user123",
			RefreshToken: "token1",
			UserAgent:    "agent1",
			IPAddress:    "127.0.0.1",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			LastUsedAt:   time.Now(),
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			IsActive:     true,
		},
		{
			ID:           "session2",
			UserID:       "user123",
			RefreshToken: "token2",
			UserAgent:    "agent2",
			IPAddress:    "127.0.0.2",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			LastUsedAt:   time.Now(),
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			IsActive:     true,
		},
		{
			ID:           "session3",
			UserID:       "user456",
			RefreshToken: "token3",
			UserAgent:    "agent3",
			IPAddress:    "127.0.0.3",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			LastUsedAt:   time.Now(),
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			IsActive:     true,
		},
	}

	for _, session := range sessions {
		mockSessionRepo.Create(context.Background(), session)
	}

	// Тестируем получение сессий пользователя
	userSessions, err := authService.GetUserSessions(context.Background(), "user123")
	if err != nil {
		t.Fatalf("Failed to get user sessions: %v", err)
	}

	if len(userSessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(userSessions))
	}

	// Проверяем, что все сессии принадлежат правильному пользователю
	for _, session := range userSessions {
		if session.ID != "session1" && session.ID != "session2" {
			t.Errorf("Unexpected session ID: %s", session.ID)
		}
	}
}

func TestAuthServiceRevokeSession(t *testing.T) {
	mockUserRepo := NewMockUserRepository()
	mockSessionRepo := NewMockSessionRepository()
	mockJWTManager := NewMockJWTManager()
	mockHasher := NewMockPasswordHasher()

	authService := &TestAuthServiceImpl{
		UserRepo:    mockUserRepo,
		SessionRepo: mockSessionRepo,
		JWTManager:  mockJWTManager,
		Hasher:      mockHasher,
	}

	// Создаем сессию для тестирования
	session := &model.Session{
		ID:           "session123",
		UserID:       "user123",
		RefreshToken: "refresh_token_123",
		UserAgent:    "test-agent",
		IPAddress:    "127.0.0.1",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastUsedAt:   time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		IsActive:     true,
	}
	mockSessionRepo.Create(context.Background(), session)

	// Тестируем отзыв сессии
	err := authService.RevokeSession(context.Background(), "user123", "session123")
	if err != nil {
		t.Fatalf("Failed to revoke session: %v", err)
	}

	// Проверяем, что сессия деактивирована
	updatedSession, err := mockSessionRepo.GetByID(context.Background(), "session123")
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if updatedSession.IsActive {
		t.Error("Expected session to be inactive after revocation")
	}

	// Тестируем отзыв сессии другого пользователя
	err = authService.RevokeSession(context.Background(), "user456", "session123")
	if err != model.ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}
}

func TestAuthServiceLogoutAll(t *testing.T) {
	mockUserRepo := NewMockUserRepository()
	mockSessionRepo := NewMockSessionRepository()
	mockJWTManager := NewMockJWTManager()
	mockHasher := NewMockPasswordHasher()

	authService := &TestAuthServiceImpl{
		UserRepo:    mockUserRepo,
		SessionRepo: mockSessionRepo,
		JWTManager:  mockJWTManager,
		Hasher:      mockHasher,
	}

	// Создаем сессии для тестирования
	sessions := []*model.Session{
		{
			ID:           "session1",
			UserID:       "user123",
			RefreshToken: "token1",
			UserAgent:    "agent1",
			IPAddress:    "127.0.0.1",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			LastUsedAt:   time.Now(),
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			IsActive:     true,
		},
		{
			ID:           "session2",
			UserID:       "user123",
			RefreshToken: "token2",
			UserAgent:    "agent2",
			IPAddress:    "127.0.0.2",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			LastUsedAt:   time.Now(),
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			IsActive:     true,
		},
		{
			ID:           "session3",
			UserID:       "user456",
			RefreshToken: "token3",
			UserAgent:    "agent3",
			IPAddress:    "127.0.0.3",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			LastUsedAt:   time.Now(),
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			IsActive:     true,
		},
	}

	for _, session := range sessions {
		mockSessionRepo.Create(context.Background(), session)
	}

	// Тестируем выход из всех устройств
	err := authService.LogoutAll(context.Background(), "user123")
	if err != nil {
		t.Fatalf("Failed to logout from all devices: %v", err)
	}

	// Проверяем, что все сессии пользователя удалены
	userSessions, err := mockSessionRepo.GetByUserID(context.Background(), "user123")
	if err != nil {
		t.Fatalf("Failed to get user sessions: %v", err)
	}

	if len(userSessions) != 0 {
		t.Errorf("Expected 0 sessions after logout all, got %d", len(userSessions))
	}

	// Проверяем, что сессии других пользователей остались
	otherUserSessions, err := mockSessionRepo.GetByUserID(context.Background(), "user456")
	if err != nil {
		t.Fatalf("Failed to get other user sessions: %v", err)
	}

	if len(otherUserSessions) != 1 {
		t.Errorf("Expected 1 session for other user, got %d", len(otherUserSessions))
	}
}

func TestAuthServiceChangePassword(t *testing.T) {
	mockUserRepo := NewMockUserRepository()
	mockSessionRepo := NewMockSessionRepository()
	mockJWTManager := NewMockJWTManager()
	mockHasher := NewMockPasswordHasher()

	authService := &TestAuthServiceImpl{
		UserRepo:    mockUserRepo,
		SessionRepo: mockSessionRepo,
		JWTManager:  mockJWTManager,
		Hasher:      mockHasher,
	}

	// Создаем пользователя для тестирования
	password := "oldpassword"
	hash, salt, _ := mockHasher.HashPassword(password)
	user := &model.User{
		ID:           "user123",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hash,
		Salt:         salt,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	mockUserRepo.Create(context.Background(), user)

	// Тестируем успешную смену пароля
	req := &model.PasswordChangeRequest{
		CurrentPassword: "oldpassword",
		NewPassword:     "newpassword123",
	}

	err := authService.ChangePassword(context.Background(), "user123", req)
	if err != nil {
		t.Fatalf("Failed to change password: %v", err)
	}

	// Проверяем, что пароль изменился
	updatedUser, err := mockUserRepo.GetByID(context.Background(), "user123")
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updatedUser.PasswordHash == hash {
		t.Error("Expected password hash to change")
	}

	// Тестируем смену пароля с неправильным текущим паролем
	req.CurrentPassword = "wrongpassword"
	err = authService.ChangePassword(context.Background(), "user123", req)
	if err == nil {
		t.Error("Expected error for wrong current password")
	}
}

func TestAuthServiceValidateToken(t *testing.T) {
	mockUserRepo := NewMockUserRepository()
	mockSessionRepo := NewMockSessionRepository()
	mockJWTManager := NewMockJWTManager()
	mockHasher := NewMockPasswordHasher()

	authService := &TestAuthServiceImpl{
		UserRepo:    mockUserRepo,
		SessionRepo: mockSessionRepo,
		JWTManager:  mockJWTManager,
		Hasher:      mockHasher,
	}

	// Тестируем валидацию правильного токена
	claims, err := authService.ValidateToken(context.Background(), "valid_token")
	if err != nil {
		t.Fatalf("Failed to validate valid token: %v", err)
	}

	if claims == nil {
		t.Fatal("Expected non-nil claims")
	}

	if claims.UserID != "user123" {
		t.Errorf("Expected user ID 'user123', got '%s'", claims.UserID)
	}

	// Тестируем валидацию неправильного токена
	_, err = authService.ValidateToken(context.Background(), "invalid_token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

// TestAuthServiceImpl представляет тестовую версию AuthServiceImpl с интерфейсами
type TestAuthServiceImpl struct {
	UserRepo    repository.UserRepository
	SessionRepo repository.SessionRepository
	JWTManager  JWTManagerInterface
	Hasher      PasswordHasherInterface
}

// Register регистрирует нового пользователя в системе
func (s *TestAuthServiceImpl) Register(ctx context.Context, req *model.UserRegistration) (*model.AuthResponse, error) {
	// Проверяем уникальность username
	if s.UserRepo.Exists(ctx, req.Username) {
		return nil, model.ErrUsernameAlreadyExists
	}

	// Проверяем уникальность email
	if _, err := s.UserRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, model.ErrEmailAlreadyExists
	}

	// Хешируем пароль
	hashedPassword, salt, err := s.Hasher.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаем пользователя
	user := &model.User{
		ID:           s.generateUserID(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Salt:         salt,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	// Сохраняем пользователя
	if err := s.UserRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Создаем JWT токены
	tokenPair, err := s.JWTManager.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Возвращаем ответ с токенами
	return &model.AuthResponse{
		User: model.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			IsActive:  user.IsActive,
		},
		TokenPair: *tokenPair,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

// Login выполняет аутентификацию пользователя
func (s *TestAuthServiceImpl) Login(ctx context.Context, req *model.UserLogin, userAgent, ipAddress string) (*model.AuthResponse, error) {
	// Получаем пользователя по username
	user, err := s.UserRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, model.ErrUserNotFound
	}

	// Проверяем пароль
	if !s.Hasher.CheckPassword(req.Password, user.PasswordHash, user.Salt) {
		return nil, model.ErrInvalidCredentials
	}

	// Создаем JWT токены
	tokenPair, err := s.JWTManager.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Создаем сессию
	session := &model.Session{
		ID:           s.generateSessionID(),
		UserID:       user.ID,
		RefreshToken: tokenPair.RefreshToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastUsedAt:   time.Now(),
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour), // 30 дней
		IsActive:     true,
	}

	if err := s.SessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Обновляем время последнего входа
	if err := s.UserRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}

	return &model.AuthResponse{
		User: model.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			IsActive:  user.IsActive,
		},
		TokenPair: *tokenPair,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

// Refresh обновляет токены доступа
func (s *TestAuthServiceImpl) Refresh(ctx context.Context, req *model.RefreshTokenRequest) (*model.TokenPair, error) {
	// Получаем сессию по refresh токену
	session, err := s.SessionRepo.GetByToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, model.ErrSessionNotFound
	}

	// Проверяем активность сессии
	if !session.IsActive {
		return nil, model.ErrSessionExpired
	}

	// Получаем пользователя
	user, err := s.UserRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, model.ErrUserNotFound
	}

	// Генерируем новую пару токенов
	tokenPair, err := s.JWTManager.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Обновляем сессию
	session.RefreshToken = tokenPair.RefreshToken
	session.UpdatedAt = time.Now()
	session.LastUsedAt = time.Now()

	if err := s.SessionRepo.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return tokenPair, nil
}

// Logout выполняет выход пользователя
func (s *TestAuthServiceImpl) Logout(ctx context.Context, refreshToken string) error {
	// Получаем сессию по refresh токену
	session, err := s.SessionRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return model.ErrSessionNotFound
	}

	// Деактивируем сессию
	session.IsActive = false
	session.UpdatedAt = time.Now()

	return s.SessionRepo.Update(ctx, session)
}

// LogoutAll выполняет выход из всех устройств
func (s *TestAuthServiceImpl) LogoutAll(ctx context.Context, userID string) error {
	return s.SessionRepo.DeleteByUserID(ctx, userID)
}

// GetUserSessions возвращает активные сессии пользователя
func (s *TestAuthServiceImpl) GetUserSessions(ctx context.Context, userID string) ([]*model.SessionResponse, error) {
	sessions, err := s.SessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Преобразуем в ответы
	responses := make([]*model.SessionResponse, 0, len(sessions))
	for _, session := range sessions {
		if session.IsActive {
			responses = append(responses, &model.SessionResponse{
				ID:         session.ID,
				UserAgent:  session.UserAgent,
				IPAddress:  session.IPAddress,
				CreatedAt:  session.CreatedAt,
				LastUsedAt: session.LastUsedAt,
				ExpiresAt:  session.ExpiresAt,
			})
		}
	}

	return responses, nil
}

// RevokeSession отзывает конкретную сессию пользователя
func (s *TestAuthServiceImpl) RevokeSession(ctx context.Context, userID, sessionID string) error {
	session, err := s.SessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return model.ErrSessionNotFound
	}

	// Проверяем, что сессия принадлежит пользователю
	if session.UserID != userID {
		return model.ErrSessionNotFound
	}

	session.IsActive = false
	session.UpdatedAt = time.Now()

	return s.SessionRepo.Update(ctx, session)
}

// ValidateToken проверяет валидность JWT токена
func (s *TestAuthServiceImpl) ValidateToken(ctx context.Context, token string) (*model.TokenClaims, error) {
	claims, err := s.JWTManager.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return claims, nil
}

// ChangePassword изменяет пароль пользователя
func (s *TestAuthServiceImpl) ChangePassword(ctx context.Context, userID string, req *model.PasswordChangeRequest) error {
	// Получаем пользователя
	user, err := s.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return model.ErrUserNotFound
	}

	// Проверяем текущий пароль
	if !s.Hasher.CheckPassword(req.CurrentPassword, user.PasswordHash, user.Salt) {
		return model.ErrInvalidCurrentPassword
	}

	// Хешируем новый пароль
	newPasswordHash, newSalt, err := s.Hasher.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Обновляем пароль
	user.PasswordHash = newPasswordHash
	user.Salt = newSalt
	user.UpdatedAt = time.Now()

	return s.UserRepo.Update(ctx, user)
}

// generateUserID генерирует уникальный идентификатор для пользователя
func (s *TestAuthServiceImpl) generateUserID() string {
	return fmt.Sprintf("user_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// generateSessionID генерирует уникальный идентификатор для сессии
func (s *TestAuthServiceImpl) generateSessionID() string {
	return fmt.Sprintf("session_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}
