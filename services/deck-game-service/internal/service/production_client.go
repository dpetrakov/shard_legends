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

// productionClient implements ProductionClient interface
type productionClient struct {
	httpClient        *http.Client
	productionBaseURL string
	logger            *slog.Logger
}

// NewProductionClient creates a new production service client
func NewProductionClient(productionBaseURL string, logger *slog.Logger) ProductionClient {
	return &productionClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		productionBaseURL: productionBaseURL,
		logger:            logger,
	}
}

// StartProduction starts a production task
func (c *productionClient) StartProduction(ctx context.Context, jwtToken string, userID uuid.UUID, recipeID uuid.UUID, executionCount int) (*ProductionStartResponse, error) {
	requestBody := map[string]interface{}{
		"recipe_id":       recipeID.String(),
		"execution_count": executionCount,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request body")
	}

	url := fmt.Sprintf("%s/production/factory/start", c.productionBaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwtToken) // JWT authentication

	c.logger.Info("Starting production task",
		"user_id", userID,
		"recipe_id", recipeID,
		"execution_count", executionCount,
		"url", url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to start production", "error", err, "url", url)
		return nil, errors.Wrap(err, "failed to make request to production service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		c.logger.Error("Production service returned error",
			"status_code", resp.StatusCode,
			"user_id", userID,
			"recipe_id", recipeID)
		return nil, fmt.Errorf("production service returned status %d", resp.StatusCode)
	}

	var response ProductionStartResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	c.logger.Info("Production task started successfully",
		"user_id", userID,
		"recipe_id", recipeID,
		"task_id", response.TaskID)

	return &response, nil
}

// ClaimProduction claims the results of a completed production task
func (c *productionClient) ClaimProduction(ctx context.Context, jwtToken string, userID uuid.UUID, taskID uuid.UUID) (*ProductionClaimResponse, error) {
	requestBody := map[string]interface{}{
		"task_id": taskID.String(),
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request body")
	}

	url := fmt.Sprintf("%s/production/factory/claim", c.productionBaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwtToken) // JWT authentication

	c.logger.Info("Claiming production task",
		"user_id", userID,
		"task_id", taskID,
		"url", url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to claim production", "error", err, "url", url)
		return nil, errors.Wrap(err, "failed to make request to production service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Production service returned error on claim",
			"status_code", resp.StatusCode,
			"user_id", userID,
			"task_id", taskID)
		return nil, fmt.Errorf("production service returned status %d", resp.StatusCode)
	}

	var response ProductionClaimResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	c.logger.Info("Production task claimed successfully",
		"user_id", userID,
		"task_id", taskID,
		"items_count", len(response.ItemsReceived))

	return &response, nil
}
