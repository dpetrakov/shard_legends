package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shard-legends/inventory-service/internal/models"
)

// ReserveItems handles POST /inventory/reserve
func (h *InventoryHandler) ReserveItems(c *gin.Context) {
	var req models.ReserveItemsRequest

	// 1. Парсинг ReserveItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse reserve request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format",
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	// 2. Валидация входных данных
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Request validation failed", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_failed",
			Message: "Request validation failed",
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	// 3. Вызов сервиса для резервирования предметов
	operationIDs, err := h.inventoryService.ReserveItems(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to reserve items",
			"user_id", req.UserID,
			"operation_id", req.OperationID,
			"error", err)

		// Check if it's insufficient balance error
		if insufficientErr, ok := err.(*models.InsufficientItemsError); ok {
			c.JSON(http.StatusBadRequest, insufficientErr)
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to reserve items",
		})
		return
	}

	// 4. Возврат OperationResponse
	response := models.OperationResponse{
		Success:      true,
		OperationIDs: operationIDs,
	}

	h.logger.Info("Successfully reserved items",
		"user_id", req.UserID,
		"operation_id", req.OperationID,
		"operations_count", len(operationIDs))

	c.JSON(http.StatusOK, response)
}

// ReturnReservedItems handles POST /inventory/return-reserve
func (h *InventoryHandler) ReturnReservedItems(c *gin.Context) {
	var req models.ReturnReserveRequest

	// 1. Парсинг ReturnReserveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse return reserve request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format",
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	// 2. Валидация входных данных
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Request validation failed", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_failed",
			Message: "Request validation failed",
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	// 3. Вызов сервиса для возврата резерва
	err := h.inventoryService.ReturnReservedItems(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to return reserved items",
			"user_id", req.UserID,
			"operation_id", req.OperationID,
			"error", err)

		// Check if operation not found
		if err.Error() == "operation not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "operation_not_found",
				Message: "Reservation operation not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to return reserved items",
		})
		return
	}

	// 4. Возврат OperationResponse
	response := models.OperationResponse{
		Success:      true,
		OperationIDs: []uuid.UUID{}, // Will be filled by service
	}

	h.logger.Info("Successfully returned reserved items",
		"user_id", req.UserID,
		"operation_id", req.OperationID)

	c.JSON(http.StatusOK, response)
}

// ConsumeReservedItems handles POST /inventory/consume-reserve
func (h *InventoryHandler) ConsumeReservedItems(c *gin.Context) {
	var req models.ConsumeReserveRequest

	// 1. Парсинг ConsumeReserveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse consume reserve request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format",
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	// 2. Валидация входных данных
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Request validation failed", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_failed",
			Message: "Request validation failed",
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	// 3. Вызов сервиса для потребления резерва
	err := h.inventoryService.ConsumeReservedItems(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to consume reserved items",
			"user_id", req.UserID,
			"operation_id", req.OperationID,
			"error", err)

		// Check if operation not found
		if err.Error() == "operation not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "operation_not_found",
				Message: "Reservation operation not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to consume reserved items",
		})
		return
	}

	// 4. Возврат OperationResponse
	response := models.OperationResponse{
		Success:      true,
		OperationIDs: []uuid.UUID{}, // Will be filled by service
	}

	h.logger.Info("Successfully consumed reserved items",
		"user_id", req.UserID,
		"operation_id", req.OperationID)

	c.JSON(http.StatusOK, response)
}

// AddItems handles POST /inventory/add-items
func (h *InventoryHandler) AddItems(c *gin.Context) {
	var req models.AddItemsRequest

	// 1. Парсинг AddItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse add items request", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format",
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	// 2. Валидация входных данных
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Request validation failed", "error", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_failed",
			Message: "Request validation failed",
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	// 3. Вызов сервиса для добавления предметов
	operationIDs, err := h.inventoryService.AddItems(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to add items",
			"user_id", req.UserID,
			"operation_id", req.OperationID,
			"error", err)

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to add items",
		})
		return
	}

	// 4. Возврат OperationResponse
	response := models.OperationResponse{
		Success:      true,
		OperationIDs: operationIDs,
	}

	h.logger.Info("Successfully added items",
		"user_id", req.UserID,
		"section", req.Section,
		"operations_count", len(operationIDs))

	c.JSON(http.StatusOK, response)
}

// GetReservationStatus handles GET /inventory/reservation/{operationID}
func (h *InventoryHandler) GetReservationStatus(c *gin.Context) {
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
