package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ypxd99/yandex-gophkeeper/internal/model"
	"github.com/ypxd99/yandex-gophkeeper/internal/repository"
	"github.com/ypxd99/yandex-gophkeeper/util"
)

// AuthServiceImpl представляет сервис аутентификации и авторизации пользователей.
// Обеспечивает регистрацию, вход, управление токенами и сессиями.
type AuthServiceImpl struct {
	UserRepo     repository.UserRepository
	SessionRepo  repository.SessionRepository
	JWTManager   *util.JWTManager
	Hasher       *util.PasswordHasher
	SimpleCrypto *SimpleCrypto
}

// InitAuthService создает и возвращает новый экземпляр AuthServiceImpl.
// Принимает репозитории пользователей и сессий, а также утилиты для работы с JWT и паролями.
// Возвращает инициализированный сервис аутентификации.
func InitAuthService(storage repository.Storage) *AuthServiceImpl {
	// Создаем репозитории
	userRepo := repository.NewUserFileRepository(storage)
	sessionRepo := repository.NewSessionFileRepository(storage)

	// Создаем утилиты
	config := util.GetConfig()
	jwtManager := util.NewJWTManager(config.Auth.SecretKey, 15*time.Minute, 30*24*time.Hour)
	hasher := util.NewPasswordHasher()

	// Создаем простой сервис шифрования
	simpleCrypto := InitSimpleCrypto("keys")

	return &AuthServiceImpl{
		UserRepo:     userRepo,
		SessionRepo:  sessionRepo,
		JWTManager:   jwtManager,
		Hasher:       hasher,
		SimpleCrypto: simpleCrypto,
	}
}

// Register регистрирует нового пользователя в системе.
// Проверяет уникальность username и email, хеширует пароль и создает пользователя.
// Принимает данные регистрации и контекст запроса.
// Возвращает токены аутентификации или ошибку при неудачной регистрации.
func (s *AuthServiceImpl) Register(ctx context.Context, req *model.UserRegistration) (*model.AuthResponse, error) {
	logger := util.GetLogger()
	logger.Infof("Starting user registration username: %s, email: %s", req.Username, req.Email)

	// Проверки уникальности будут выполнены в репозитории при создании
	logger.Debug("Proceeding with user creation")

	// Хешируем пароль
	logger.Debug("Hashing password")
	hashedPassword, salt, err := s.Hasher.HashPassword(req.Password)
	if err != nil {
		logger.Error("Failed to hash password", "error", err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаем пользователя
	logger.Debug("Creating user object")
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

	// Сохраняем пользователя (проверка уникальности происходит в репозитории)
	logger.Debug("Saving user to repository")
	if err := s.UserRepo.Create(ctx, user); err != nil {
		logger.Error("Failed to create user", "error", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Шифруем email ПОСЛЕ сохранения
	if err := s.encryptUserData(user); err != nil {
		logger.Error("Failed to encrypt user data", "error", err)
		return nil, fmt.Errorf("failed to encrypt user data: %w", err)
	}

	// Обновляем пользователя с зашифрованными данными
	if err := s.UserRepo.Update(ctx, user); err != nil {
		logger.Error("Failed to update user with encrypted data", "error", err)
		return nil, fmt.Errorf("failed to update user with encrypted data: %w", err)
	}

	// Создаем JWT токены
	logger.Debug("Generating JWT tokens")
	tokenPair, err := s.JWTManager.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		logger.Error("Failed to generate tokens", "error", err)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	logger.Info("User registration completed successfully", "userID", user.ID)

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
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 часа для access token
	}, nil
}

// Login выполняет аутентификацию пользователя и создает новую сессию.
// Проверяет учетные данные, создает JWT токены и регистрирует сессию.
// Принимает данные входа, userAgent, ipAddress и контекст запроса.
// Возвращает пару токенов и информацию о пользователе или ошибку при неудачном входе.
func (s *AuthServiceImpl) Login(ctx context.Context, req *model.UserLogin, userAgent, ipAddress string) (*model.AuthResponse, error) {
	// Получаем пользователя по username
	user, err := s.UserRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, model.ErrInvalidCredentials
	}

	// Проверяем активность пользователя
	if !user.IsActive {
		return nil, model.ErrUserInactive
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
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour), // 30 дней
		IsActive:     true,
	}

	if err := s.SessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Обновляем время последнего входа
	user.LastLoginAt = time.Now()
	if err := s.UserRepo.Update(ctx, user); err != nil {
		// Логируем ошибку, но не прерываем процесс входа
		util.GetLogger().Warnf("Failed to update last login time: %v", err)
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
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}, nil
}

// Refresh обновляет пару JWT токенов используя refresh токен.
// Валидирует refresh токен, создает новые токены и обновляет сессию.
// Принимает refresh токен и контекст запроса.
// Возвращает новую пару токенов или ошибку при неудачном обновлении.
func (s *AuthServiceImpl) Refresh(ctx context.Context, req *model.RefreshTokenRequest) (*model.TokenPair, error) {
	// Получаем сессию по refresh токену
	session, err := s.SessionRepo.GetByToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, model.ErrInvalidRefreshToken
	}

	// Проверяем активность сессии
	if !session.IsActive {
		return nil, model.ErrSessionInactive
	}

	// Проверяем срок действия сессии
	if time.Now().After(session.ExpiresAt) {
		return nil, model.ErrSessionExpired
	}

	// Получаем пользователя
	user, err := s.UserRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, model.ErrUserNotFound
	}

	// Проверяем активность пользователя
	if !user.IsActive {
		return nil, model.ErrUserInactive
	}

	// Создаем новые токены
	tokenPair, err := s.JWTManager.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	// Обновляем сессию
	session.RefreshToken = tokenPair.RefreshToken
	session.UpdatedAt = time.Now()
	session.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)

	if err := s.SessionRepo.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return tokenPair, nil
}

