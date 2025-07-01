package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

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

	return &response, nil
}
