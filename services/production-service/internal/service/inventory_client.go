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
	url := fmt.Sprintf("%s/api/inventory/reserve", c.baseURL)

	// Преобразуем наши модели в формат Inventory Service
	inventoryItems := make([]map[string]interface{}, len(items))
	for i, item := range items {
		inventoryItem := map[string]interface{}{
			"item_id":  item.ItemID,
			"quantity": item.Quantity,
		}

		// Используем коды из модели, если они есть, иначе "base" по умолчанию
		if item.CollectionCode != nil && *item.CollectionCode != "" {
			inventoryItem["collection"] = *item.CollectionCode
		} else {
			inventoryItem["collection"] = "base"
		}

		if item.QualityLevelCode != nil && *item.QualityLevelCode != "" {
			inventoryItem["quality_level"] = *item.QualityLevelCode
		} else {
			inventoryItem["quality_level"] = "base"
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
	url := fmt.Sprintf("%s/api/inventory/return-reserve", c.baseURL)

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
	url := fmt.Sprintf("%s/api/inventory/consume-reserve", c.baseURL)

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
		if code, ok := errorResp["error"].(string); ok && code == "operation_not_found" {
			// резерв отсутствует – это не критично, считаем, что всё уже потреблено/не требовалось
			return nil
		}
		return fmt.Errorf("consume reserve failed: %v", errorResp)
	}

	return nil
}

// AddItems добавляет предметы в инвентарь
func (c *HTTPInventoryClient) AddItems(ctx context.Context, userID uuid.UUID, section string, operationType string, operationID uuid.UUID, items []models.AddItem) error {
	url := fmt.Sprintf("%s/api/inventory/add-items", c.baseURL)

	// Преобразуем наши модели в формат Inventory Service
	inventoryItems := make([]map[string]interface{}, len(items))
	for i, item := range items {
		inventoryItem := map[string]interface{}{
			"item_id":  item.ItemID,
			"quantity": item.Quantity,
		}

		// Используем коды из модели, если они есть, иначе "base" по умолчанию
		if item.CollectionCode != nil && *item.CollectionCode != "" {
			inventoryItem["collection"] = *item.CollectionCode
		} else {
			inventoryItem["collection"] = "base"
		}

		if item.QualityLevelCode != nil && *item.QualityLevelCode != "" {
			inventoryItem["quality_level"] = *item.QualityLevelCode
		} else {
			inventoryItem["quality_level"] = "base"
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

	// DEBUG: Логируем что отправляем
	c.logger.Error("DEBUG AddItems request", zap.String("payload", string(body)))

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