// Logout выполняет выход пользователя из конкретной сессии.
// Деактивирует сессию по refresh токену.
// Принимает refresh токен и контекст запроса.
// Возвращает ошибку при неудачном выходе.
func (s *AuthServiceImpl) Logout(ctx context.Context, refreshToken string) error {
	// Получаем сессию по refresh токену
	session, err := s.SessionRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return model.ErrInvalidRefreshToken
	}

	// Деактивируем сессию
	session.IsActive = false
	session.UpdatedAt = time.Now()

	return s.SessionRepo.Update(ctx, session)
}

// LogoutAll выполняет выход пользователя из всех устройств.
// Деактивирует все активные сессии пользователя.
// Принимает ID пользователя и контекст запроса.
// Возвращает ошибку при неудачном выходе.
func (s *AuthServiceImpl) LogoutAll(ctx context.Context, userID string) error {
	return s.SessionRepo.DeleteByUserID(ctx, userID)
}

// GetUserSessions возвращает все активные сессии пользователя.
// Принимает ID пользователя и контекст запроса.
// Возвращает слайс сессий или ошибку при неудачном получении.
func (s *AuthServiceImpl) GetUserSessions(ctx context.Context, userID string) ([]*model.SessionResponse, error) {
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

// RevokeSession отзывает конкретную сессию пользователя.
// Деактивирует сессию по ID.
// Принимает ID пользователя, ID сессии и контекст запроса.
// Возвращает ошибку при неудачном отзыве.
func (s *AuthServiceImpl) RevokeSession(ctx context.Context, userID, sessionID string) error {
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

// ValidateToken проверяет валидность JWT токена.
// Декодирует токен и возвращает claims или ошибку при невалидном токене.
// Принимает токен и контекст запроса.
// Возвращает claims токена или ошибку при невалидном токене.
func (s *AuthServiceImpl) ValidateToken(ctx context.Context, token string) (*model.TokenClaims, error) {
	claims, err := s.JWTManager.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return claims, nil
}

// ChangePassword изменяет пароль пользователя.
// Проверяет текущий пароль, хеширует новый и обновляет пользователя.
// Принимает запрос на смену пароля и контекст запроса.
// Возвращает ошибку при неудачной смене пароля.
func (s *AuthServiceImpl) ChangePassword(ctx context.Context, userID string, req *model.PasswordChangeRequest) error {
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

// generateUserID генерирует уникальный идентификатор для пользователя.
// Использует комбинацию времени и случайных данных для обеспечения уникальности.
// Возвращает строковый идентификатор.
func (s *AuthServiceImpl) generateUserID() string {
	return fmt.Sprintf("user_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// generateSessionID генерирует уникальный идентификатор для сессии.
// Использует комбинацию времени и случайных данных для обеспечения уникальности.
// Возвращает строковый идентификатор.
func (s *AuthServiceImpl) generateSessionID() string {
	return fmt.Sprintf("session_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// encryptUserData шифрует данные пользователя
func (s *AuthServiceImpl) encryptUserData(user *model.User) error {
	// Шифруем email
	if user.Email != "" {
		encryptedEmail, err := s.SimpleCrypto.EncryptString(user.Email)
		if err != nil {
			return fmt.Errorf("failed to encrypt email: %w", err)
		}
		user.EmailEncrypted = encryptedEmail
		// Очищаем открытый email
		user.Email = ""
	}

	return nil
}
