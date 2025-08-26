package handler

import (
	"net/http"
	_ "net/http/pprof"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ypxd99/yandex-gophkeeper/internal/model"
	"github.com/ypxd99/yandex-gophkeeper/internal/repository/middleware"
	"github.com/ypxd99/yandex-gophkeeper/internal/service"
	"github.com/ypxd99/yandex-gophkeeper/util"
)

// Handler представляет HTTP-обработчик сервиса GophKeeper.
// Содержит бизнес-логику сервиса и предоставляет HTTP-эндпоинты
// для операций управления секретами и аутентификации пользователей.
type Handler struct {
	authService   service.AuthService
	secretService service.SecretService
}

// InitHandler создает и возвращает новый экземпляр Handler с предоставленными сервисами.
// Принимает реализации интерфейсов AuthService и SecretService.
// Возвращает инициализированный обработчик.
func InitHandler(authService service.AuthService, secretService service.SecretService) *Handler {
	return &Handler{
		authService:   authService,
		secretService: secretService,
	}
}

// InitRoutes настраивает все HTTP-маршруты для сервиса GophKeeper.
// Настраивает middleware, аутентификацию и все API-эндпоинты.
// Маршруты включают:
// - Эндпоинты отладки для профилирования (/debug/pprof/*)
// - Эндпоинты метрик и проверки работоспособности (/metrics, /health)
// - Эндпоинты аутентификации (/api/v1/auth/*)
// - Эндпоинты управления секретами (/api/v1/secrets/*)
// - Эндпоинты управления профилем (/api/v1/profile/*)
// Принимает экземпляр gin.Engine для настройки маршрутов.
func (h *Handler) InitRoutes(r *gin.Engine) {
	// Настройка эндпоинтов профилирования
	r.GET("/debug/pprof/", gin.WrapF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.DefaultServeMux.ServeHTTP(w, r)
	})))
	r.GET("/debug/pprof/:profile", gin.WrapF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.DefaultServeMux.ServeHTTP(w, r)
	})))

	// Настройка эндпоинтов метрик и проверки работоспособности
	util.GetMetricsRoute(r)
	util.GetHealthcheckRoute(r)
	util.GetRouteList(r)

	// Настройка middleware
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.CORSMiddleware())

	// Настройка основных эндпоинтов
	r.GET("/health", h.healthCheck)
	r.GET("/version", h.getVersion)

	// Настройка API эндпоинтов
	rAPI := r.Group("/api/v1")

	// Эндпоинты аутентификации (публичные)
	authAPI := rAPI.Group("/auth")
	{
		authAPI.POST("/register", h.register)
		authAPI.POST("/login", h.login)
		authAPI.POST("/refresh", h.refresh)
		authAPI.POST("/logout", h.logout)
	}

	// Эндпоинты управления секретами (защищенные)
	secretsAPI := rAPI.Group("/secrets")
	secretsAPI.Use(middleware.AuthMiddleware(h.authService))
	{
		secretsAPI.POST("", h.createSecret)
		secretsAPI.GET("", h.getAllSecrets)
		secretsAPI.GET("/:id", h.getSecretByID)
		secretsAPI.PUT("/:id", h.updateSecret)
		secretsAPI.DELETE("/:id", h.deleteSecret)
		secretsAPI.GET("/type/:type", h.getSecretsByType)
		secretsAPI.GET("/search/tags", h.searchByTags)
		secretsAPI.GET("/search/title", h.searchByTitle)
	}

	// Эндпоинты управления сессиями (защищенные)
	sessionsAPI := rAPI.Group("/sessions")
	sessionsAPI.Use(middleware.AuthMiddleware(h.authService))
	{
		sessionsAPI.GET("", h.getUserSessions)
		sessionsAPI.DELETE("/:id", h.revokeSession)
		sessionsAPI.DELETE("", h.logoutAll)
	}

	// Эндпоинты управления профилем (защищенные)
	profileAPI := rAPI.Group("/profile")
	profileAPI.Use(middleware.AuthMiddleware(h.authService))
	{
		profileAPI.PUT("/password", h.changePassword)
	}
}

// healthCheck проверяет состояние системы
func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": util.GetCurrentTime()})
}

// getVersion возвращает информацию о версии приложения
func (h *Handler) getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":    "1.0.0",
		"build_time": util.GetCurrentTime(),
	})
}

// register регистрирует нового пользователя
func (h *Handler) register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	// Создаем запрос на регистрацию
	registrationReq := &model.UserRegistration{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	// Вызываем сервис регистрации
	authResp, err := h.authService.Register(c.Request.Context(), registrationReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, authResp)
}

// login выполняет вход пользователя
func (h *Handler) login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	// Получаем User-Agent и IP адрес
	userAgent := c.GetHeader("User-Agent")
	if userAgent == "" {
		userAgent = "Unknown"
	}

	ipAddress := c.ClientIP()
	if ipAddress == "" {
		ipAddress = "Unknown"
	}

	// Создаем запрос на вход
	loginReq := &model.UserLogin{
		Username: req.Username,
		Password: req.Password,
	}

	// Вызываем сервис входа
	authResp, err := h.authService.Login(c.Request.Context(), loginReq, userAgent, ipAddress)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, authResp)
}

