package public

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/shard-legends/production-service/internal/auth"
	"github.com/shard-legends/production-service/internal/models"
	"github.com/shard-legends/production-service/internal/service"
	"go.uber.org/zap"
)

// RecipeHandler обрабатывает HTTP запросы для работы с рецептами
type RecipeHandler struct {
	recipeService service.RecipeService
	logger        *zap.Logger
	validator     *validator.Validate
}

// NewRecipeHandler создает новый экземпляр RecipeHandler
func NewRecipeHandler(recipeService service.RecipeService, logger *zap.Logger) *RecipeHandler {
	return &RecipeHandler{
		recipeService: recipeService,
		logger:        logger,
		validator:     validator.New(),
	}
}

// GetRecipes обрабатывает GET /recipes
func (h *RecipeHandler) GetRecipes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем информацию о пользователе из JWT
	userContext, err := auth.GetUser(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, models.ErrorCodeMissingUserID, "User ID not found in context", nil)
		return
	}

	userID, err := uuid.Parse(userContext.UserID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, models.ErrorCodeBadRequest, "Invalid user ID format", nil)
		return
	}

	h.logger.Info("Getting recipes for user",
		zap.String("user_id", userContext.UserID),
		zap.String("request_id", getRequestID(r)),
	)

	// Парсим фильтры из query параметров
	filters := &models.RecipeFilters{}

	// Фильтр по классу операции
	if operationClass := r.URL.Query().Get("operation_class"); operationClass != "" {
		// Валидируем класс операции
		if !isValidOperationClass(operationClass) {
			h.writeErrorResponse(w, http.StatusBadRequest, models.ErrorCodeValidation,
				"Invalid operation class", map[string]interface{}{
					"allowed_values": []string{
						models.OperationClassCrafting,
						models.OperationClassSmelting,
						models.OperationClassChestOpening,
						models.OperationClassResourceGathering,
						models.OperationClassSpecial,
					},
				})
			return
		}
		filters.OperationClassCode = &operationClass
	}

	// Добавляем пользователя в фильтры для проверки лимитов
	filters.UserID = &userID

	// Получаем рецепты от сервиса
	recipes, err := h.recipeService.GetRecipesForUser(ctx, userID, filters)
	if err != nil {
		h.logger.Error("Failed to get recipes",
			zap.Error(err),
			zap.String("user_id", userContext.UserID),
			zap.String("request_id", getRequestID(r)),
		)
		h.writeErrorResponse(w, http.StatusInternalServerError, models.ErrorCodeInternalError,
			"Failed to get recipes", nil)
		return
	}

	// Формируем ответ
	response := models.RecipesResponse{
		Recipes: recipes,
	}

	h.writeJSONResponse(w, http.StatusOK, response)

	h.logger.Info("Recipes retrieved successfully",
		zap.String("user_id", userContext.UserID),
		zap.Int("recipes_count", len(recipes)),
		zap.String("request_id", getRequestID(r)),
	)
}

// isValidOperationClass проверяет валидность класса операции
func isValidOperationClass(operationClass string) bool {
	validClasses := map[string]bool{
		models.OperationClassCrafting:          true,
		models.OperationClassSmelting:          true,
		models.OperationClassChestOpening:      true,
		models.OperationClassResourceGathering: true,
		models.OperationClassSpecial:           true,
	}
	return validClasses[operationClass]
}

// writeJSONResponse отправляет JSON ответ
func (h *RecipeHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

// writeErrorResponse отправляет JSON ответ с ошибкой
func (h *RecipeHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string, details map[string]interface{}) {
	errorResponse := models.ErrorResponse{
		Error:   errorCode,
		Message: message,
		Details: details,
	}

	h.writeJSONResponse(w, statusCode, errorResponse)
}

// getRequestID извлекает request ID из заголовков или контекста
func getRequestID(r *http.Request) string {
	// Пытаемся получить из middleware
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}

	// Fallback - генерируем новый
	return "unknown"
}
