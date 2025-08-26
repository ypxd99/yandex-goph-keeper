package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ypxd99/yandex-gophkeeper/internal/client"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
	serverURL = "http://localhost:8080"
)

func main() {
	// Парсим флаги командной строки
	var (
		showVersion = flag.Bool("version", false, "Show version information")
		showHelp    = flag.Bool("help", false, "Show help information")
		url         = flag.String("server", serverURL, "Server URL")
	)
	flag.Parse()

	// Показываем версию если запрошено
	if *showVersion {
		fmt.Printf("GophKeeper Client v%s\n", version)
		fmt.Printf("Build time: %s\n", buildTime)
		os.Exit(0)
	}

	// Показываем помощь если запрошено
	if *showHelp {
		showHelpInfo()
		os.Exit(0)
	}

	// Создаем клиент
	cli := client.NewClient(*url)
	commands := client.NewCommands(cli)

	// Запускаем интерактивный режим
	if err := runInteractiveMode(commands); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runInteractiveMode запускает интерактивный режим клиента
func runInteractiveMode(commands *client.Commands) error {
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== GophKeeper Client - Password Manager ===")
	fmt.Printf("Подключение к серверу: %s\n", serverURL)
	fmt.Println()

	for {
		showMainMenu()
		fmt.Print("Выберите действие: ")
		choice, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read choice: %w", err)
		}

		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			if err := commands.Register(ctx); err != nil {
				fmt.Printf("Ошибка регистрации: %v\n", err)
			}
		case "2":
			if err := commands.Login(ctx); err != nil {
				fmt.Printf("Ошибка входа: %v\n", err)
			}
		case "3":
			if err := commands.CreateSecret(ctx); err != nil {
				fmt.Printf("Ошибка создания секрета: %v\n", err)
			}
		case "4":
			if err := commands.ListSecrets(ctx); err != nil {
				fmt.Printf("Ошибка получения списка секретов: %v\n", err)
			}
		case "5":
			fmt.Print("Введите ID секрета: ")
			secretID, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Ошибка чтения ID: %v\n", err)
				continue
			}
			secretID = strings.TrimSpace(secretID)
			if err := commands.GetSecret(ctx, secretID); err != nil {
				fmt.Printf("Ошибка получения секрета: %v\n", err)
			}
		case "6":
			fmt.Print("Введите поисковый запрос: ")
			query, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Ошибка чтения запроса: %v\n", err)
				continue
			}
			query = strings.TrimSpace(query)
			if err := commands.SearchSecrets(ctx, query); err != nil {
				fmt.Printf("Ошибка поиска: %v\n", err)
			}
		case "7":
			fmt.Print("Введите ID секрета для удаления: ")
			secretID, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Ошибка чтения ID: %v\n", err)
				continue
			}
			secretID = strings.TrimSpace(secretID)
			if err := commands.DeleteSecret(ctx, secretID); err != nil {
				fmt.Printf("Ошибка удаления секрета: %v\n", err)
			}
		case "0":
			fmt.Println("До свидания!")
			return nil
		default:
			fmt.Println("Неверный выбор. Попробуйте снова.")
		}

		fmt.Println()
		fmt.Print("Нажмите Enter для продолжения...")
		reader.ReadString('\n')
		fmt.Println()
	}
}

// showMainMenu показывает главное меню
func showMainMenu() {
	fmt.Println("Главное меню:")
	fmt.Println("1. Регистрация")
	fmt.Println("2. Вход в систему")
	fmt.Println("3. Создать секрет")
	fmt.Println("4. Список секретов")
	fmt.Println("5. Просмотр секрета")
	fmt.Println("6. Поиск секретов")
	fmt.Println("7. Удалить секрет")
	fmt.Println("0. Выход")
}

func showHelpInfo() {
	fmt.Println("GophKeeper Client - Password Manager")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gophkeeper-client [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -version       Show version information")
	fmt.Println("  -help          Show this help message")
	fmt.Println("  -server URL    Server URL (default: http://localhost:8080)")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  The client runs in interactive mode with the following commands:")
	fmt.Println("  1. Register - Create new user account")
	fmt.Println("  2. Login - Authenticate existing user")
	fmt.Println("  3. Create Secret - Add new secret (login/password, text, card, OTP)")
	fmt.Println("  4. List Secrets - Show all user secrets")
	fmt.Println("  5. View Secret - Show secret details by ID")
	fmt.Println("  6. Search Secrets - Find secrets by title")
	fmt.Println("  7. Delete Secret - Remove secret by ID")
	fmt.Println("  0. Exit - Close the client")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gophkeeper-client -version")
	fmt.Println("  gophkeeper-client -help")
	fmt.Println("  gophkeeper-client -server http://localhost:9000")
}
