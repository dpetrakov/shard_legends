package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/shard-legends/deck-game-service/internal/auth"
	"github.com/shard-legends/deck-game-service/internal/models"
)

// MockDeckGameService is a mock implementation of DeckGameService
type MockDeckGameService struct {
	mock.Mock
}

func (m *MockDeckGameService) GetDailyChestStatus(ctx context.Context, userID uuid.UUID) (*models.StatusResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.StatusResponse), args.Error(1)
}

func (m *MockDeckGameService) ClaimDailyChest(ctx context.Context, jwtToken string, userID uuid.UUID, request *models.ClaimRequest) (*models.ClaimResponse, error) {
	args := m.Called(ctx, jwtToken, userID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ClaimResponse), args.Error(1)
}

func (m *MockDeckGameService) OpenChest(ctx context.Context, jwtToken string, userID uuid.UUID, request *models.OpenChestRequest) (*models.OpenChestResponse, error) {
	args := m.Called(ctx, jwtToken, userID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OpenChestResponse), args.Error(1)
}

func (m *MockDeckGameService) BuyItem(ctx context.Context, jwtToken string, userID uuid.UUID, request *models.BuyItemRequest) (*models.BuyItemResponse, error) {
	args := m.Called(ctx, jwtToken, userID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BuyItemResponse), args.Error(1)
}

func setupTestRouter() (*gin.Engine, *MockDeckGameService) {
	gin.SetMode(gin.TestMode)

	mockService := &MockDeckGameService{}
	handler := NewDeckGameHandler(mockService, slog.Default())

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Mock user context
		user := &auth.UserContext{
			UserID: "550e8400-e29b-41d4-a716-446655440000",
		}
		c.Set("user", user)

		// Mock JWT token
		c.Set("jwt_token", "test-jwt-token")

		c.Next()
	})

	deckAPI := router.Group("/deck")
	{
		deckAPI.GET("/daily-chest/status", handler.GetDailyChestStatus)
		deckAPI.POST("/daily-chest/claim", handler.ClaimDailyChest)
		deckAPI.POST("/chest/open", handler.OpenChest)
		deckAPI.POST("/item/buy", handler.BuyItem)
	}

	return router, mockService
}

// stringPtr returns a pointer to the given string value
func stringPtr(s string) *string {
	return &s
}

func TestGetDailyChestStatus_Success(t *testing.T) {
	router, mockService := setupTestRouter()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	expectedResponse := &models.StatusResponse{
		ExpectedCombo: 7,
		Finished:      false,
		CraftsDone:    2,
		LastRewardAt:  nil,
	}

	mockService.On("GetDailyChestStatus", mock.Anything, userID).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/deck/daily-chest/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.StatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.ExpectedCombo, response.ExpectedCombo)
	assert.Equal(t, expectedResponse.Finished, response.Finished)
	assert.Equal(t, expectedResponse.CraftsDone, response.CraftsDone)

	mockService.AssertExpectations(t)
}

func TestGetDailyChestStatus_ServiceError(t *testing.T) {
	router, mockService := setupTestRouter()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	mockService.On("GetDailyChestStatus", mock.Anything, userID).Return(nil, errors.New("database error"))

	req, _ := http.NewRequest("GET", "/deck/daily-chest/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "internal_error", response.Error)

	mockService.AssertExpectations(t)
}

func TestClaimDailyChest_Success(t *testing.T) {
	router, mockService := setupTestRouter()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	requestBody := models.ClaimRequest{
		Combo:        7,
		ChestIndices: []int{1, 3, 5},
	}
	expectedResponse := &models.ClaimResponse{
		Items: []models.ItemInfo{
			{
				ItemID:       "359e86d5-d094-4b2b-b96e-6114e3c66d6b",
				ItemClass:    "chests",
				ItemType:     "reagent_chest",
				Name:         "Большой сундук реагентов",
				Description:  "Содержит большое количество реагентов",
				ImageURL:     "/images/items/default.png",
				Collection:   stringPtr("base"),
				QualityLevel: stringPtr("base"),
				Quantity:     1,
			},
		},
		NextExpectedCombo: 8,
		CraftsDone:        3,
	}

	mockService.On("ClaimDailyChest", mock.Anything, mock.AnythingOfType("string"), userID, &requestBody).Return(expectedResponse, nil)

	reqBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/deck/daily-chest/claim", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ClaimResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.NextExpectedCombo, response.NextExpectedCombo)
	assert.Equal(t, expectedResponse.CraftsDone, response.CraftsDone)
	assert.Len(t, response.Items, 1)
	assert.Equal(t, expectedResponse.Items[0].ItemID, response.Items[0].ItemID)

	mockService.AssertExpectations(t)
}

func TestClaimDailyChest_InvalidCombo(t *testing.T) {
	router, mockService := setupTestRouter()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	requestBody := models.ClaimRequest{
		Combo:        5,
		ChestIndices: []int{1, 3},
	}

	mockService.On("ClaimDailyChest", mock.Anything, mock.AnythingOfType("string"), userID, &requestBody).Return(nil, errors.New("invalid_combo"))

	reqBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/deck/daily-chest/claim", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_combo", response.Error)

	mockService.AssertExpectations(t)
}

func TestClaimDailyChest_DailyFinished(t *testing.T) {
	router, mockService := setupTestRouter()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	requestBody := models.ClaimRequest{
		Combo:        15,
		ChestIndices: []int{1},
	}

	mockService.On("ClaimDailyChest", mock.Anything, mock.AnythingOfType("string"), userID, &requestBody).Return(nil, errors.New("daily_finished"))

	reqBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/deck/daily-chest/claim", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "daily_finished", response.Error)

	mockService.AssertExpectations(t)
}

