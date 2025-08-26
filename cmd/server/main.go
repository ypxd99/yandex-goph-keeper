package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ypxd99/yandex-gophkeeper/internal/repository/storage"
	"github.com/ypxd99/yandex-gophkeeper/internal/server"
	"github.com/ypxd99/yandex-gophkeeper/internal/service"
	"github.com/ypxd99/yandex-gophkeeper/internal/transport/handler"
	"github.com/ypxd99/yandex-gophkeeper/util"
)

// buildVersion содержит версию сборки, может быть переопределена через ldflags.
// buildDate содержит дату сборки, может быть переопределена через ldflags.
// buildCommit содержит хеш коммита, может быть переопределён через ldflags.
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	// Парсим флаги командной строки
	var showVersion = flag.Bool("version", false, "Show version information")
	var showHelp = flag.Bool("help", false, "Show help information")
	flag.Parse()

	// Показываем версию если запрошено
	if *showVersion {
		fmt.Printf("GophKeeper Server v%s\n", ifNA(buildVersion))
		fmt.Printf("Build date: %s\n", ifNA(buildDate))
		fmt.Printf("Build commit: %s\n", ifNA(buildCommit))
		os.Exit(0)
	}

	// Показываем помощь если запрошено
	if *showHelp {
		showHelpInfo()
		os.Exit(0)
	}

	fmt.Printf("Build version: %s\n", ifNA(buildVersion))
	fmt.Printf("Build date: %s\n", ifNA(buildDate))
	fmt.Printf("Build commit: %s\n", ifNA(buildCommit))

	cfg := util.GetConfig()
	util.InitLogger(cfg.Logger)
	logger := util.GetLogger()
	logger.Info("start GophKeeper service")

	// Инициализация простого сервиса шифрования
	simpleCrypto := service.InitSimpleCrypto("keys")

	// Инициализация файлового хранилища
	repo, err := storage.InitStorage(cfg.FileStoragePath, simpleCrypto)
	if err != nil {
		logger.Errorf("Failed to initialize Storage: %v", err)
		return
	}
	defer repo.Close()

	// Инициализация сервисов
	authService := service.InitAuthService(repo)
	secretService := service.InitSecretService(repo)

	// Инициализация обработчика
	h := handler.InitHandler(authService, secretService)

	// Настройка роутера
	router := gin.Default()
	h.InitRoutes(router)

	// Создание и запуск сервера
	srv := server.NewServer(router)

	go func() {
		util.GetLogger().Infof("GOPHKEEPER server listening at: %s", cfg.Server.ServerAddress)
		if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("error occurred while running HTTP server: %s\n", err.Error())
		}
	}()

	<-ctx.Done()
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		logger.Errorf("HTTP server forced to shutdown: %s", err.Error())
	}

	logger.Info("HTTP GOPHKEEPER service stopped")
}

// ifNA возвращает строку "N/A" если значение пустое
// для корректного отображения информации о сборке.
func ifNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

// showHelpInfo показывает информацию о помощи
func showHelpInfo() {
	fmt.Println("GophKeeper Server - Password Manager Backend")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gophkeeper-server [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -version    Show version information")
	fmt.Println("  -help       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gophkeeper-server -version")
	fmt.Println("  gophkeeper-server -help")
	fmt.Println("  gophkeeper-server")
}
