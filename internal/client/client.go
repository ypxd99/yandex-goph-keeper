package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ypxd99/yandex-gophkeeper/internal/model"
)

// Client представляет HTTP клиент для взаимодействия с GophKeeper сервером
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewClient создает новый экземпляр клиента
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken устанавливает токен аутентификации для клиента
func (c *Client) SetToken(token string) {
	c.token = token
}

// Register регистрирует нового пользователя
func (c *Client) Register(ctx context.Context, req *model.UserRegistration) (*model.AuthResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.makeRequest(ctx, "POST", "/api/v1/auth/register", data)
	if err != nil {
		return nil, err
	}

	var authResp model.AuthResponse
	if err := json.Unmarshal(resp, &authResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &authResp, nil
}

// Login выполняет вход пользователя
func (c *Client) Login(ctx context.Context, req *model.UserLogin) (*model.AuthResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.makeRequest(ctx, "POST", "/api/v1/auth/login", data)
	if err != nil {
		return nil, err
	}

	var authResp model.AuthResponse
	if err := json.Unmarshal(resp, &authResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &authResp, nil
}

// Refresh обновляет токены доступа
func (c *Client) Refresh(ctx context.Context, req *model.RefreshTokenRequest) (*model.TokenPair, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.makeRequest(ctx, "POST", "/api/v1/auth/refresh", data)
	if err != nil {
		return nil, err
	}

	var tokenPair model.TokenPair
	if err := json.Unmarshal(resp, &tokenPair); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &tokenPair, nil
}

// CreateSecret создает новый секрет
func (c *Client) CreateSecret(ctx context.Context, req *model.CreateSecretRequest) (*model.Secret, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.makeRequest(ctx, "POST", "/api/v1/secrets", data)
	if err != nil {
		return nil, err
	}

	var secret model.Secret
	if err := json.Unmarshal(resp, &secret); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &secret, nil
}

// GetSecret возвращает секрет по ID
func (c *Client) GetSecret(ctx context.Context, secretID string) (*model.Secret, error) {
	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/api/v1/secrets/%s", secretID), nil)
	if err != nil {
		return nil, err
	}

	var secret model.Secret
	if err := json.Unmarshal(resp, &secret); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &secret, nil
}

// GetAllSecrets возвращает все секреты пользователя
func (c *Client) GetAllSecrets(ctx context.Context) ([]*model.Secret, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/v1/secrets", nil)
	if err != nil {
		return nil, err
	}

	var secrets []*model.Secret
	if err := json.Unmarshal(resp, &secrets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return secrets, nil
}

// UpdateSecret обновляет существующий секрет
func (c *Client) UpdateSecret(ctx context.Context, secretID string, req *model.UpdateSecretRequest) (*model.Secret, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.makeRequest(ctx, "PUT", fmt.Sprintf("/api/v1/secrets/%s", secretID), data)
	if err != nil {
		return nil, err
	}

	var secret model.Secret
	if err := json.Unmarshal(resp, &secret); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &secret, nil
}

// DeleteSecret удаляет секрет
func (c *Client) DeleteSecret(ctx context.Context, secretID string) error {
	_, err := c.makeRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/secrets/%s", secretID), nil)
	return err
}

// SearchSecrets ищет секреты по названию
func (c *Client) SearchSecrets(ctx context.Context, query string) ([]*model.Secret, error) {
	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/api/v1/secrets/search/title?q=%s", query), nil)
	if err != nil {
		return nil, err
	}

	var secrets []*model.Secret
	if err := json.Unmarshal(resp, &secrets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return secrets, nil
}

// makeRequest выполняет HTTP запрос к серверу
func (c *Client) makeRequest(ctx context.Context, method, path string, data []byte) ([]byte, error) {
	var req *http.Request
	var err error

	if data != nil {
		req, err = http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewBuffer(data))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
	}

	// Добавляем токен аутентификации если есть
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}