func TestClaimDailyChest_RecipeNotFound(t *testing.T) {
	router, mockService := setupTestRouter()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	requestBody := models.ClaimRequest{
		Combo:        10,
		ChestIndices: []int{1, 2},
	}

	mockService.On("ClaimDailyChest", mock.Anything, mock.AnythingOfType("string"), userID, &requestBody).Return(nil, errors.New("invalid recipe ID configuration"))

	reqBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/deck/daily-chest/claim", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "recipe_not_found", response.Error)

	mockService.AssertExpectations(t)
}

func TestClaimDailyChest_InvalidRequestBody(t *testing.T) {
	router, _ := setupTestRouter()

	invalidJSON := `{"combo": "invalid", "chest_indices": []}`
	req, _ := http.NewRequest("POST", "/deck/daily-chest/claim", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_request", response.Error)
}

// Test cases for OpenChest handler

func TestOpenChest_Success_WithQuantity(t *testing.T) {
	router, mockService := setupTestRouter()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	requestBody := models.OpenChestRequest{
		ChestType:    "resource_chest",
		QualityLevel: "medium",
		Quantity:     intPtr(3),
	}
	expectedResponse := &models.OpenChestResponse{
		Items: []models.ItemInfo{
			{
				ItemID:       "1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60",
				ItemClass:    "resources",
				ItemType:     "stone",
				Name:         "Камень",
				Description:  "Базовый строительный ресурс",
				ImageURL:     "/images/items/stone.png",
				Collection:   nil,
				QualityLevel: nil,
				Quantity:     4200, // 3 chests * 1400 stone each
			},
		},
		QuantityOpened: 3,
	}

	mockService.On("OpenChest", mock.Anything, mock.AnythingOfType("string"), userID, &requestBody).Return(expectedResponse, nil)

	reqBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/deck/chest/open", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.OpenChestResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.QuantityOpened, response.QuantityOpened)
	assert.Len(t, response.Items, 1)
	assert.Equal(t, expectedResponse.Items[0].Quantity, response.Items[0].Quantity)

	mockService.AssertExpectations(t)
}

func TestOpenChest_Success_WithOpenAll(t *testing.T) {
	router, mockService := setupTestRouter()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	requestBody := models.OpenChestRequest{
		ChestType:    "resource_chest",
		QualityLevel: "small",
		OpenAll:      boolPtr(true),
	}
	expectedResponse := &models.OpenChestResponse{
		Items: []models.ItemInfo{
			{
				ItemID:       "1ac8c2b0-0a7d-4e0e-a6d2-9a90b9094b60",
				ItemClass:    "resources",
				ItemType:     "stone",
				Name:         "Камень",
				Description:  "Базовый строительный ресурс",
				ImageURL:     "/images/items/stone.png",
				Collection:   nil,
				QualityLevel: nil,
				Quantity:     200, // 5 chests * 40 stone each
			},
		},
		QuantityOpened: 5,
	}

	mockService.On("OpenChest", mock.Anything, mock.AnythingOfType("string"), userID, &requestBody).Return(expectedResponse, nil)

	reqBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/deck/chest/open", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.OpenChestResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.QuantityOpened, response.QuantityOpened)
	assert.Len(t, response.Items, 1)

	mockService.AssertExpectations(t)
}

func TestOpenChest_InvalidInput(t *testing.T) {
	router, mockService := setupTestRouter()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	requestBody := models.OpenChestRequest{
		ChestType:    "resource_chest",
		QualityLevel: "medium",
		// Both quantity and open_all missing - should trigger validation error
	}

	mockService.On("OpenChest", mock.Anything, mock.AnythingOfType("string"), userID, &requestBody).Return(nil, errors.New("invalid_input: exactly one of 'quantity' or 'open_all' must be specified"))

	reqBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/deck/chest/open", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_input", response.Error)

	mockService.AssertExpectations(t)
}

func TestOpenChest_RecipeNotFound(t *testing.T) {
	router, mockService := setupTestRouter()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	requestBody := models.OpenChestRequest{
		ChestType:    "reagent_chest", // Recipe not implemented yet
		QualityLevel: "medium",
		Quantity:     intPtr(1),
	}

	mockService.On("OpenChest", mock.Anything, mock.AnythingOfType("string"), userID, &requestBody).Return(nil, errors.New("recipe_not_found: unsupported chest type: reagent_chest"))

	reqBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/deck/chest/open", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "recipe_not_found", response.Error)

	mockService.AssertExpectations(t)
}

func TestOpenChest_InsufficientChests(t *testing.T) {
	router, mockService := setupTestRouter()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	requestBody := models.OpenChestRequest{
		ChestType:    "resource_chest",
		QualityLevel: "large",
		OpenAll:      boolPtr(true),
	}

	mockService.On("OpenChest", mock.Anything, mock.AnythingOfType("string"), userID, &requestBody).Return(nil, errors.New("insufficient_chests"))

	reqBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/deck/chest/open", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "insufficient_chests", response.Error)

	mockService.AssertExpectations(t)
}

func TestOpenChest_InvalidRequestBody(t *testing.T) {
	router, _ := setupTestRouter()

	invalidJSON := `{"chest_type": "invalid_type", "quality_level": "invalid_quality"}`
	req, _ := http.NewRequest("POST", "/deck/chest/open", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_request", response.Error)
}

// Helper functions for test cases
func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
