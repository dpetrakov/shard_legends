package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/shard-legends/production-service/internal/models"
	"github.com/shard-legends/production-service/internal/storage"
)

// TaskService реализует бизнес-логику для работы с производственными заданиями
type TaskService struct {
	taskRepo        storage.TaskRepository
	recipeRepo      storage.RecipeRepository
	classifierRepo  storage.ClassifierRepository
	codeConverter   CodeConverterService
	inventoryClient InventoryClient
	userClient      UserClient
	modifierService *ModifierService
	calculator      *ProductionCalculator
	logger          *zap.Logger
}

// NewTaskService создает новый экземпляр TaskService
func NewTaskService(
	taskRepo storage.TaskRepository,
	recipeRepo storage.RecipeRepository,
	classifierRepo storage.ClassifierRepository,
	codeConverter CodeConverterService,
	inventoryClient InventoryClient,
	userClient UserClient,
	logger *zap.Logger,
) *TaskService {
	modifierService := NewModifierService(userClient, logger)
	calculator := NewProductionCalculator(classifierRepo, modifierService, logger)

	return &TaskService{
		taskRepo:        taskRepo,
		recipeRepo:      recipeRepo,
		classifierRepo:  classifierRepo,
		codeConverter:   codeConverter,
		inventoryClient: inventoryClient,
		userClient:      userClient,
		modifierService: modifierService,
		calculator:      calculator,
		logger:          logger,
	}
}

// GetUserQueue возвращает очередь производственных заданий пользователя
func (s *TaskService) GetUserQueue(ctx context.Context, userID uuid.UUID) ([]models.ProductionTask, error) {
	s.logger.Error("DEBUG: GetUserQueue called", zap.String("userID", userID.String()))

	// Для очереди возвращаем только pending и in_progress, completed показываются в отдельном эндпоинте
	statuses := []string{
		models.TaskStatusPending,
		models.TaskStatusInProgress,
	}

	tasks, err := s.taskRepo.GetUserTasks(ctx, userID, statuses)
	if err != nil {
		s.logger.Error("Failed to get user tasks", zap.Error(err), zap.String("userID", userID.String()))
		return nil, fmt.Errorf("failed to get user tasks: %w", err)
	}

	// Пытаемся автоматически запустить pending задачи если есть свободные слоты
	s.logger.Error("DEBUG: Starting auto-start logic", zap.String("userID", userID.String()), zap.Int("tasks_count", len(tasks)))
	err = s.tryStartPendingTasks(ctx, userID, tasks)
	if err != nil {
		s.logger.Error("Failed to auto-start pending tasks", zap.Error(err), zap.String("userID", userID.String()))
		// Не возвращаем ошибку, просто логируем - пользователь все равно должен получить очередь
	}

	// Перезагружаем задачи после возможных изменений статусов
	tasks, err = s.taskRepo.GetUserTasks(ctx, userID, statuses)
	if err != nil {
		s.logger.Error("Failed to reload user tasks", zap.Error(err), zap.String("userID", userID.String()))
		return nil, fmt.Errorf("failed to reload user tasks: %w", err)
	}

	return tasks, nil
}

// GetCompletedTasks возвращает завершенные задания пользователя
func (s *TaskService) GetCompletedTasks(ctx context.Context, userID uuid.UUID) ([]models.ProductionTask, error) {
	tasks, err := s.taskRepo.GetUserTasks(ctx, userID, []string{models.TaskStatusCompleted})
	if err != nil {
		s.logger.Error("Failed to get completed tasks", zap.Error(err), zap.String("userID", userID.String()))
		return nil, fmt.Errorf("failed to get completed tasks: %w", err)
	}

	return tasks, nil
}

