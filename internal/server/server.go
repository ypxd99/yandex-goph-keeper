package server

import (
	"context"
	"net/http"
	"time"

	"github.com/ypxd99/yandex-gophkeeper/util"
)

// Server представляет HTTP-сервер для сервиса GophKeeper.
// Обертывает стандартный http.Server с дополнительной конфигурацией
// и предоставляет методы для запуска и остановки сервера.
type Server struct {
	httpServer *http.Server
}

// Run запускает HTTP-сервер и начинает прослушивание запросов.
// Возвращает ошибку, если сервер не удалось запустить.
func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

// Stop корректно останавливает сервер с учетом заданного контекста.
// Ожидает завершения существующих соединений перед остановкой.
// Возвращает ошибку, если остановка не удалась.
func (s *Server) Stop(ctx context.Context) error {
	logger := util.GetLogger()
	logger.Info("Shutting down HTTP server...")

	// Сначала пытаемся корректно завершить сервер
	if err := s.httpServer.Shutdown(ctx); err != nil {
		logger.Warnf("Graceful shutdown failed: %v, forcing close", err)
		// Если graceful shutdown не удался, принудительно закрываем
		return s.httpServer.Close()
	}

	logger.Info("HTTP server gracefully stopped")
	return nil
}

// NewServer создает и возвращает новый экземпляр Server с предоставленным обработчиком.
// Настраивает сервер с параметрами из конфигурации приложения,
// включая адрес, таймаут чтения и таймаут записи.
func NewServer(handler http.Handler) *Server {
	cfg := util.GetConfig().Server
	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.ServerAddress,
			ReadTimeout:  time.Duration(cfg.RTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.WTimeout) * time.Second,
			Handler:      handler,
		},
	}
}
