package storage

import (
	"context"
	"fmt"
	"time"

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
	start := time.Now()
	defer func() {
		r.metrics.ObserveDBQueryDuration("create_task", time.Since(start))
	}()
	r.metrics.IncDBQuery("create_task")

	// Начинаем транзакцию
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Вставляем основную запись задания
	query := `
		INSERT INTO production.production_tasks (
			id, user_id, recipe_id, slot_number, execution_count, status,
			started_at, completion_time, claimed_at, pre_calculated_results, 
			modifiers_applied, reservation_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)`

	err = tx.Exec(ctx, query,
		task.ID,
		task.UserID,
		task.RecipeID,
		task.SlotNumber,
		task.ExecutionCount,
		task.Status,
		task.StartedAt,
		task.CompletionTime,
		task.ClaimedAt,
		task.PreCalculatedResults,
		task.ModifiersApplied,
		task.ReservationID,
		task.CreatedAt,
		task.UpdatedAt,
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

	return nil
}

// GetTaskByID возвращает задание по ID
func (r *taskRepository) GetTaskByID(ctx context.Context, taskID uuid.UUID) (*models.ProductionTask, error) {
	var task models.ProductionTask

	query := `
		SELECT 
			id, user_id, recipe_id, slot_number, execution_count, status,
			started_at, completion_time, claimed_at, pre_calculated_results, 
			modifiers_applied, reservation_id, created_at, updated_at
		FROM production.production_tasks
		WHERE id = $1`

	row := r.db.QueryRow(ctx, query, taskID)
	err := row.Scan(
		&task.ID,
		&task.UserID,
		&task.RecipeID,
		&task.SlotNumber,
		&task.ExecutionCount,
		&task.Status,
		&task.StartedAt,
		&task.CompletionTime,
		&task.ClaimedAt,
		&task.PreCalculatedResults,
		&task.ModifiersApplied,
		&task.ReservationID,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Загружаем выходные предметы
	outputQuery := `
		SELECT 
			toi.task_id, toi.item_id, toi.quantity,
			coll_ci.code as collection_code,
			qual_ci.code as quality_level_code,
			toi.collection_id,
			toi.quality_level_id
		FROM production.task_output_items toi
		LEFT JOIN inventory.classifier_items coll_ci ON toi.collection_id = coll_ci.id
		LEFT JOIN inventory.classifiers coll_c ON coll_ci.classifier_id = coll_c.id AND coll_c.code = 'collection'
		LEFT JOIN inventory.classifier_items qual_ci ON toi.quality_level_id = qual_ci.id  
		LEFT JOIN inventory.classifiers qual_c ON qual_ci.classifier_id = qual_c.id AND qual_c.code = 'quality_level'
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
			t.id, t.user_id, t.recipe_id, t.slot_number, t.execution_count, t.status,
			t.started_at, t.completion_time, t.claimed_at, t.pre_calculated_results, 
			t.modifiers_applied, t.reservation_id, t.created_at, t.updated_at
		FROM production.production_tasks t
		WHERE t.user_id = $1`

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
			&task.SlotNumber,
			&task.ExecutionCount,
			&task.Status,
			&task.StartedAt,
			&task.CompletionTime,
			&task.ClaimedAt,
			&task.PreCalculatedResults,
			&task.ModifiersApplied,
			&task.ReservationID,
			&task.CreatedAt,
			&task.UpdatedAt,
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
				coll_ci.code as collection_code,
				qual_ci.code as quality_level_code,
				toi.collection_id,
				toi.quality_level_id
			FROM production.task_output_items toi
			LEFT JOIN inventory.classifier_items coll_ci ON toi.collection_id = coll_ci.id
			LEFT JOIN inventory.classifiers coll_c ON coll_ci.classifier_id = coll_c.id AND coll_c.code = 'collection'
			LEFT JOIN inventory.classifier_items qual_ci ON toi.quality_level_id = qual_ci.id  
			LEFT JOIN inventory.classifiers qual_c ON qual_ci.classifier_id = qual_c.id AND qual_c.code = 'quality_level'
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
		UPDATE production.production_tasks 
		SET status = $2`

	// Добавляем специфичные поля в зависимости от статуса
	switch status {
	case models.TaskStatusInProgress:
		query += ", started_at = CURRENT_TIMESTAMP"
	case models.TaskStatusCompleted:
		query += ", completion_time = CURRENT_TIMESTAMP"
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

// DeleteTask удаляет задание (для компенсации в Saga pattern)
func (r *taskRepository) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	// Начинаем транзакцию
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Удаляем выходные предметы задания (CASCADE должен сработать, но лучше быть явным)
	outputQuery := `DELETE FROM production.task_output_items WHERE task_id = $1`
	err = tx.Exec(ctx, outputQuery, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task output items: %w", err)
	}

	// Удаляем само задание
	taskQuery := `DELETE FROM production.production_tasks WHERE id = $1`
	err = tx.Exec(ctx, taskQuery, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Инкрементируем метрики
	r.metrics.IncDBQuery("task_delete")

	return nil
}

// GetOrphanedDraftTasks возвращает задания в статусе DRAFT старше указанного времени
func (r *taskRepository) GetOrphanedDraftTasks(ctx context.Context, olderThan time.Time) ([]models.ProductionTask, error) {
	query := `
		SELECT 
			t.id, t.user_id, t.recipe_id, t.slot_number, t.status,
			t.started_at, t.completion_time, t.claimed_at, t.pre_calculated_results,
			t.modifiers_applied, t.reservation_id, t.created_at, t.updated_at
		FROM production.production_tasks t
		WHERE t.status = $1 AND t.created_at < $2
		ORDER BY t.created_at ASC`

	rows, err := r.db.Query(ctx, query, models.TaskStatusDraft, olderThan)
	if err != nil {
		return nil, fmt.Errorf("failed to query orphaned draft tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.ProductionTask
	for rows.Next() {
		var task models.ProductionTask
		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.RecipeID,
			&task.SlotNumber,
			&task.Status,
			&task.StartedAt,
			&task.CompletionTime,
			&task.ClaimedAt,
			&task.PreCalculatedResults,
			&task.ModifiersApplied,
			&task.ReservationID,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan orphaned draft task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate orphaned draft tasks: %w", err)
	}

	// Инкрементируем метрики
	r.metrics.IncDBQuery("orphaned_draft_tasks_query")

	return tasks, nil
}

// GetTasksStats возвращает статистику заданий для административной панели
func (r *taskRepository) GetTasksStats(ctx context.Context, filters map[string]interface{}, page, limit int) (*models.TasksStatsResponse, error) {
	// TODO: реализовать в рамках задачи D-7
	return nil, fmt.Errorf("GetTasksStats not implemented yet")
}

// UpdateTaskSlotNumber обновляет номер слота задания
func (r *taskRepository) UpdateTaskSlotNumber(ctx context.Context, taskID uuid.UUID, slotNumber int) error {
	query := `
		UPDATE production.production_tasks 
		SET slot_number = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	if err := r.db.Exec(ctx, query, taskID, slotNumber); err != nil {
		return fmt.Errorf("failed to update task slot number: %w", err)
	}

	// Инкрементируем метрики
	r.metrics.IncDBQuery("task_slot_update")

	return nil
}