// StartProduction создает новое производственное задание
func (s *TaskService) StartProduction(ctx context.Context, userID uuid.UUID, request models.StartProductionRequest) (*models.ProductionTask, error) {
	// 1. Получаем рецепт
	recipe, err := s.recipeRepo.GetRecipeByID(ctx, request.RecipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe: %w", err)
	}

	if !recipe.IsActive {
		return nil, fmt.Errorf("recipe is not active")
	}

	// 2. Проверяем лимиты рецепта
	limits, err := s.recipeRepo.CheckRecipeLimits(ctx, userID, request.RecipeID, request.ExecutionCount)
	if err != nil {
		return nil, fmt.Errorf("failed to check recipe limits: %w", err)
	}

	for _, limit := range limits {
		if limit.IsExceeded {
			return nil, fmt.Errorf("recipe limit exceeded: %s", limit.LimitType)
		}
	}

	// 3. Получаем слоты пользователя
	userSlots, err := s.userClient.GetUserProductionSlots(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user slots: %w", err)
	}

	// 4. Проверяем доступность слотов
	if !s.hasAvailableSlot(ctx, userID, recipe.OperationClassCode, userSlots) {
		return nil, fmt.Errorf("no available slots for operation class: %s", recipe.OperationClassCode)
	}

	// 5. Выполняем предрасчет производства с применением модификаторов
	calcCtx, err := s.calculator.PrecalculateProduction(ctx, userID, recipe, request)
	if err != nil {
		return nil, fmt.Errorf("failed to precalculate production: %w", err)
	}

	// 6. Рассчитываем выходные предметы
	outputItems, err := s.calculator.CalculateOutputItems(ctx, calcCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate output items: %w", err)
	}

	// 7. Получаем итоговое время производства
	finalProductionTime := s.calculator.GetModifiedProductionTime(calcCtx)

	// 8. Получаем модифицированные входные предметы для резервирования
	modifiedInputItems := s.calculator.GetModifiedInputItems(calcCtx)
	itemsToReserve, err := s.prepareItemsForReservation(ctx, modifiedInputItems, request)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare items for reservation: %w", err)
	}

	// 9. Создаем задание (статус будет установлен в createTaskWithReservation)
	task := &models.ProductionTask{
		ID:               uuid.New(),
		UserID:           userID,
		RecipeID:         recipe.ID,
		SlotNumber:       1, // placeholder, будет обновлён при запуске
		ExecutionCount:   request.ExecutionCount,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		ModifiersApplied: s.modifierService.BuildAppliedModifiersForAudit(calcCtx.ModificationResults),
		OutputItems:      outputItems,
	}

	// 10. Начинаем транзакционную операцию: создание задания + резервирование предметов
	err = s.createTaskWithReservation(ctx, task, itemsToReserve)
	if err != nil {
		return nil, err
	}

	// 11. Логируем применение модификаторов для аудита
	s.modifierService.LogModifiersApplication(userID, task.ID, calcCtx.ModificationResults)

	// 12. Проверяем, можно ли сразу запустить задание
	canStart := s.canStartImmediately(ctx, userID, recipe.OperationClassCode)
	s.logger.Error("DEBUG: Can start immediately",
		zap.Bool("can_start", canStart),
		zap.Int("final_production_time", finalProductionTime),
		zap.String("recipe_id", recipe.ID.String()))

	// Мгновенные задания (время производства 0) завершаются сразу независимо от слотов
	if finalProductionTime == 0 {
		now := time.Now()
		task.StartedAt = &now
		task.Status = models.TaskStatusCompleted
		task.CompletionTime = &now

		err = s.taskRepo.UpdateTaskStatus(ctx, task.ID, models.TaskStatusCompleted)
		if err != nil {
			s.logger.Error("Failed to complete instant task", zap.Error(err), zap.String("taskID", task.ID.String()))
		}
		s.logger.Info("Instantly completed task",
			zap.String("taskID", task.ID.String()),
			zap.String("recipe_id", recipe.ID.String()))
	} else if canStart {
		// Обычные задания запускаются только если есть свободные слоты
		now := time.Now()
		task.StartedAt = &now
		task.Status = models.TaskStatusInProgress
		completionTime := now.Add(time.Duration(finalProductionTime) * time.Second)
		task.CompletionTime = &completionTime

		err = s.taskRepo.UpdateTaskStatus(ctx, task.ID, models.TaskStatusInProgress)
		if err != nil {
			s.logger.Error("Failed to start task immediately", zap.Error(err), zap.String("taskID", task.ID.String()))
		}
	}

	return task, nil
}

