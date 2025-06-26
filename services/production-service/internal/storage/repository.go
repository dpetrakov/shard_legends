package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shard-legends/production-service/internal/models"
)

// RecipeRepository определяет интерфейс для работы с производственными рецептами
type RecipeRepository interface {
	// GetActiveRecipes возвращает список активных рецептов с возможностью фильтрации
	GetActiveRecipes(ctx context.Context, filters *models.RecipeFilters) ([]models.ProductionRecipe, error)
	
	// GetRecipeByID возвращает рецепт по ID с полной информацией
	GetRecipeByID(ctx context.Context, recipeID uuid.UUID) (*models.ProductionRecipe, error)
	
	// GetRecipeLimits возвращает все лимиты для рецепта
	GetRecipeLimits(ctx context.Context, recipeID uuid.UUID) ([]models.RecipeLimit, error)
	
	// GetRecipeUsageStats возвращает статистику использования рецепта пользователем
	GetRecipeUsageStats(ctx context.Context, userID uuid.UUID, recipeID uuid.UUID, limitType string, limitObject string, targetItemID *uuid.UUID, periodStart, periodEnd time.Time) (*models.RecipeUsageStats, error)
	
	// CheckRecipeLimits проверяет все лимиты рецепта для пользователя
	CheckRecipeLimits(ctx context.Context, userID uuid.UUID, recipeID uuid.UUID, requestedExecutions int) ([]models.UserRecipeLimit, error)
}

// TaskRepository определяет интерфейс для работы с производственными заданиями
type TaskRepository interface {
	// CreateTask создает новое производственное задание
	CreateTask(ctx context.Context, task *models.ProductionTask) error
	
	// GetTaskByID возвращает задание по ID
	GetTaskByID(ctx context.Context, taskID uuid.UUID) (*models.ProductionTask, error)
	
	// GetUserTasks возвращает задания пользователя с фильтрацией
	GetUserTasks(ctx context.Context, userID uuid.UUID, statuses []string) ([]models.ProductionTask, error)
	
	// UpdateTaskStatus обновляет статус задания
	UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status string) error
	
	// GetTasksStats возвращает статистику заданий для административной панели
	GetTasksStats(ctx context.Context, filters map[string]interface{}, page, limit int) (*models.TasksStatsResponse, error)
}

// ClassifierRepository определяет интерфейс для работы с классификаторами
type ClassifierRepository interface {
	// GetClassifierMapping возвращает маппинг код -> UUID для классификатора
	GetClassifierMapping(ctx context.Context, classifierName string) (map[string]uuid.UUID, error)
	
	// GetReverseClassifierMapping возвращает маппинг UUID -> код для классификатора
	GetReverseClassifierMapping(ctx context.Context, classifierName string) (map[uuid.UUID]string, error)
	
	// ConvertCodeToUUID преобразует код в UUID через классификатор
	ConvertCodeToUUID(ctx context.Context, classifierName, code string) (*uuid.UUID, error)
	
	// ConvertUUIDToCode преобразует UUID в код через классификатор
	ConvertUUIDToCode(ctx context.Context, classifierName string, id uuid.UUID) (*string, error)
}

// Repository объединяет все репозитории
type Repository struct {
	Recipe     RecipeRepository
	Task       TaskRepository
	Classifier ClassifierRepository
}

// RepositoryDependencies содержит зависимости для создания репозиториев
type RepositoryDependencies struct {
	DB           DatabaseInterface
	Cache        CacheInterface
	MetricsCollector MetricsInterface
}

// DatabaseInterface определяет интерфейс для работы с базой данных
type DatabaseInterface interface {
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	Exec(ctx context.Context, query string, args ...interface{}) error
	BeginTx(ctx context.Context) (Tx, error)
	Health(ctx context.Context) error
}

// CacheInterface определяет интерфейс для работы с кешем
type CacheInterface interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	Health(ctx context.Context) error
}

// MetricsInterface определяет интерфейс для сбора метрик
type MetricsInterface interface {
	IncDBQuery(operation string)
	IncCacheHit(cacheType string)
	IncCacheMiss(cacheType string)
	ObserveDBQueryDuration(operation string, duration time.Duration)
}

// Row интерфейс для работы с результатом одной строки
type Row interface {
	Scan(dest ...interface{}) error
}

// Rows интерфейс для работы с результатом множества строк
type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
	Close()
}

// Tx интерфейс для работы с транзакциями
type Tx interface {
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	Exec(ctx context.Context, query string, args ...interface{}) error
	Commit() error
	Rollback() error
}