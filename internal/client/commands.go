package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/ypxd99/yandex-gophkeeper/internal/model"
	"golang.org/x/term"
)

// Commands представляет доступные CLI команды
type Commands struct {
	client *Client
}

// NewCommands создает новый экземпляр команд
func NewCommands(client *Client) *Commands {
	return &Commands{client: client}
}

// Register выполняет регистрацию нового пользователя
func (c *Commands) Register(ctx context.Context) error {
	fmt.Println("=== Регистрация нового пользователя ===")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read username: %w", err)
	}
	username = strings.TrimSpace(username)

	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read email: %w", err)
	}
	email = strings.TrimSpace(email)

	fmt.Print("Password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // Новая строка после ввода пароля

	req := &model.UserRegistration{
		Username: username,
		Email:    email,
		Password: string(password),
	}

	// Отладочная информация
	fmt.Printf("DEBUG: Отправляем запрос: username='%s', email='%s', password='%s' (длина: %d)\n",
		username, email, strings.Repeat("*", len(password)), len(password))

	resp, err := c.client.Register(ctx, req)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	fmt.Println("Регистрация успешна!")
	fmt.Printf("Пользователь: %s\n", resp.User.Username)
	fmt.Printf("Email: %s\n", resp.User.Email)
	fmt.Printf("ID: %s\n", resp.User.ID)

	// Сохраняем токен для дальнейшего использования
	c.client.SetToken(resp.TokenPair.AccessToken)

	return nil
}

// Login выполняет вход пользователя
func (c *Commands) Login(ctx context.Context) error {
	fmt.Println("=== Вход в систему ===")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read username: %w", err)
	}
	username = strings.TrimSpace(username)

	fmt.Print("Password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // Новая строка после ввода пароля

	req := &model.UserLogin{
		Username: username,
		Password: string(password),
	}

	resp, err := c.client.Login(ctx, req)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	fmt.Println("Вход выполнен успешно!")
	fmt.Printf("Пользователь: %s\n", resp.User.Username)

	// Сохраняем токен для дальнейшего использования
	c.client.SetToken(resp.TokenPair.AccessToken)

	return nil
}

// CreateSecret создает новый секрет
func (c *Commands) CreateSecret(ctx context.Context) error {
	fmt.Println("=== Создание нового секрета ===")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Название: ")
	title, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read title: %w", err)
	}
	title = strings.TrimSpace(title)

	fmt.Print("Описание (опционально): ")
	description, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read description: %w", err)
	}
	description = strings.TrimSpace(description)

	fmt.Println("Выберите тип секрета:")
	fmt.Println("1. Логин/Пароль")
	fmt.Println("2. Текстовый секрет")
	fmt.Println("3. Банковская карта")
	fmt.Println("4. OTP")

	fmt.Print("Тип (1-4): ")
	typeChoice, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read type choice: %w", err)
	}
	typeChoice = strings.TrimSpace(typeChoice)

	var secretType model.SecretType
	var secretData model.SecretData

	switch typeChoice {
	case "1":
		secretType = model.SecretTypeLoginPassword
		secretData = c.readLoginPasswordData(reader)
	case "2":
		secretType = model.SecretTypeText
		secretData = c.readTextData(reader)
	case "3":
		secretType = model.SecretTypeCard
		secretData = c.readCardData(reader)
	case "4":
		secretType = model.SecretTypeOTP
		secretData = c.readOTPData(reader)
	default:
		return fmt.Errorf("неверный выбор типа: %s", typeChoice)
	}

	req := &model.CreateSecretRequest{
		Type:        secretType,
		Title:       title,
		Description: description,
	}

	// Сериализуем SecretData в JSON
	dataJSON, err := json.Marshal(secretData)
	if err != nil {
		return fmt.Errorf("failed to marshal secret data: %w", err)
	}

	req.Data = dataJSON

	secret, err := c.client.CreateSecret(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	fmt.Println("Секрет создан успешно!")
	fmt.Printf("ID: %s\n", secret.ID)
	fmt.Printf("Название: %s\n", secret.Title)
	fmt.Printf("Тип: %s\n", secret.Type)

	return nil
}