// prepareItemsForReservation подготавливает список предметов для резервирования
func (s *TaskService) prepareItemsForReservation(_ context.Context, inputItems []models.RecipeInputItem, request models.StartProductionRequest) ([]models.ReservationItem, error) {
	var items []models.ReservationItem

	// Добавляем входные предметы рецепта (уже модифицированные)
	for _, input := range inputItems {
		item := models.ReservationItem{
			ItemID:           input.ItemID,
			Quantity:         input.Quantity * request.ExecutionCount,
			CollectionID:     input.CollectionID,
			QualityLevelID:   input.QualityLevelID,
			CollectionCode:   input.CollectionCode,
			QualityLevelCode: input.QualityLevelCode,
		}
		items = append(items, item)
	}

	// Добавляем ускорители
	for _, booster := range request.Boosters {
		item := models.ReservationItem{
			ItemID:   booster.ItemID,
			Quantity: booster.Quantity,
		}
		items = append(items, item)
	}

	return items, nil
}

// createTaskWithReservation создает задание с атомарным резервированием предметов (Saga Pattern)
func (s *TaskService) createTaskWithReservation(ctx context.Context, task *models.ProductionTask, itemsToReserve []models.ReservationItem) error {
	// Phase 1: Create draft task (database transaction)
	task.Status = models.TaskStatusDraft
	err := s.taskRepo.CreateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create draft task: %w", err)
	}

	// Phase 2: Reserve inventory (HTTP call with idempotency) - только если есть предметы для резервирования
	if len(itemsToReserve) > 0 {
		err = s.inventoryClient.ReserveItems(ctx, task.UserID, task.ID, itemsToReserve)
		if err != nil {
			// Compensation: Delete draft task completely
			deleteErr := s.taskRepo.DeleteTask(ctx, task.ID)
			if deleteErr != nil {
				s.logger.Error("Failed to delete draft task during compensation",
					zap.String("task_id", task.ID.String()),
					zap.Error(deleteErr))
			}
			return fmt.Errorf("failed to reserve inventory: %w", err)
		}
	}

	// Phase 3: Confirm task (database transaction)
	err = s.taskRepo.UpdateTaskStatus(ctx, task.ID, models.TaskStatusPending)
	if err != nil {
		// Compensation: Return inventory reservation and delete task - только если было резервирование
		if len(itemsToReserve) > 0 {
			returnErr := s.inventoryClient.ReturnReserve(ctx, task.UserID, task.ID)
			if returnErr != nil {
				s.logger.Error("Failed to return inventory reservation during compensation",
					zap.String("task_id", task.ID.String()),
					zap.Error(returnErr))
			}
		}

		deleteErr := s.taskRepo.DeleteTask(ctx, task.ID)
		if deleteErr != nil {
			s.logger.Error("Failed to delete draft task during compensation",
				zap.String("task_id", task.ID.String()),
				zap.Error(deleteErr))
		}

		return fmt.Errorf("failed to confirm task: %w", err)
	}

	return nil
}

// hasAvailableSlot проверяет наличие доступного слота для операции
func (s *TaskService) hasAvailableSlot(ctx context.Context, userID uuid.UUID, operationClass string, userSlots *models.UserProductionSlots) bool {
	// Подсчитываем занятые слоты
	activeTasks, err := s.taskRepo.GetUserTasks(ctx, userID, []string{models.TaskStatusInProgress})
	if err != nil {
		return false
	}

	occupiedSlots := 0
	for _, task := range activeTasks {
		// Get the recipe to determine operation class
		recipe, err := s.recipeRepo.GetRecipeByID(ctx, task.RecipeID)
		if err != nil {
			s.logger.Error("Failed to get recipe for task", zap.Error(err), zap.String("taskID", task.ID.String()))
			continue
		}
		if s.canUseSlot(recipe.OperationClassCode, operationClass, userSlots) {
			occupiedSlots++
		}
	}

	// Проверяем, есть ли свободные слоты
	totalSlots := 0
	for _, slot := range userSlots.Slots {
		if s.slotSupportsOperation(slot, operationClass) {
			totalSlots += slot.Count
		}
	}

	return occupiedSlots < totalSlots
}

// canStartImmediately проверяет, можно ли сразу запустить задание
func (s *TaskService) canStartImmediately(ctx context.Context, userID uuid.UUID, operationClass string) bool {
	// Проверяем, нет ли заданий в статусе pending для этого класса операций
	pendingTasks, err := s.taskRepo.GetUserTasks(ctx, userID, []string{models.TaskStatusPending})
	if err != nil {
		return false
	}

	for _, task := range pendingTasks {
		// Get the recipe to determine operation class
		recipe, err := s.recipeRepo.GetRecipeByID(ctx, task.RecipeID)
		if err != nil {
			s.logger.Error("Failed to get recipe for task", zap.Error(err), zap.String("taskID", task.ID.String()))
			continue
		}
		if recipe.OperationClassCode == operationClass {
			return false
		}
	}

	return true
}

