package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shard-legends/production-service/internal/models"
	"go.uber.org/zap"
)

// InventoryClient интерфейс для работы с Inventory Service
type InventoryClient interface {
	ReserveItems(ctx context.Context, userID uuid.UUID, operationID uuid.UUID, items []models.ReservationItem) error
	ReturnReserve(ctx context.Context, userID uuid.UUID, operationID uuid.UUID) error
	ConsumeReserve(ctx context.Context, userID uuid.UUID, operationID uuid.UUID) error
	AddItems(ctx context.Context, userID uuid.UUID, section string, operationType string, operationID uuid.UUID, items []models.AddItem) error
}

// HTTPInventoryClient реализация клиента через HTTP
type HTTPInventoryClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewHTTPInventoryClient создает новый HTTP клиент для Inventory Service
func NewHTTPInventoryClient(baseURL string, logger *zap.Logger) InventoryClient {
	return NewHTTPInventoryClientWithTimeout(baseURL, 10*time.Second, logger)
}

// NewHTTPInventoryClientWithTimeout создает новый HTTP клиент для Inventory Service с настраиваемым таймаутом
func NewHTTPInventoryClientWithTimeout(baseURL string, timeout time.Duration, logger *zap.Logger) InventoryClient {
	return &HTTPInventoryClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

// ReserveItems резервирует предметы в инвентаре
func (c *HTTPInventoryClient) ReserveItems(ctx context.Context, userID uuid.UUID, operationID uuid.UUID, items []models.ReservationItem) error {
	url := fmt.Sprintf("%s/inventory/reserve", c.baseURL)

	// Преобразуем наши модели в формат Inventory Service
	inventoryItems := make([]map[string]interface{}, len(items))
	for i, item := range items {
		inventoryItem := map[string]interface{}{
			"item_id":  item.ItemID,
			"quantity": item.Quantity,
		}

		if item.CollectionID != nil {
			inventoryItem["collection_id"] = item.CollectionID
		}
		if item.QualityLevelID != nil {
			inventoryItem["quality_level_id"] = item.QualityLevelID
		}

		inventoryItems[i] = inventoryItem
	}

	payload := map[string]interface{}{
		"user_id":      userID,
		"operation_id": operationID,
		"items":        inventoryItems,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return fmt.Errorf("reservation failed: %v", errorResp)
	}

	return nil
}

// ReturnReserve возвращает зарезервированные предметы
func (c *HTTPInventoryClient) ReturnReserve(ctx context.Context, userID uuid.UUID, operationID uuid.UUID) error {
	url := fmt.Sprintf("%s/inventory/return-reserve", c.baseURL)

	payload := map[string]interface{}{
		"user_id":      userID,
		"operation_id": operationID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return fmt.Errorf("return reserve failed: %v", errorResp)
	}

	return nil
}

// ConsumeReserve потребляет зарезервированные предметы
func (c *HTTPInventoryClient) ConsumeReserve(ctx context.Context, userID uuid.UUID, operationID uuid.UUID) error {
	url := fmt.Sprintf("%s/inventory/consume-reserve", c.baseURL)

	payload := map[string]interface{}{
		"user_id":      userID,
		"operation_id": operationID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return fmt.Errorf("consume reserve failed: %v", errorResp)
	}

	return nil
}

// AddItems добавляет предметы в инвентарь
func (c *HTTPInventoryClient) AddItems(ctx context.Context, userID uuid.UUID, section string, operationType string, operationID uuid.UUID, items []models.AddItem) error {
	url := fmt.Sprintf("%s/inventory/add-items", c.baseURL)

	// Преобразуем наши модели в формат Inventory Service
	inventoryItems := make([]map[string]interface{}, len(items))
	for i, item := range items {
		inventoryItem := map[string]interface{}{
			"item_id":  item.ItemID,
			"quantity": item.Quantity,
		}

		if item.CollectionID != nil {
			inventoryItem["collection_id"] = item.CollectionID
		}
		if item.QualityLevelID != nil {
			inventoryItem["quality_level_id"] = item.QualityLevelID
		}

		inventoryItems[i] = inventoryItem
	}

	payload := map[string]interface{}{
		"user_id":        userID,
		"section":        section,
		"operation_type": operationType,
		"operation_id":   operationID,
		"items":          inventoryItems,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return fmt.Errorf("add items failed: %v", errorResp)
	}

	return nil
}
