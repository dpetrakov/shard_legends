package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/shard-legends/inventory-service/internal/models"
	"github.com/shard-legends/inventory-service/pkg/logger"
)

// SimpleInventoryServiceMock for testing GetReservationStatus only
type SimpleInventoryServiceMock struct {
	mock.Mock
}

func (m *SimpleInventoryServiceMock) GetReservationStatus(ctx context.Context, operationID uuid.UUID) (*models.ReservationStatusResponse, error) {
	args := m.Called(ctx, operationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ReservationStatusResponse), args.Error(1)
}

// SimpleClassifierServiceMock for testing
type SimpleClassifierServiceMock struct {
	mock.Mock
}

func (m *SimpleClassifierServiceMock) GetClassifierMapping(ctx context.Context, classifierCode string) (map[string]uuid.UUID, error) {
	args := m.Called(ctx, classifierCode)
	return args.Get(0).(map[string]uuid.UUID), args.Error(1)
}

func (m *SimpleClassifierServiceMock) GetReverseClassifierMapping(ctx context.Context, classifierCode string) (map[uuid.UUID]string, error) {
	args := m.Called(ctx, classifierCode)
	return args.Get(0).(map[uuid.UUID]string), args.Error(1)
}

func (m *SimpleClassifierServiceMock) RefreshClassifierCache(ctx context.Context, classifierCode string) error {
	args := m.Called(ctx, classifierCode)
	return args.Error(0)
}

// TestInventoryHandler специально для тестов с минимальным интерфейсом
type TestInventoryHandler struct {
	inventoryService  ReservationStatusService
	classifierService *SimpleClassifierServiceMock
	logger            *slog.Logger
}

// ReservationStatusService interface for testing
type ReservationStatusService interface {
	GetReservationStatus(ctx context.Context, operationID uuid.UUID) (*models.ReservationStatusResponse, error)
}

func (h *TestInventoryHandler) GetReservationStatus(c *gin.Context) {
	// 1. Извлечение operationID из пути
	operationIDParam := c.Param("operationID")
	if operationIDParam == "" {
		h.logger.Error("Missing operationID parameter")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "missing_parameter",
			Message: "operationID parameter is required",
		})
		return
	}

	// 2. Парсинг UUID
	operationID, err := uuid.Parse(operationIDParam)
	if err != nil {
		h.logger.Error("Invalid operationID format", "operationID", operationIDParam, "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_operation_id",
			Message: "Invalid operationID format",
			Details: map[string]interface{}{"operationID": operationIDParam},
		})
		return
	}

	// 3. Вызов сервиса для получения статуса резервирования
	response, err := h.inventoryService.GetReservationStatus(c.Request.Context(), operationID)
	if err != nil {
		h.logger.Error("Failed to get reservation status",
			"operationID", operationID,
			"error", err)

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get reservation status",
		})
		return
	}

	// 4. Определение HTTP статуса ответа
	if !response.ReservationExists {
		h.logger.Info("Reservation not found", "operationID", operationID)
		c.JSON(http.StatusNotFound, response)
		return
	}

	// 5. Возврат успешного ответа
	h.logger.Info("Successfully retrieved reservation status",
		"operationID", operationID,
		"userID", response.UserID,
		"status", response.Status)

	c.JSON(http.StatusOK, response)
}

func setupSimpleTestHandler() (*TestInventoryHandler, *SimpleInventoryServiceMock, *SimpleClassifierServiceMock) {
	gin.SetMode(gin.TestMode)
	
	mockInventoryService := &SimpleInventoryServiceMock{}
	mockClassifierService := &SimpleClassifierServiceMock{}
	logger := logger.NewLogger("debug")
	
	handler := &TestInventoryHandler{
		inventoryService:   mockInventoryService,
		classifierService:  mockClassifierService,
		logger:             logger,
	}
	
	return handler, mockInventoryService, mockClassifierService
}