// ListSecrets выводит список всех секретов
func (c *Commands) ListSecrets(ctx context.Context) error {
	fmt.Println("=== Список секретов ===")

	secrets, err := c.client.GetAllSecrets(ctx)
	if err != nil {
		return fmt.Errorf("failed to get secrets: %w", err)
	}

	if len(secrets) == 0 {
		fmt.Println("Секреты не найдены")
		return nil
	}

	for i, secret := range secrets {
		fmt.Printf("%d. %s (%s)\n", i+1, secret.Title, secret.Type)
		fmt.Printf("   ID: %s\n", secret.ID)
		fmt.Printf("   Описание: %s\n", secret.Description)
		fmt.Printf("   Создан: %s\n", secret.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	return nil
}

// GetSecret выводит детали секрета
func (c *Commands) GetSecret(ctx context.Context, secretID string) error {
	fmt.Printf("=== Детали секрета %s ===\n", secretID)

	secret, err := c.client.GetSecret(ctx, secretID)
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	fmt.Printf("Название: %s\n", secret.Title)
	fmt.Printf("Описание: %s\n", secret.Description)
	fmt.Printf("Тип: %s\n", secret.Type)
	fmt.Printf("Создан: %s\n", secret.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Обновлен: %s\n", secret.UpdatedAt.Format("2006-01-02 15:04:05"))

	// Выводим данные в зависимости от типа
	switch secret.Type {
	case model.SecretTypeLoginPassword:
		if data, ok := secret.Data.(*model.LoginPasswordData); ok {
			fmt.Printf("Логин: %s\n", data.Login)
			fmt.Printf("Пароль: %s\n", strings.Repeat("*", len(data.Password)))
			fmt.Printf("URL: %s\n", data.URL)
			fmt.Printf("Заметки: %s\n", data.Notes)
		}
	case model.SecretTypeText:
		if data, ok := secret.Data.(*model.TextData); ok {
			fmt.Printf("Текст: %s\n", data.Text)
			fmt.Printf("Заметки: %s\n", data.Notes)
		}
	case model.SecretTypeCard:
		if data, ok := secret.Data.(*model.CardData); ok {
			fmt.Printf("Номер: %s\n", data.Number)
			fmt.Printf("Срок действия: %s\n", data.ExpiryDate)
			fmt.Printf("CVV: %s\n", strings.Repeat("*", len(data.CVV)))
			fmt.Printf("Держатель: %s\n", data.Cardholder)
			fmt.Printf("Заметки: %s\n", data.Notes)
		}
	case model.SecretTypeOTP:
		if data, ok := secret.Data.(*model.OTPData); ok {
			fmt.Printf("Секрет: %s\n", strings.Repeat("*", len(data.Secret)))
			fmt.Printf("Тип: %s\n", data.Type)
			fmt.Printf("Цифры: %d\n", data.Digits)
			fmt.Printf("Период: %d\n", data.Period)
			fmt.Printf("Заметки: %s\n", data.Notes)
		}
	}

	return nil
}

// SearchSecrets ищет секреты по названию
func (c *Commands) SearchSecrets(ctx context.Context, query string) error {
	fmt.Printf("=== Поиск секретов: %s ===\n", query)

	secrets, err := c.client.SearchSecrets(ctx, query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(secrets) == 0 {
		fmt.Println("Секреты не найдены")
		return nil
	}

	for i, secret := range secrets {
		fmt.Printf("%d. %s (%s)\n", i+1, secret.Title, secret.Type)
		fmt.Printf("   ID: %s\n", secret.ID)
		fmt.Printf("   Описание: %s\n", secret.Description)
		fmt.Println()
	}

	return nil
}

// DeleteSecret удаляет секрет
func (c *Commands) DeleteSecret(ctx context.Context, secretID string) error {
	fmt.Printf("=== Удаление секрета %s ===\n", secretID)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Вы уверены? (y/N): ")
	confirm, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	confirm = strings.TrimSpace(strings.ToLower(confirm))
	if confirm != "y" && confirm != "yes" {
		fmt.Println("Удаление отменено")
		return nil
	}

	if err := c.client.DeleteSecret(ctx, secretID); err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	fmt.Println("Секрет удален успешно!")
	return nil
}

// Вспомогательные методы для чтения данных секретов

func (c *Commands) readLoginPasswordData(reader *bufio.Reader) *model.LoginPasswordData {
	fmt.Print("Логин: ")
	login, _ := reader.ReadString('\n')
	login = strings.TrimSpace(login)

	fmt.Print("Пароль: ")
	password, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()

	fmt.Print("URL (опционально): ")
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)

	fmt.Print("Заметки (опционально): ")
	notes, _ := reader.ReadString('\n')
	notes = strings.TrimSpace(notes)

	return &model.LoginPasswordData{
		Login:    login,
		Password: string(password),
		URL:      url,
		Notes:    notes,
	}
}

func (c *Commands) readTextData(reader *bufio.Reader) *model.TextData {
	fmt.Print("Текст: ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)

	fmt.Print("Заметки (опционально): ")
	notes, _ := reader.ReadString('\n')
	notes = strings.TrimSpace(notes)

	return &model.TextData{
		Text:  text,
		Notes: notes,
	}
}

func (c *Commands) readCardData(reader *bufio.Reader) *model.CardData {
	fmt.Print("Номер карты: ")
	number, _ := reader.ReadString('\n')
	number = strings.TrimSpace(number)

	fmt.Print("Срок действия (MM/YY): ")
	expiry, _ := reader.ReadString('\n')
	expiry = strings.TrimSpace(expiry)

	fmt.Print("CVV: ")
	cvv, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()

	fmt.Print("Держатель карты: ")
	cardholder, _ := reader.ReadString('\n')
	cardholder = strings.TrimSpace(cardholder)

	fmt.Print("Заметки (опционально): ")
	notes, _ := reader.ReadString('\n')
	notes = strings.TrimSpace(notes)

	return &model.CardData{
		Number:     number,
		ExpiryDate: expiry,
		CVV:        string(cvv),
		Cardholder: cardholder,
		Notes:      notes,
	}
}

func (c *Commands) readOTPData(reader *bufio.Reader) *model.OTPData {
	fmt.Print("Секрет: ")
	secret, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()

	fmt.Print("Тип (TOTP/HOTP): ")
	otpType, _ := reader.ReadString('\n')
	otpType = strings.TrimSpace(otpType)

	fmt.Print("Количество цифр (6): ")
	digitsStr, _ := reader.ReadString('\n')
	digitsStr = strings.TrimSpace(digitsStr)
	digits := 6
	if digitsStr != "" {
		fmt.Sscanf(digitsStr, "%d", &digits)
	}

	fmt.Print("Период для TOTP (30): ")
	periodStr, _ := reader.ReadString('\n')
	periodStr = strings.TrimSpace(periodStr)
	period := 30
	if periodStr != "" {
		fmt.Sscanf(periodStr, "%d", &period)
	}

	fmt.Print("Заметки (опционально): ")
	notes, _ := reader.ReadString('\n')
	notes = strings.TrimSpace(notes)

	return &model.OTPData{
		Secret: string(secret),
		Type:   otpType,
		Digits: digits,
		Period: period,
		Notes:  notes,
	}
}
