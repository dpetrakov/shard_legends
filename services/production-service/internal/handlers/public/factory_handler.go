package public

import (
	"encoding/json"
	"net/http"

	"strings"

	"github.com/google/uuid"
	"github.com/shard-legends/production-service/internal/auth"
	"github.com/shard-legends/production-service/internal/models"
	"github.com/shard-legends/production-service/internal/service"
	"go.uber.org/zap"
)

// FactoryHandler обрабатывает запросы к factory эндпоинтам
type FactoryHandler struct {
	taskService *service.TaskService
	logger      *zap.Logger
}

// NewFactoryHandler создает новый экземпляр FactoryHandler
func NewFactoryHandler(taskService *service.TaskService, logger *zap.Logger) *FactoryHandler {
	return &FactoryHandler{
		taskService: taskService,
		logger:      logger,
	}
}

// GetQueue обрабатывает GET /factory/queue
func (h *FactoryHandler) GetQueue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем userID из контекста аутентификации
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, models.ErrorCodeMissingUserID, "User ID not found in context")
		return
	}

	// Получаем очередь заданий
	tasks, err := h.taskService.GetUserQueue(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get user queue", zap.Error(err), zap.String("userID", userID.String()))
		h.writeError(w, http.StatusInternalServerError, models.ErrorCodeInternalError, "Failed to get queue")
		return
	}

	// Конвертируем в публичный формат (без output_items)
	publicTasks := h.convertToPublicTasks(tasks)

	// Логируем для отладки
	h.logger.Error("DEBUG: Converting tasks to public format",
		zap.Int("original_tasks_count", len(tasks)),
		zap.Int("public_tasks_count", len(publicTasks)))

	// Проверим первый элемент если есть
	if len(tasks) > 0 {
		h.logger.Error("DEBUG: Original task has output_items",
			zap.Int("output_items_count", len(tasks[0].OutputItems)))
	}
	if len(publicTasks) > 0 {
		h.logger.Error("DEBUG: Public task type",
			zap.String("task_id", publicTasks[0].ID.String()))
	}

	// Рассчитываем информацию о слотах с данными от User Service
	slotInfo := h.taskService.CalculateSlotInfoWithUserService(ctx, userID, tasks)

	response := models.QueueResponse{
		Tasks:          publicTasks,
		AvailableSlots: slotInfo,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetCompleted обрабатывает GET /factory/completed
func (h *FactoryHandler) GetCompleted(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем userID из контекста аутентификации
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, models.ErrorCodeMissingUserID, "User ID not found in context")
		return
	}

	// Получаем завершенные задания
	tasks, err := h.taskService.GetCompletedTasks(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get completed tasks", zap.Error(err), zap.String("userID", userID.String()))
		h.writeError(w, http.StatusInternalServerError, models.ErrorCodeInternalError, "Failed to get completed tasks")
		return
	}

	response := models.CompletedTasksResponse{
		Tasks: tasks,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// StartProduction обрабатывает POST /factory/start
func (h *FactoryHandler) StartProduction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем userID из контекста аутентификации
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, models.ErrorCodeMissingUserID, "User ID not found in context")
		return
	}

	// Парсим запрос
	var request models.StartProductionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, models.ErrorCodeBadRequest, "Invalid JSON body")
		return
	}

	// Валидируем запрос
	if request.RecipeID == uuid.Nil {
		h.writeError(w, http.StatusBadRequest, models.ErrorCodeValidation, "Recipe ID is required")
		return
	}

	if request.ExecutionCount <= 0 {
		h.writeError(w, http.StatusBadRequest, models.ErrorCodeValidation, "Execution count must be positive")
		return
	}

	// Запускаем производство
	task, err := h.taskService.StartProduction(ctx, userID, request)
	if err != nil {
		h.logger.Error("Failed to start production", zap.Error(err), zap.String("userID", userID.String()), zap.String("recipeID", request.RecipeID.String()))

		// Определяем тип ошибки для корректного HTTP статуса
		statusCode := http.StatusInternalServerError
		errorCode := models.ErrorCodeInternalError

		// Детальная обработка бизнес-ошибок
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "insufficient balance") || strings.Contains(errorMsg, "insufficient_items") {
			statusCode = http.StatusBadRequest
			errorCode = models.ErrorCodeInsufficientItems
		}

		h.writeError(w, statusCode, errorCode, err.Error())
		return
	}

	// Формируем ответ
	expectedCompletion := ""
	if task.CompletionTime != nil {
		expectedCompletion = task.CompletionTime.Format("2006-01-02T15:04:05Z07:00")
	}

	response := models.StartProductionResponse{
		Success:                true,
		TaskID:                 task.ID,
		Status:                 task.Status,
		ExpectedCompletionTime: expectedCompletion,
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// Claim обрабатывает POST /factory/claim
func (h *FactoryHandler) Claim(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем userID из контекста аутентификации
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, models.ErrorCodeMissingUserID, "User ID not found in context")
		return
	}

	// Парсим запрос
	var request models.ClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, models.ErrorCodeBadRequest, "Invalid JSON body")
		return
	}

	// Выполняем claim операцию
	response, err := h.taskService.ClaimTaskResults(ctx, userID, request.TaskID)
	if err != nil {
		h.logger.Error("Failed to claim task results", zap.Error(err), zap.String("userID", userID.String()))

		// Определяем тип ошибки для корректного HTTP статуса
		statusCode := http.StatusInternalServerError
		errorCode := models.ErrorCodeInternalError

		// Проверяем конкретные типы ошибок
		errorMsg := err.Error()
		if errorMsg == "task does not belong to user" {
			statusCode = http.StatusForbidden
			errorCode = models.ErrorCodeForbidden
		} else if errorMsg == "task is not in completed status" || errorMsg == "failed to get task: sql: no rows in result set" {
			statusCode = http.StatusBadRequest
			errorCode = models.ErrorCodeBadRequest
		}

		h.writeError(w, statusCode, errorCode, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

// Cancel обрабатывает POST /factory/cancel
func (h *FactoryHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем userID из контекста аутентификации
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, models.ErrorCodeMissingUserID, "User ID not found in context")
		return
	}

	// Парсим запрос
	var request models.CancelRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, models.ErrorCodeBadRequest, "Invalid JSON body")
		return
	}

	// Валидируем запрос
	if request.TaskID == uuid.Nil {
		h.writeError(w, http.StatusBadRequest, models.ErrorCodeValidation, "Task ID is required")
		return
	}

	// Выполняем cancel операцию
	err = h.taskService.CancelTask(ctx, userID, request.TaskID)
	if err != nil {
		h.logger.Error("Failed to cancel task", zap.Error(err), zap.String("userID", userID.String()), zap.String("taskID", request.TaskID.String()))

		// Определяем тип ошибки для корректного HTTP статуса
		statusCode := http.StatusInternalServerError
		errorCode := models.ErrorCodeInternalError

		// Проверяем конкретные типы ошибок
		errorMsg := err.Error()
		if errorMsg == "task does not belong to user" {
			statusCode = http.StatusForbidden
			errorCode = models.ErrorCodeForbidden
		} else if errorMsg == "task cannot be cancelled" || errorMsg == "failed to get task: sql: no rows in result set" {
			statusCode = http.StatusBadRequest
			errorCode = models.ErrorCodeBadRequest
		}

		h.writeError(w, statusCode, errorCode, err.Error())
		return
	}

	// Возвращаем успешный ответ
	response := models.OperationResponse{
		Success: true,
		Message: "Task cancelled successfully",
	}

	h.writeJSON(w, http.StatusOK, response)
}

// writeJSON отправляет JSON ответ
func (h *FactoryHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeError отправляет ошибку в JSON формате
func (h *FactoryHandler) writeError(w http.ResponseWriter, statusCode int, errorCode string, message string) {
	response := models.ErrorResponse{
		Error:   errorCode,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// convertToPublicTasks конвертирует ProductionTask в PublicProductionTask (убирает спойлеры)
func (h *FactoryHandler) convertToPublicTasks(tasks []models.ProductionTask) []models.PublicProductionTask {
	publicTasks := make([]models.PublicProductionTask, len(tasks))
	for i, task := range tasks {
		publicTasks[i] = models.PublicProductionTask{
			ID:             task.ID,
			UserID:         task.UserID,
			RecipeID:       task.RecipeID,
			SlotNumber:     task.SlotNumber,
			Status:         task.Status,
			StartedAt:      task.StartedAt,
			CompletionTime: task.CompletionTime,
			CreatedAt:      task.CreatedAt,
			UpdatedAt:      task.UpdatedAt,
			// OutputItems намеренно исключены
		}
	}
	return publicTasks
}