// refresh обновляет токены доступа
func (h *Handler) refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	// Создаем запрос на обновление токена
	refreshReq := &model.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}

	// Вызываем сервис обновления токена
	tokenPair, err := h.authService.Refresh(c.Request.Context(), refreshReq)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token refresh failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokenPair)
}

// logout выполняет выход пользователя
func (h *Handler) logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	// Вызываем сервис выхода
	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Logout failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

// createSecret создает новый секрет
func (h *Handler) createSecret(c *gin.Context) {
	var req model.CreateSecretRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	// Получаем ID пользователя из контекста (устанавливается middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Вызываем сервис создания секрета
	secret, err := h.secretService.Create(c.Request.Context(), userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create secret: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, secret)
}

// getAllSecrets возвращает все секреты пользователя
func (h *Handler) getAllSecrets(c *gin.Context) {
	// Получаем ID пользователя из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Вызываем сервис получения всех секретов
	secrets, err := h.secretService.GetAll(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get secrets: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, secrets)
}

// getSecretByID возвращает секрет по ID
func (h *Handler) getSecretByID(c *gin.Context) {
	// Получаем ID секрета из URL
	secretID := c.Param("id")
	if secretID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Secret ID is required"})
		return
	}

	// Получаем ID пользователя из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Вызываем сервис получения секрета
	secret, err := h.secretService.GetByID(c.Request.Context(), userID.(string), secretID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, secret)
}

// updateSecret обновляет секрет
func (h *Handler) updateSecret(c *gin.Context) {
	// Получаем ID секрета из URL
	secretID := c.Param("id")
	if secretID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Secret ID is required"})
		return
	}

	// Получаем ID пользователя из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Парсим тело запроса
	var req model.UpdateSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Вызываем сервис обновления секрета
	updatedSecret, err := h.secretService.Update(c.Request.Context(), userID.(string), secretID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update secret: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedSecret)
}

// deleteSecret удаляет секрет
func (h *Handler) deleteSecret(c *gin.Context) {
	// Получаем ID секрета из URL
	secretID := c.Param("id")
	if secretID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Secret ID is required"})
		return
	}

	// Получаем ID пользователя из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Вызываем сервис удаления секрета
	if err := h.secretService.Delete(c.Request.Context(), userID.(string), secretID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete secret: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Secret deleted successfully"})
}

// getSecretsByType возвращает секреты определенного типа
func (h *Handler) getSecretsByType(c *gin.Context) {
	// Получаем тип секрета из path параметра
	secretTypeStr := c.Param("type")
	if secretTypeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Secret type is required"})
		return
	}

	// Парсим тип секрета
	var secretType model.SecretType
	switch secretTypeStr {
	case "text":
		secretType = model.SecretTypeText
	case "login_password":
		secretType = model.SecretTypeLoginPassword
	case "card":
		secretType = model.SecretTypeCard
	case "binary":
		secretType = model.SecretTypeBinary
	case "otp":
		secretType = model.SecretTypeOTP
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid secret type"})
		return
	}

	// Получаем ID пользователя из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Вызываем сервис получения секретов по типу
	secrets, err := h.secretService.GetByType(c.Request.Context(), userID.(string), secretType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get secrets by type: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, secrets)
}

// searchByTags ищет секреты по тегам
func (h *Handler) searchByTags(c *gin.Context) {
	// Получаем теги из query параметра
	tagsStr := c.Query("tags")
	if tagsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tags are required"})
		return
	}

	// Парсим теги (разделенные запятой)
	tags := strings.Split(tagsStr, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}

	// Получаем ID пользователя из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Вызываем сервис поиска секретов по тегам
	secrets, err := h.secretService.SearchByTags(c.Request.Context(), userID.(string), tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, secrets)
}

// searchByTitle ищет секреты по названию
func (h *Handler) searchByTitle(c *gin.Context) {
	// Получаем поисковый запрос из query параметра
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	// Получаем ID пользователя из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Вызываем сервис поиска секретов
	secrets, err := h.secretService.SearchByTitle(c.Request.Context(), userID.(string), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, secrets)
}

// getUserSessions возвращает активные сессии пользователя
func (h *Handler) getUserSessions(c *gin.Context) {
	// Получаем ID пользователя из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Вызываем сервис получения сессий пользователя
	sessions, err := h.authService.GetUserSessions(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user sessions: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// revokeSession отзывает конкретную сессию
func (h *Handler) revokeSession(c *gin.Context) {
	// Получаем ID сессии из URL
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	// Получаем ID пользователя из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Вызываем сервис отзыва сессии
	if err := h.authService.RevokeSession(c.Request.Context(), userID.(string), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke session: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Session revoked successfully"})
}

// logoutAll выполняет выход из всех устройств
func (h *Handler) logoutAll(c *gin.Context) {
	// Получаем ID пользователя из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Вызываем сервис выхода из всех устройств
	if err := h.authService.LogoutAll(c.Request.Context(), userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout from all devices: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out from all devices successfully"})
}

// changePassword меняет пароль пользователя
func (h *Handler) changePassword(c *gin.Context) {
	// Получаем ID пользователя из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Парсим тело запроса
	var req model.PasswordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Вызываем сервис смены пароля
	if err := h.authService.ChangePassword(c.Request.Context(), userID.(string), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