// canUseSlot проверяет, может ли задание использовать слот
func (s *TaskService) canUseSlot(taskOperation, targetOperation string, userSlots *models.UserProductionSlots) bool {
	for _, slot := range userSlots.Slots {
		if s.slotSupportsOperation(slot, taskOperation) && s.slotSupportsOperation(slot, targetOperation) {
			return true
		}
	}
	return false
}

// slotSupportsOperation проверяет, поддерживает ли слот операцию
func (s *TaskService) slotSupportsOperation(slot models.ProductionSlot, operation string) bool {
	if slot.SlotType == "universal" {
		return true
	}

	return slices.Contains(slot.SupportedOperations, operation)
}

// ClaimTaskResults получает результаты завершенного производственного задания
func (s *TaskService) ClaimTaskResults(ctx context.Context, userID uuid.UUID, taskID *uuid.UUID) (*models.ClaimResponse, error) {
	var tasksToProcess []*models.ProductionTask

	if taskID != nil {
		// Claim конкретного задания
		task, err := s.taskRepo.GetTaskByID(ctx, *taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task: %w", err)
		}

		if task.UserID != userID {
			return nil, fmt.Errorf("task does not belong to user")
		}

		if task.Status != models.TaskStatusCompleted {
			return nil, fmt.Errorf("task is not in completed status: %s", task.Status)
		}

		tasksToProcess = append(tasksToProcess, task)
	} else {
// Claim всех готовых заданий
completedTasks, err := s.GetCompletedTasks(ctx, userID)
if err != nil {
    return nil, fmt.Errorf("failed to get completed tasks: %w", err)
}

for i := range completedTasks {
    tasksToProcess = append(tasksToProcess, &completedTasks[i])
}
	}

	var allItemsReceived []models.TaskOutputItem
	var failedTasks []string

	// Обрабатываем каждое задание
	for _, task := range tasksToProcess {
		err := s.processTaskClaim(ctx, task)
		if err != nil {
			s.logger.Error("Failed to process task claim",
				zap.Error(err),
				zap.String("taskID", task.ID.String()),
				zap.String("userID", userID.String()))
			
			// Для единичного клайма возвращаем ошибку
			if taskID != nil {
				return nil, fmt.Errorf("failed to process claim for task %s: %w", task.ID, err)
			}
			
			// Для массового клайма добавляем в список неудачных и продолжаем
			failedTasks = append(failedTasks, task.ID.String())
			continue
		}

		// Если нет выходных предметов, это не ошибка
		if len(task.OutputItems) == 0 {
			s.logger.Info("No output items to claim for task",
				zap.String("taskID", task.ID.String()),
				zap.String("userID", userID.String()))
		}

		// Добавляем полученные предметы к общему списку
		allItemsReceived = append(allItemsReceived, task.OutputItems...)

		// Логируем операцию claim для аудита
		s.logger.Info("Task claimed successfully",
			zap.String("taskID", task.ID.String()),
			zap.String("userID", userID.String()),
			zap.Int("itemsReceived", len(task.OutputItems)))
	}

	// Получаем обновленную очередь
	s.logger.Error("DEBUG: Getting updated queue after claim", zap.String("userID", userID.String()))
	updatedQueue, err := s.GetUserQueue(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get updated queue after claim", zap.Error(err))
		s.logger.Error("DEBUG: updatedQueue is nil due to error", zap.String("userID", userID.String()))
		// Не возвращаем ошибку, так как claim прошел успешно
	} else {
		s.logger.Error("DEBUG: Got updated queue successfully",
			zap.String("userID", userID.String()),
			zap.Int("tasks_count", len(updatedQueue)))
	}

	// DEBUG: Логируем что возвращаем
	for i, item := range allItemsReceived {
		s.logger.Error("DEBUG: ClaimResponse item",
			zap.Int("index", i),
			zap.String("item_id", item.ItemID.String()),
			zap.Any("collection_code", item.CollectionCode),
			zap.Any("quality_level_code", item.QualityLevelCode),
			zap.Int("quantity", item.Quantity))
	}

	response := &models.ClaimResponse{
		Success:       len(failedTasks) == 0,
		ItemsReceived: allItemsReceived,
		FailedTasks:   failedTasks,
	}

	if updatedQueue != nil {
		s.logger.Error("DEBUG: updatedQueue is not nil, calling CalculateSlotInfoWithUserService",
			zap.String("userID", userID.String()),
			zap.Int("tasks_count", len(updatedQueue)))
		slotInfo := s.CalculateSlotInfoWithUserService(ctx, userID, updatedQueue)
		response.UpdatedQueueStatus = &models.QueueResponse{
			Tasks:          s.convertToPublicTasks(updatedQueue),
			AvailableSlots: slotInfo,
		}
	} else {
		s.logger.Error("DEBUG: updatedQueue is nil, skipping CalculateSlotInfoWithUserService",
			zap.String("userID", userID.String()))
	}

	return response, nil
}

