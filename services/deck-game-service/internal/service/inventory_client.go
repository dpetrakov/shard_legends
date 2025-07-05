package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// inventoryClient implements InventoryClient interface
type inventoryClient struct {
	httpClient       *http.Client
	inventoryBaseURL string
	logger           *slog.Logger
}

// NewInventoryClient creates a new inventory service client
func NewInventoryClient(inventoryBaseURL string, logger *slog.Logger) InventoryClient {
	return &inventoryClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		inventoryBaseURL: inventoryBaseURL,
		logger:           logger,
	}
}

// GetItemsDetails returns detailed information about items
func (c *inventoryClient) GetItemsDetails(ctx context.Context, jwtToken string, items []ItemDetailsRequest, lang string) (*ItemDetailsResponse, error) {
	requestBody := map[string]interface{}{
		"items": items,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request body")
	}

	url := fmt.Sprintf("%s/api/inventory/items/details", c.inventoryBaseURL)
	if lang != "" {
		url = fmt.Sprintf("%s?lang=%s", url, lang)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwtToken) // JWT authentication

	c.logger.Info("Getting items details",
		"items_count", len(items),
		"lang", lang,
		"url", url)

	// Log request details for debugging
	for i, item := range items {
		c.logger.Info("Request item details",
			"index", i,
			"item_id", item.ItemID,
			"collection", item.Collection,
			"quality_level", item.QualityLevel)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to get items details", "error", err, "url", url)
		return nil, errors.Wrap(err, "failed to make request to inventory service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Inventory service returned error",
			"status_code", resp.StatusCode,
			"items_count", len(items))
		return nil, fmt.Errorf("inventory service returned status %d", resp.StatusCode)
	}

	var response ItemDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	c.logger.Info("Items details retrieved successfully",
		"requested_items", len(items),
		"returned_items", len(response.Items))

	// Log response details for debugging
	for i, item := range response.Items {
		c.logger.Info("Response item details",
			"index", i,
			"item_id", item.ItemID,
			"item_class", item.ItemClass,
			"item_type", item.ItemType,
			"collection", item.Collection,
			"quality_level", item.QualityLevel)
	}

	return &response, nil
}

// GetItemQuantity returns the quantity of a specific item in the user's inventory
func (c *inventoryClient) GetItemQuantity(ctx context.Context, jwtToken string, itemID uuid.UUID) (int, error) {
	requestBody := map[string]interface{}{
		"item_id": itemID.String(),
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return 0, errors.Wrap(err, "failed to marshal request body")
	}

	url := fmt.Sprintf("%s/api/inventory/items/quantity", c.inventoryBaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	c.logger.Info("Getting item quantity",
		"item_id", itemID,
		"url", url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to get item quantity", "error", err, "url", url)
		return 0, errors.Wrap(err, "failed to make request to inventory service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Inventory service returned error for quantity",
			"status_code", resp.StatusCode,
			"item_id", itemID)
		return 0, fmt.Errorf("inventory service returned status %d", resp.StatusCode)
	}

	var response struct {
		Quantity int `json:"quantity"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, errors.Wrap(err, "failed to decode response")
	}

	c.logger.Info("Item quantity retrieved successfully",
		"item_id", itemID,
		"quantity", response.Quantity)

	return response.Quantity, nil
}

// GetInventory retrieves the user's full inventory via public endpoint
func (c *inventoryClient) GetInventory(ctx context.Context, jwtToken string) ([]InventoryItem, error) {
	url := fmt.Sprintf("%s/api/inventory", c.inventoryBaseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create inventory request")
	}

	req.Header.Set("Authorization", "Bearer "+jwtToken)

	c.logger.Info("Fetching user inventory", "url", url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to fetch inventory", "error", err)
		return nil, errors.Wrap(err, "failed to fetch inventory")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Inventory service returned error for inventory", "status_code", resp.StatusCode)
		return nil, fmt.Errorf("inventory service returned status %d", resp.StatusCode)
	}

	var response struct {
		Items []InventoryItem `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to decode inventory response")
	}

	c.logger.Info("Inventory retrieved successfully", "items_count", len(response.Items))
	return response.Items, nil
}