func TestGetReservationStatus_Success_ActiveReservation(t *testing.T) {
	handler, mockService, _ := setupSimpleTestHandler()

	// Test data
	operationID := uuid.New()
	userID := uuid.New()
	reservationDate := "2025-06-28T10:30:00Z"
	status := "active"

	expectedResponse := &models.ReservationStatusResponse{
		ReservationExists: true,
		OperationID:       &operationID,
		UserID:            &userID,
		ReservedItems: []models.ReservationItemResponse{
			{
				ItemCode:         "wood_plank",
				CollectionCode:   "basic",
				QualityLevelCode: "common",
				Quantity:         5,
			},
		},
		ReservationDate: &reservationDate,
		Status:          &status,
	}

	// Mock expectations
	mockService.On("GetReservationStatus", mock.Anything, operationID).Return(expectedResponse, nil)

	// Setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	// Create a fake HTTP request
	req := httptest.NewRequest("GET", "/api/inventory/reservation/"+operationID.String(), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "operationID", Value: operationID.String()}}

	// Execute
	handler.GetReservationStatus(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ReservationStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.ReservationExists)
	assert.Equal(t, operationID, *response.OperationID)
	assert.Equal(t, userID, *response.UserID)
	assert.Equal(t, "active", *response.Status)
	assert.Len(t, response.ReservedItems, 1)
	assert.Equal(t, "wood_plank", response.ReservedItems[0].ItemCode)
	assert.Equal(t, int64(5), response.ReservedItems[0].Quantity)

	mockService.AssertExpectations(t)
}

func TestGetReservationStatus_Success_ReservationNotFound(t *testing.T) {
	handler, mockService, _ := setupSimpleTestHandler()

	// Test data
	operationID := uuid.New()
	errorMsg := "Reservation not found"

	expectedResponse := &models.ReservationStatusResponse{
		ReservationExists: false,
		Error:             &errorMsg,
	}

	// Mock expectations
	mockService.On("GetReservationStatus", mock.Anything, operationID).Return(expectedResponse, nil)

	// Setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	// Create a fake HTTP request
	req := httptest.NewRequest("GET", "/api/inventory/reservation/"+operationID.String(), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "operationID", Value: operationID.String()}}

	// Execute
	handler.GetReservationStatus(c)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.ReservationStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.ReservationExists)
	assert.Equal(t, "Reservation not found", *response.Error)

	mockService.AssertExpectations(t)
}

func TestGetReservationStatus_InvalidOperationID(t *testing.T) {
	handler, _, _ := setupSimpleTestHandler()

	// Setup request with invalid UUID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "operationID", Value: "invalid-uuid"}}

	// Execute
	handler.GetReservationStatus(c)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "invalid_operation_id", response.Error)
	assert.Contains(t, response.Message, "Invalid operationID format")
}

func TestGetReservationStatus_MissingOperationID(t *testing.T) {
	handler, _, _ := setupSimpleTestHandler()

	// Setup request without operationID parameter
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	// Create a fake HTTP request
	req := httptest.NewRequest("GET", "/api/inventory/reservation/", nil)
	c.Request = req
	c.Params = gin.Params{} // Empty params

	// Execute
	handler.GetReservationStatus(c)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "missing_parameter", response.Error)
	assert.Contains(t, response.Message, "operationID parameter is required")
}

func TestGetReservationStatus_ServiceError(t *testing.T) {
	handler, mockService, _ := setupSimpleTestHandler()

	// Test data
	operationID := uuid.New()

	// Mock expectations - service returns error
	mockService.On("GetReservationStatus", mock.Anything, operationID).Return((*models.ReservationStatusResponse)(nil), assert.AnError)

	// Setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	// Create a fake HTTP request
	req := httptest.NewRequest("GET", "/api/inventory/reservation/"+operationID.String(), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "operationID", Value: operationID.String()}}

	// Execute
	handler.GetReservationStatus(c)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "internal_error", response.Error)
	assert.Contains(t, response.Message, "Failed to get reservation status")

	mockService.AssertExpectations(t)
}

func TestGetReservationStatus_ConsumedReservation(t *testing.T) {
	handler, mockService, _ := setupSimpleTestHandler()

	// Test data
	operationID := uuid.New()
	userID := uuid.New()
	reservationDate := "2025-06-28T10:30:00Z"
	status := "consumed"

	expectedResponse := &models.ReservationStatusResponse{
		ReservationExists: true,
		OperationID:       &operationID,
		UserID:            &userID,
		ReservedItems: []models.ReservationItemResponse{
			{
				ItemCode:         "stone",
				CollectionCode:   "winter_2025",
				QualityLevelCode: "stone",
				Quantity:         10,
			},
		},
		ReservationDate: &reservationDate,
		Status:          &status,
	}

	// Mock expectations
	mockService.On("GetReservationStatus", mock.Anything, operationID).Return(expectedResponse, nil)

	// Setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	// Create a fake HTTP request
	req := httptest.NewRequest("GET", "/api/inventory/reservation/"+operationID.String(), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "operationID", Value: operationID.String()}}

	// Execute
	handler.GetReservationStatus(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ReservationStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.ReservationExists)
	assert.Equal(t, "consumed", *response.Status)

	mockService.AssertExpectations(t)
}