// processTaskClaim обрабатывает claim одного задания атомарно
func (s *TaskService) processTaskClaim(ctx context.Context, task *models.ProductionTask) error {
	// 1. Проверяем, есть ли выходные предметы для добавления
	if len(task.OutputItems) > 0 {
		// Готовим данные для AddItems
		itemsToAdd := make([]models.AddItem, len(task.OutputItems))
		for i, outputItem := range task.OutputItems {
			itemsToAdd[i] = models.AddItem{
				ItemID:           outputItem.ItemID,
				Quantity:         outputItem.Quantity,
				CollectionID:     outputItem.CollectionID,
				QualityLevelID:   outputItem.QualityLevelID,
				CollectionCode:   outputItem.CollectionCode,
				QualityLevelCode: outputItem.QualityLevelCode,
			}
		}

		// 2. Пытаемся добавить предметы игроку
		if err := s.inventoryClient.AddItems(ctx, task.UserID, "main", "craft_result", task.ID, itemsToAdd); err != nil {
			// Неудача – не трогаем резерв, оставляем задание в completed, чтобы пользователь попробовал ещё раз
			return fmt.Errorf("failed to add items to inventory: %w", err)
		}
	} else {
		s.logger.Info("Task has no output items, skipping inventory addition",
			zap.String("taskID", task.ID.String()),
			zap.String("userID", task.UserID.String()))
	}

	// 3. Проверяем, был ли создан резерв для этой задачи
	// Получаем рецепт для проверки входных предметов
	recipe, err := s.recipeRepo.GetRecipeByID(ctx, task.RecipeID)
	if err != nil {
		s.logger.Error("Failed to get recipe for task", zap.Error(err), zap.String("taskID", task.ID.String()))
		s.logger.Error("DEBUG: Recipe loading failed, continuing without ConsumeReserve check", zap.String("taskID", task.ID.String()))
		// Продолжаем без проверки резерва - лучше пропустить ConsumeReserve чем упасть
	} else {
		// ERROR: Логируем информацию о рецепте (временно ERROR чтобы точно видеть)
		s.logger.Error("DEBUG: Recipe loaded for claim",
			zap.String("taskID", task.ID.String()),
			zap.String("recipeID", recipe.ID.String()),
			zap.String("recipeName", recipe.Name),
			zap.Int("inputItemsCount", len(recipe.InputItems)))

		// Проверяем есть ли входные предметы у рецепта
		if len(recipe.InputItems) > 0 {
			s.logger.Info("Recipe has input items, attempting ConsumeReserve",
				zap.String("taskID", task.ID.String()),
				zap.Int("inputItemsCount", len(recipe.InputItems)))

			// Только если есть входные предметы - пытаемся уничтожить резерв
			if err := s.inventoryClient.ConsumeReserve(ctx, task.UserID, task.ID); err != nil {
				// ConsumeReserve возвращает ошибку только в случае реальных проблем
				// Ситуация "резервирование не найдено" уже обрабатывается в inventory_client.go
				s.logger.Error("Failed to consume reserve after AddItems", zap.Error(err), zap.String("taskID", task.ID.String()))
				if len(task.OutputItems) > 0 {
					// Пытаемся откатить AddItems через ReturnReserve (best-effort)
					if returnErr := s.inventoryClient.ReturnReserve(ctx, task.UserID, task.ID); returnErr != nil {
						s.logger.Error("ReturnReserve also failed", zap.Error(returnErr), zap.String("taskID", task.ID.String()))
					}
				}
				return fmt.Errorf("failed to consume reserved items: %w", err)
			} else {
				// Успешно потребили резерв или он не был найден (оба случая нормальные)
				s.logger.Info("Successfully consumed reserve or no reservation was needed",
					zap.String("taskID", task.ID.String()),
					zap.String("userID", task.UserID.String()))
			}
		} else {
			s.logger.Info("Recipe has no input items, skipping ConsumeReserve", zap.String("taskID", task.ID.String()))
		}
	}

	// 4. Обновляем статус задания на claimed
	err = s.taskRepo.UpdateTaskStatus(ctx, task.ID, models.TaskStatusClaimed)
	if err != nil {
		s.logger.Error("Failed to update task status to claimed",
			zap.Error(err),
			zap.String("taskID", task.ID.String()))
		// Не возвращаем ошибку, так как предметы уже добавлены в инвентарь
	}

	// 5. Пытаемся запустить следующее задание в очереди
	// Get the recipe to determine operation class (используем уже полученный recipe)
	if recipe != nil {
		s.tryStartNextTask(ctx, task.UserID, recipe.OperationClassCode)
	} else {
		// Fallback если не удалось получить рецепт выше
		recipe, err := s.recipeRepo.GetRecipeByID(ctx, task.RecipeID)
		if err != nil {
			s.logger.Error("Failed to get recipe for task", zap.Error(err), zap.String("taskID", task.ID.String()))
		} else {
			s.tryStartNextTask(ctx, task.UserID, recipe.OperationClassCode)
		}
	}

	return nil
}

