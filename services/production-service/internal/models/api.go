package models

import (
	"github.com/google/uuid"
)

// RecipesResponse представляет ответ эндпоинта GET /recipes
type RecipesResponse struct {
	Recipes []ProductionRecipeWithLimits `json:"recipes"`
}

// ProductionRecipeWithLimits расширяет ProductionRecipe информацией о лимитах пользователя
type ProductionRecipeWithLimits struct {
	ProductionRecipe
	UserLimits []UserRecipeLimit `json:"user_limits,omitempty"`
}

// UserRecipeLimit представляет информацию о лимите пользователя для конкретного рецепта
type UserRecipeLimit struct {
	LimitType      string    `json:"limit_type"`
	LimitObject    string    `json:"limit_object"`
	TargetItemID   *uuid.UUID `json:"target_item_id,omitempty"`
	CurrentUsage   int       `json:"current_usage"`
	MaxAllowed     int       `json:"max_allowed"`
	IsExceeded     bool      `json:"is_exceeded"`
	ResetTime      *string   `json:"reset_time,omitempty"` // ISO 8601 format
}

// StartProductionRequest представляет запрос POST /factory/start
type StartProductionRequest struct {
	RecipeID   uuid.UUID   `json:"recipe_id" validate:"required"`
	Executions int         `json:"executions" validate:"required,min=1"`
	Boosters   []uuid.UUID `json:"boosters,omitempty"`
}

// StartProductionResponse представляет ответ POST /factory/start
type StartProductionResponse struct {
	Success                bool      `json:"success"`
	TaskID                 uuid.UUID `json:"task_id"`
	Status                 string    `json:"status"`
	ExpectedCompletionTime string    `json:"expected_completion_time"` // ISO 8601 format
}

// ClaimRequest представляет запрос POST /factory/claim
type ClaimRequest struct {
	TaskID *uuid.UUID `json:"task_id,omitempty"` // если nil - claim всех готовых
}

// ClaimResponse представляет ответ POST /factory/claim
type ClaimResponse struct {
	Success            bool                 `json:"success"`
	ItemsReceived      []TaskOutputItem     `json:"items_received"`
	UpdatedQueueStatus *QueueResponse       `json:"updated_queue_status,omitempty"`
}

// QueueResponse представляет ответ GET /factory/queue
type QueueResponse struct {
	Tasks          []ProductionTask `json:"tasks"`
	AvailableSlots SlotInfo         `json:"available_slots"`
}

// SlotInfo представляет информацию о доступных слотах
type SlotInfo struct {
	Total int `json:"total"`
	Used  int `json:"used"`
	Free  int `json:"free"`
}

// CompletedTasksResponse представляет ответ GET /factory/completed
type CompletedTasksResponse struct {
	Tasks []ProductionTask `json:"tasks"`
}

// CancelRequest представляет запрос POST /factory/cancel
type CancelRequest struct {
	TaskID uuid.UUID `json:"task_id" validate:"required"`
}

// OperationResponse представляет стандартный ответ операции
type OperationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// UserSlotsResponse представляет ответ GET /internal/user-slots/{user_id}
type UserSlotsResponse struct {
	UserID uuid.UUID `json:"user_id"`
	Slots  []Slot    `json:"slots"`
}

// Slot представляет производственный слот пользователя
type Slot struct {
	SlotType            string   `json:"slot_type"`
	SupportedOperations []string `json:"supported_operations"`
	Count               int      `json:"count"`
}

// TasksStatsResponse представляет ответ GET /admin/tasks
type TasksStatsResponse struct {
	Tasks      []ProductionTask `json:"tasks"`
	Stats      TaskStats        `json:"stats"`
	Pagination Pagination       `json:"pagination"`
}

// TaskStats представляет статистику заданий
type TaskStats struct {
	TotalTasks          int                    `json:"total_tasks"`
	ByStatus            map[string]int         `json:"by_status"`
	ByOperationClass    map[string]int         `json:"by_operation_class"`
	SystemLoad          SystemLoad             `json:"system_load"`
}

// SystemLoad представляет загрузку системы
type SystemLoad struct {
	AverageQueueLength       float64 `json:"average_queue_length"`
	SlotUtilizationPercent   float64 `json:"slot_utilization_percent"`
}

// Pagination представляет информацию о пагинации
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
	TotalItems int `json:"total_items"`
}

// ErrorResponse представляет стандартный ответ с ошибкой
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ValidationError представляет ошибку валидации
type ValidationError struct {
	Error            string                 `json:"error"`
	Message          string                 `json:"message"`
	ValidationErrors []ValidationFieldError `json:"validation_errors,omitempty"`
	MissingItems     []MissingItem          `json:"missing_items,omitempty"`
}

// ValidationFieldError представляет ошибку валидации поля
type ValidationFieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// MissingItem представляет недостающий предмет
type MissingItem struct {
	ItemID           uuid.UUID `json:"item_id"`
	CollectionCode   *string   `json:"collection_code,omitempty"`
	QualityLevelCode *string   `json:"quality_level_code,omitempty"`
	Required         int       `json:"required"`
	Available        int       `json:"available"`
}

// LimitExceededError представляет ошибку превышения лимита
type LimitExceededError struct {
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	LimitInfo LimitInfo `json:"limit_info"`
}

// LimitInfo представляет информацию о превышенном лимите
type LimitInfo struct {
	LimitType    string `json:"limit_type"`
	CurrentUsage int    `json:"current_usage"`
	MaxAllowed   int    `json:"max_allowed"`
	ResetTime    string `json:"reset_time"` // ISO 8601 format
}

// Constants для типов слотов
const (
	SlotTypeUniversal   = "universal"
	SlotTypeSpecialized = "specialized"
)

// Constants для ошибок
const (
	ErrorCodeValidation       = "validation_error"
	ErrorCodeInsufficientItems = "insufficient_items"
	ErrorCodeLimitExceeded    = "limit_exceeded"
	ErrorCodeMissingToken     = "missing_token"
	ErrorCodeInvalidToken     = "invalid_token_format"
	ErrorCodeTokenSignature   = "invalid_token_signature"
	ErrorCodeTokenRevoked     = "token_revoked"
	ErrorCodeMissingUserID    = "missing_user_id"
	ErrorCodeForbidden        = "forbidden"
	ErrorCodeNotFound         = "not_found"
	ErrorCodeBadRequest       = "bad_request"
	ErrorCodeInternalError    = "internal_error"
)