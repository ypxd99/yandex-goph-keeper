package model

import "errors"

var (
	// ErrInvalidData возникает при невалидных данных
	ErrInvalidData = errors.New("invalid data")

	// ErrSecretNotFound возникает когда секрет не найден
	ErrSecretNotFound = errors.New("secret not found")

	// ErrUserNotFound возникает когда пользователь не найден
	ErrUserNotFound = errors.New("user not found")

	// ErrUsernameAlreadyExists возникает при попытке создать пользователя с существующим username
	ErrUsernameAlreadyExists = errors.New("username already exists")

	// ErrEmailAlreadyExists возникает при попытке создать пользователя с существующим email
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrUserAlreadyExists возникает при попытке создать пользователя с существующим username/email
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrUserInactive возникает когда пользователь неактивен
	ErrUserInactive = errors.New("user inactive")

	// ErrInvalidCredentials возникает при неверных учетных данных
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInvalidCurrentPassword возникает при неверном текущем пароле
	ErrInvalidCurrentPassword = errors.New("invalid current password")

	// ErrUnauthorized возникает при отсутствии авторизации
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden возникает при отсутствии прав доступа
	ErrForbidden = errors.New("forbidden")

	// ErrInvalidToken возникает при невалидном токене
	ErrInvalidToken = errors.New("invalid token")

	// ErrInvalidRefreshToken возникает при невалидном refresh токене
	ErrInvalidRefreshToken = errors.New("invalid refresh token")

	// ErrTokenExpired возникает при истечении срока действия токена
	ErrTokenExpired = errors.New("token expired")

	// ErrSessionNotFound возникает когда сессия не найдена
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionInactive возникает когда сессия неактивна
	ErrSessionInactive = errors.New("session inactive")

	// ErrSessionExpired возникает когда сессия истекла
	ErrSessionExpired = errors.New("session expired")
)