// CancelTask отменяет задание в статусе pending с возвратом материалов
func (s *TaskService) CancelTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error {
	// 1. Получаем задание
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// 2. Проверяем права доступа
	if task.UserID != userID {
		return fmt.Errorf("task does not belong to user")
	}

	// 3. Проверяем статус задания
	if task.Status != models.TaskStatusPending {
		return fmt.Errorf("task cannot be cancelled: status is %s", task.Status)
	}

	// 4. Возвращаем зарезервированные предметы
	err = s.inventoryClient.ReturnReserve(ctx, task.UserID, task.ID)
	if err != nil {
		return fmt.Errorf("failed to return reserved items: %w", err)
	}

	// 5. Обновляем статус задания на cancelled
	err = s.taskRepo.UpdateTaskStatus(ctx, task.ID, models.TaskStatusCancelled)
	if err != nil {
		s.logger.Error("Failed to update task status to cancelled",
			zap.Error(err),
			zap.String("taskID", task.ID.String()))
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// 6. Логируем операцию отмены для аудита
	s.logger.Info("Task cancelled successfully",
		zap.String("taskID", task.ID.String()),
		zap.String("userID", userID.String()))

	return nil
}

// tryStartNextTask пытается запустить следующее задание в очереди
func (s *TaskService) tryStartNextTask(ctx context.Context, userID uuid.UUID, operationClass string) {
	// Получаем задания в статусе pending для данного класса операций
	pendingTasks, err := s.taskRepo.GetUserTasks(ctx, userID, []string{models.TaskStatusPending})
	if err != nil {
		s.logger.Error("Failed to get pending tasks", zap.Error(err))
		return
	}

	// Ищем первое подходящее задание
	for _, task := range pendingTasks {
		// Get the recipe to determine operation class
		recipe, err := s.recipeRepo.GetRecipeByID(ctx, task.RecipeID)
		if err != nil {
			s.logger.Error("Failed to get recipe for task", zap.Error(err), zap.String("taskID", task.ID.String()))
			continue
		}
		if recipe.OperationClassCode == operationClass {
			// Проверяем, можно ли его запустить
			if s.canStartImmediately(ctx, userID, operationClass) {
				err = s.taskRepo.UpdateTaskStatus(ctx, task.ID, models.TaskStatusInProgress)
				if err != nil {
					s.logger.Error("Failed to start next task",
						zap.Error(err),
						zap.String("taskID", task.ID.String()))
					return
				}

				s.logger.Info("Started next task in queue",
					zap.String("taskID", task.ID.String()),
					zap.String("operationClass", operationClass))
				return
			}
		}
	}
}

// CalculateSlotInfo рассчитывает информацию о слотах на основе текущих заданий
func (s *TaskService) CalculateSlotInfo(tasks []models.ProductionTask) models.SlotInfo {
	// TODO: Получить реальную информацию о слотах от User Service
	total := 2 // Временно захардкожено
	used := 0

	for _, task := range tasks {
		if task.Status == models.TaskStatusInProgress {
			used++
		}
	}

	return models.SlotInfo{
		Total: total,
		Used:  used,
		Free:  total - used,
	}
}

// CalculateSlotInfoWithUserService рассчитывает информацию о слотах с обращением к User Service
func (s *TaskService) CalculateSlotInfoWithUserService(ctx context.Context, userID uuid.UUID, tasks []models.ProductionTask) models.SlotInfo {
	// Получаем реальную информацию о слотах от User Service
	s.logger.Error("DEBUG: CalculateSlotInfoWithUserService called", zap.String("userID", userID.String()))

	userSlots, err := s.userClient.GetUserProductionSlots(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to get user slots from User Service, using fallback",
			zap.Error(err), zap.String("userID", userID.String()))
		s.logger.Error("DEBUG: Using fallback CalculateSlotInfo", zap.String("userID", userID.String()))
		// Fallback к старой функции если User Service недоступен
		return s.CalculateSlotInfo(tasks)
	}

	s.logger.Error("DEBUG: User Service responded successfully",
		zap.String("userID", userID.String()),
		zap.Int("total_slots", userSlots.TotalSlots))

	// Рассчитываем общее количество слотов
	total := userSlots.TotalSlots
	used := 0

	// Подсчитываем занятые слоты (задачи в процессе выполнения)
	for _, task := range tasks {
		if task.Status == models.TaskStatusInProgress {
			used++
		}
	}

	result := models.SlotInfo{
		Total: total,
		Used:  used,
		Free:  total - used,
	}

	s.logger.Error("DEBUG: CalculateSlotInfoWithUserService result",
		zap.String("userID", userID.String()),
		zap.Int("total", result.Total),
		zap.Int("used", result.Used),
		zap.Int("free", result.Free))

	return result
}

