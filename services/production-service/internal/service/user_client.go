package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shard-legends/production-service/internal/models"
	"go.uber.org/zap"
)

// UserClient интерфейс для работы с User Service
type UserClient interface {
	GetUserProductionSlots(ctx context.Context, userID uuid.UUID) (*models.UserProductionSlots, error)
	GetUserProductionModifiers(ctx context.Context, userID uuid.UUID) (*models.UserProductionModifiers, error)
}

// HTTPUserClient реализация клиента через HTTP
type HTTPUserClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewHTTPUserClient создает новый HTTP клиент для User Service
func NewHTTPUserClient(baseURL string, logger *zap.Logger) UserClient {
	return NewHTTPUserClientWithTimeout(baseURL, 5*time.Second, logger)
}

// NewHTTPUserClientWithTimeout создает новый HTTP клиент для User Service с настраиваемым таймаутом
func NewHTTPUserClientWithTimeout(baseURL string, timeout time.Duration, logger *zap.Logger) UserClient {
	return &HTTPUserClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

// GetUserProductionSlots получает производственные слоты пользователя
func (c *HTTPUserClient) GetUserProductionSlots(ctx context.Context, userID uuid.UUID) (*models.UserProductionSlots, error) {
	url := fmt.Sprintf("%s/users/%s/production-slots", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user slots: status %d", resp.StatusCode)
	}

	var slots models.UserProductionSlots
	err = json.NewDecoder(resp.Body).Decode(&slots)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &slots, nil
}

// GetUserProductionModifiers получает производственные модификаторы пользователя
func (c *HTTPUserClient) GetUserProductionModifiers(ctx context.Context, userID uuid.UUID) (*models.UserProductionModifiers, error) {
	url := fmt.Sprintf("%s/users/%s/production-modifiers", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user modifiers: status %d", resp.StatusCode)
	}

	var modifiers models.UserProductionModifiers
	err = json.NewDecoder(resp.Body).Decode(&modifiers)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &modifiers, nil
}
