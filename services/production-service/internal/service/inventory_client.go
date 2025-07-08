package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

// ReserveItems резервирует предметы в инвентаре с retry логикой для конкурентных запросов
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

	// Retry логика для обработки конкурентных запросов
	maxRetries := 3
	baseDelay := 100 * time.Millisecond

	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			// Если это последняя попытка или ошибка не связана с сетью
			if attempt == maxRetries {
				return fmt.Errorf("failed to send request: %w", err)
			}

			// Ждем перед повтором
			delay := baseDelay * time.Duration(1<<attempt) // Экспоненциальный backoff
			c.logger.Warn("Retrying reservation request due to network error",
				zap.Int("attempt", attempt+1),
				zap.Duration("delay", delay),
				zap.Error(err))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				continue
			}
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		}

		var errorResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResp)

		// Проверяем, стоит ли повторять запрос
		if resp.StatusCode == http.StatusInternalServerError {
			// Проверяем, если это ошибка блокировки ресурса
			if errorMsg, ok := errorResp["message"].(string); ok &&
				(strings.Contains(errorMsg, "Resource is currently locked") ||
					strings.Contains(errorMsg, "context canceled") ||
					strings.Contains(errorMsg, "context deadline exceeded")) {

				if attempt < maxRetries {
					delay := baseDelay * time.Duration(1<<attempt) // Экспоненциальный backoff
					c.logger.Warn("Retrying reservation request due to resource lock",
						zap.Int("attempt", attempt+1),
						zap.Duration("delay", delay),
						zap.String("error", errorMsg))

					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(delay):
						continue
					}
				}
			}
		}

		// Для всех остальных ошибок (включая insufficient_items) не повторяем
		return fmt.Errorf("reservation failed: %v", errorResp)
	}

	return fmt.Errorf("reservation failed after %d retries", maxRetries)
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
		
		// DEBUG: Логируем что получили от inventory-service
		c.logger.Error("DEBUG: ConsumeReserve response from inventory-service",
			zap.String("operation_id", operationID.String()),
			zap.Int("status_code", resp.StatusCode),
			zap.Any("error_response", errorResp))
		
		// Проверяем специфичный код ошибки "operation_not_found" (новый, корректный способ)
		if code, ok := errorResp["error"].(string); ok && code == "operation_not_found" {
			// резерв отсутствует – это нормальная ситуация, когда резервирование не было создано
			c.logger.Info("No reservation found for operation - this is normal when no reservation was made",
				zap.String("operation_id", operationID.String()),
				zap.String("error_code", code))
			return nil
		}
		
		// Проверяем ошибки со словами "no reservation found" (для обратной совместимости)
		if message, ok := errorResp["message"].(string); ok && strings.Contains(message, "no reservation found") {
			// резерв отсутствует – это нормальная ситуация, когда резервирование не было создано или не найдено
			c.logger.Info("No reservation found for operation - this is normal when no reservation was made",
				zap.String("operation_id", operationID.String()),
				zap.String("message", message))
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