// tryStartPendingTasks пытается автоматически запустить pending задачи если есть свободные слоты
func (s *TaskService) tryStartPendingTasks(ctx context.Context, userID uuid.UUID, currentTasks []models.ProductionTask) error {
	s.logger.Error("DEBUG: tryStartPendingTasks called", zap.String("userID", userID.String()))

	// Получаем слоты пользователя
	userSlots, err := s.userClient.GetUserProductionSlots(ctx, userID)
	if err != nil {
		s.logger.Error("DEBUG: Failed to get user slots", zap.Error(err))
		return fmt.Errorf("failed to get user slots: %w", err)
	}

	s.logger.Error("DEBUG: Got user slots", zap.Int("total_slots", userSlots.TotalSlots), zap.Int("slots_count", len(userSlots.Slots)))

	// Подсчитываем занятые слоты по типам операций
	occupiedSlotsByOperation := make(map[string]int)
	pendingTasksByOperation := make(map[string][]models.ProductionTask)

	for _, task := range currentTasks {
		// Получаем рецепт для определения типа операции
		recipe, err := s.recipeRepo.GetRecipeByID(ctx, task.RecipeID)
		if err != nil {
			s.logger.Error("Failed to get recipe for task", zap.Error(err), zap.String("taskID", task.ID.String()))
			continue
		}

		operationClass := recipe.OperationClassCode

		if task.Status == models.TaskStatusInProgress {
			occupiedSlotsByOperation[operationClass]++
		} else if task.Status == models.TaskStatusPending {
			pendingTasksByOperation[operationClass] = append(pendingTasksByOperation[operationClass], task)
		}
	}

	// Для каждого типа операций проверяем, можно ли запустить pending задачи
	for operationClass, pendingTasks := range pendingTasksByOperation {
		if len(pendingTasks) == 0 {
			continue
		}

		// Подсчитываем доступные слоты для данного типа операций
		totalSlots := 0
		for _, slot := range userSlots.Slots {
			if s.slotSupportsOperation(slot, operationClass) {
				totalSlots += slot.Count
			}
		}

		occupiedSlots := occupiedSlotsByOperation[operationClass]
		freeSlots := totalSlots - occupiedSlots

		s.logger.Info("Checking slots for operation",
			zap.String("operation_class", operationClass),
			zap.Int("total_slots", totalSlots),
			zap.Int("occupied_slots", occupiedSlots),
			zap.Int("free_slots", freeSlots),
			zap.Int("pending_tasks", len(pendingTasks)))

		// Запускаем pending задачи пока есть свободные слоты
		for i := 0; i < len(pendingTasks) && i < freeSlots; i++ {
			taskPtr := &pendingTasks[i]
			err := s.startPendingTask(ctx, taskPtr)
			if err != nil {
				s.logger.Error("Failed to start pending task",
					zap.Error(err),
					zap.String("taskID", taskPtr.ID.String()))
				continue
			}

			s.logger.Info("Auto-started pending task",
				zap.String("taskID", taskPtr.ID.String()),
				zap.String("operation_class", operationClass))
		}
	}

	return nil
}

