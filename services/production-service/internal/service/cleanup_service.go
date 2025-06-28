package service

import (
	"context"
	"time"

	"github.com/shard-legends/production-service/internal/models"
	"github.com/shard-legends/production-service/internal/storage"
	"go.uber.org/zap"
)

// CleanupService отвечает за очистку orphaned задач и поддержание целостности данных
type CleanupService struct {
	taskRepo        storage.TaskRepository
	inventoryClient InventoryClient
	logger          *zap.Logger
	config          CleanupConfig
}

// CleanupConfig конфигурация для сервиса очистки
type CleanupConfig struct {
	// OrphanedTaskTimeout время после которого DRAFT задачи считаются orphaned
	OrphanedTaskTimeout time.Duration
	// CleanupInterval интервал запуска очистки
	CleanupInterval time.Duration
}

// NewCleanupService создает новый сервис очистки
func NewCleanupService(
	taskRepo storage.TaskRepository,
	inventoryClient InventoryClient,
	logger *zap.Logger,
	config CleanupConfig,
) *CleanupService {
	return &CleanupService{
		taskRepo:        taskRepo,
		inventoryClient: inventoryClient,
		logger:          logger,
		config:          config,
	}
}

// Start запускает фоновый процесс очистки
func (s *CleanupService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	s.logger.Info("Starting cleanup service",
		zap.Duration("interval", s.config.CleanupInterval),
		zap.Duration("orphaned_timeout", s.config.OrphanedTaskTimeout))

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping cleanup service")
			return
		case <-ticker.C:
			s.runCleanup(ctx)
		}
	}
}

// runCleanup выполняет одну итерацию очистки orphaned задач
func (s *CleanupService) runCleanup(ctx context.Context) {
	cleanupCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	startTime := time.Now()
	cutoffTime := startTime.Add(-s.config.OrphanedTaskTimeout)

	s.logger.Debug("Starting cleanup run", zap.Time("cutoff_time", cutoffTime))

	// Получаем orphaned draft задачи
	orphanedTasks, err := s.taskRepo.GetOrphanedDraftTasks(cleanupCtx, cutoffTime)
	if err != nil {
		s.logger.Error("Failed to get orphaned draft tasks", zap.Error(err))
		return
	}

	if len(orphanedTasks) == 0 {
		s.logger.Debug("No orphaned tasks found")
		return
	}

	s.logger.Info("Found orphaned draft tasks", zap.Int("count", len(orphanedTasks)))

	cleanedCount := 0
	for _, task := range orphanedTasks {
		if s.cleanupOrphanedTask(cleanupCtx, task) {
			cleanedCount++
		}
	}

	duration := time.Since(startTime)
	s.logger.Info("Cleanup run completed",
		zap.Int("orphaned_found", len(orphanedTasks)),
		zap.Int("cleaned_count", cleanedCount),
		zap.Duration("duration", duration))
}

// cleanupOrphanedTask очищает одну orphaned задачу
func (s *CleanupService) cleanupOrphanedTask(ctx context.Context, task models.ProductionTask) bool {
	taskLogger := s.logger.With(
		zap.String("task_id", task.ID.String()),
		zap.String("user_id", task.UserID.String()),
		zap.String("recipe_id", task.RecipeID.String()),
		zap.Time("created_at", task.CreatedAt))

	taskLogger.Info("Cleaning up orphaned draft task")

	// Проверяем, есть ли резервирование в Inventory Service
	// Если есть - возвращаем его
	err := s.inventoryClient.ReturnReserve(ctx, task.UserID, task.ID)
	if err != nil {
		// Логируем, но не блокируем удаление задачи
		// Резервирование могло и не существовать
		taskLogger.Warn("Failed to return inventory reservation (might not exist)",
			zap.Error(err))
	} else {
		taskLogger.Info("Returned inventory reservation for orphaned task")
	}

	// Удаляем orphaned задачу
	err = s.taskRepo.DeleteTask(ctx, task.ID)
	if err != nil {
		taskLogger.Error("Failed to delete orphaned task", zap.Error(err))
		return false
	}

	taskLogger.Info("Successfully cleaned up orphaned draft task")
	return true
}

// GetDefaultCleanupConfig возвращает конфигурацию по умолчанию
func GetDefaultCleanupConfig() CleanupConfig {
	return CleanupConfig{
		OrphanedTaskTimeout: 5 * time.Minute, // 5 минут для обнаружения orphaned задач
		CleanupInterval:     5 * time.Minute, // Запуск каждые 5 минут
	}
}