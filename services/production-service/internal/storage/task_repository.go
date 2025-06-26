package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shard-legends/production-service/internal/models"
)

// taskRepository реализует TaskRepository
type taskRepository struct {
	db      DatabaseInterface
	cache   CacheInterface
	metrics MetricsInterface
}

// NewTaskRepository создает новый экземпляр репозитория заданий
func NewTaskRepository(deps *RepositoryDependencies) TaskRepository {
	return &taskRepository{
		db:      deps.DB,
		cache:   deps.Cache,
		metrics: deps.MetricsCollector,
	}
}

// CreateTask создает новое производственное задание
func (r *taskRepository) CreateTask(ctx context.Context, task *models.ProductionTask) error {
	// Начинаем транзакцию
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Вставляем основную запись задания
	query := `
		INSERT INTO production.tasks (
			id, user_id, recipe_id, operation_class_code, status,
			production_time_seconds, created_at, started_at, completed_at, applied_modifiers
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)`

	err = tx.Exec(ctx, query,
		task.ID,
		task.UserID,
		task.RecipeID,
		task.OperationClassCode,
		task.Status,
		task.ProductionTimeSeconds,
		task.CreatedAt,
		task.StartedAt,
		task.CompletedAt,
		task.AppliedModifiers,
	)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	// Вставляем выходные предметы, если они есть
	if len(task.OutputItems) > 0 {
		for _, item := range task.OutputItems {
			outputQuery := `
				INSERT INTO production.task_output_items (
					task_id, item_id, collection_id, quality_level_id, quantity
				) VALUES (
					$1, $2, $3, $4, $5
				)`

			err = tx.Exec(ctx, outputQuery,
				task.ID,
				item.ItemID,
				item.CollectionID,
				item.QualityLevelID,
				item.Quantity,
			)
			if err != nil {
				return fmt.Errorf("failed to insert output item: %w", err)
			}
		}
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Инкрементируем метрики
	r.metrics.IncDBQuery("task_create")

	return nil
}

// GetTaskByID возвращает задание по ID
func (r *taskRepository) GetTaskByID(ctx context.Context, taskID uuid.UUID) (*models.ProductionTask, error) {
	var task models.ProductionTask

	query := `
		SELECT 
			id, user_id, recipe_id, operation_class_code, status,
			production_time_seconds, created_at, started_at, completed_at, applied_modifiers
		FROM production.tasks
		WHERE id = $1`

	row := r.db.QueryRow(ctx, query, taskID)
	err := row.Scan(
		&task.ID,
		&task.UserID,
		&task.RecipeID,
		&task.OperationClassCode,
		&task.Status,
		&task.ProductionTimeSeconds,
		&task.CreatedAt,
		&task.StartedAt,
		&task.CompletedAt,
		&task.AppliedModifiers,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Загружаем выходные предметы
	outputQuery := `
		SELECT 
			toi.task_id, toi.item_id, toi.quantity,
			ic.code as collection_code,
			ql.code as quality_level_code,
			toi.collection_id,
			toi.quality_level_id
		FROM production.task_output_items toi
		LEFT JOIN classifiers.item_collections ic ON toi.collection_id = ic.id
		LEFT JOIN classifiers.quality_levels ql ON toi.quality_level_id = ql.id
		WHERE toi.task_id = $1`

	rows, err := r.db.Query(ctx, outputQuery, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get output items: %w", err)
	}
	defer rows.Close()

	task.OutputItems = make([]models.TaskOutputItem, 0)
	for rows.Next() {
		var item models.TaskOutputItem
		err = rows.Scan(
			&item.TaskID,
			&item.ItemID,
			&item.Quantity,
			&item.CollectionCode,
			&item.QualityLevelCode,
			&item.CollectionID,
			&item.QualityLevelID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan output item: %w", err)
		}
		task.OutputItems = append(task.OutputItems, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate output items: %w", err)
	}

	return &task, nil
}

// GetUserTasks возвращает задания пользователя с фильтрацией
func (r *taskRepository) GetUserTasks(ctx context.Context, userID uuid.UUID, statuses []string) ([]models.ProductionTask, error) {
	var tasks []models.ProductionTask

	// Базовый запрос
	query := `
		SELECT 
			id, user_id, recipe_id, operation_class_code, status,
			production_time_seconds, created_at, started_at, completed_at, applied_modifiers
		FROM production.tasks
		WHERE user_id = $1`

	args := []interface{}{userID}

	// Добавляем фильтр по статусам, если указан
	if len(statuses) > 0 {
		query += " AND status = ANY($2)"
		args = append(args, statuses)
	}

	// Сортируем по дате создания (новые первые)
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tasks: %w", err)
	}
	defer rows.Close()

	tasks = make([]models.ProductionTask, 0)
	for rows.Next() {
		var task models.ProductionTask
		err = rows.Scan(
			&task.ID,
			&task.UserID,
			&task.RecipeID,
			&task.OperationClassCode,
			&task.Status,
			&task.ProductionTimeSeconds,
			&task.CreatedAt,
			&task.StartedAt,
			&task.CompletedAt,
			&task.AppliedModifiers,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tasks: %w", err)
	}

	// Для каждого задания загружаем выходные предметы
	for i := range tasks {
		outputQuery := `
			SELECT 
				toi.task_id, toi.item_id, toi.quantity,
				ic.code as collection_code,
				ql.code as quality_level_code,
				toi.collection_id,
				toi.quality_level_id
			FROM production.task_output_items toi
			LEFT JOIN classifiers.item_collections ic ON toi.collection_id = ic.id
			LEFT JOIN classifiers.quality_levels ql ON toi.quality_level_id = ql.id
			WHERE toi.task_id = $1`

		outputRows, err := r.db.Query(ctx, outputQuery, tasks[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get output items for task %s: %w", tasks[i].ID, err)
		}

		tasks[i].OutputItems = make([]models.TaskOutputItem, 0)
		for outputRows.Next() {
			var item models.TaskOutputItem
			err = outputRows.Scan(
				&item.TaskID,
				&item.ItemID,
				&item.Quantity,
				&item.CollectionCode,
				&item.QualityLevelCode,
				&item.CollectionID,
				&item.QualityLevelID,
			)
			if err != nil {
				outputRows.Close()
				return nil, fmt.Errorf("failed to scan output item: %w", err)
			}
			tasks[i].OutputItems = append(tasks[i].OutputItems, item)
		}
		outputRows.Close()

		if err = outputRows.Err(); err != nil {
			return nil, fmt.Errorf("failed to iterate output items: %w", err)
		}
	}

	return tasks, nil
}

// UpdateTaskStatus обновляет статус задания
func (r *taskRepository) UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status string) error {
	query := `
		UPDATE production.tasks 
		SET status = $2`

	// Добавляем специфичные поля в зависимости от статуса
	switch status {
	case models.TaskStatusInProgress:
		query += ", started_at = CURRENT_TIMESTAMP"
	case models.TaskStatusCompleted:
		query += ", completed_at = CURRENT_TIMESTAMP"
	}

	query += " WHERE id = $1"

	err := r.db.Exec(ctx, query, taskID, status)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Инкрементируем метрики
	r.metrics.IncDBQuery("task_status_update")

	return nil
}

// GetTasksStats возвращает статистику заданий для административной панели
func (r *taskRepository) GetTasksStats(ctx context.Context, filters map[string]interface{}, page, limit int) (*models.TasksStatsResponse, error) {
	// TODO: реализовать в рамках задачи D-7
	return nil, fmt.Errorf("GetTasksStats not implemented yet")
}