// startPendingTask запускает одну pending задачу
func (s *TaskService) startPendingTask(ctx context.Context, task *models.ProductionTask) error {
	// Получаем рецепт для расчета времени завершения
	recipe, err := s.recipeRepo.GetRecipeByID(ctx, task.RecipeID)
	if err != nil {
		return fmt.Errorf("failed to get recipe: %w", err)
	}

	// Определяем свободный номер слота перед запуском
	userSlots, err := s.userClient.GetUserProductionSlots(ctx, task.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user slots: %w", err)
	}

	operationClass := recipe.OperationClassCode

	// Получаем уже занятые номера слотов для этого класса операций
	activeTasks, err := s.taskRepo.GetUserTasks(ctx, task.UserID, []string{models.TaskStatusInProgress})
	if err != nil {
		return fmt.Errorf("failed to get active tasks: %w", err)
	}

	usedNumbers := make(map[int]struct{})
	for _, t := range activeTasks {
		rec, err := s.recipeRepo.GetRecipeByID(ctx, t.RecipeID)
		if err != nil {
			continue // пропускаем ошибочный рецепт
		}
		if rec.OperationClassCode == operationClass && t.SlotNumber > 0 {
			usedNumbers[t.SlotNumber] = struct{}{}
		}
	}

	// Рассчитываем максимальное количество слотов, доступных для этого класса
	totalSlots := 0
	for _, slot := range userSlots.Slots {
		if s.slotSupportsOperation(slot, operationClass) {
			totalSlots += slot.Count
		}
	}

	freeSlot := 0
	for i := 1; i <= totalSlots; i++ {
		if _, exists := usedNumbers[i]; !exists {
			freeSlot = i
			break
		}
	}

	if freeSlot == 0 {
		return fmt.Errorf("no free slot found for operation class %s", operationClass)
	}

	// Обновляем номер слота в базе
	if err := s.taskRepo.UpdateTaskSlotNumber(ctx, task.ID, freeSlot); err != nil {
		return err
	}

	// Обновляем в объекте task для дальнейших расчётов
	task.SlotNumber = freeSlot

	now := time.Now()
	task.StartedAt = &now

	// Проверяем мгновенное выполнение
	if recipe.ProductionTimeSeconds == 0 {
		// Мгновенное выполнение - сразу completed
		task.Status = models.TaskStatusCompleted
		task.CompletionTime = &now

		err = s.taskRepo.UpdateTaskStatus(ctx, task.ID, models.TaskStatusCompleted)
		if err != nil {
			return fmt.Errorf("failed to complete instant task: %w", err)
		}

		s.logger.Info("Instantly completed task",
			zap.String("taskID", task.ID.String()),
			zap.String("recipe_id", recipe.ID.String()))
	} else {
		// Обычное производство - in_progress
		task.Status = models.TaskStatusInProgress
		completionTime := now.Add(time.Duration(recipe.ProductionTimeSeconds) * time.Second)
		task.CompletionTime = &completionTime

		err = s.taskRepo.UpdateTaskStatus(ctx, task.ID, models.TaskStatusInProgress)
		if err != nil {
			return fmt.Errorf("failed to start task: %w", err)
		}

		s.logger.Info("Started task",
			zap.String("taskID", task.ID.String()),
			zap.String("recipe_id", recipe.ID.String()),
			zap.Time("completion_time", completionTime))
	}

	return nil
}

// convertToPublicTasks конвертирует ProductionTask в PublicProductionTask (убирает спойлеры)
func (s *TaskService) convertToPublicTasks(tasks []models.ProductionTask) []models.PublicProductionTask {
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
