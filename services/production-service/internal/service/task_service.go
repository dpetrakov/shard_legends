package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/shard-legends/production-service/internal/models"
	"github.com/shard-legends/production-service/internal/storage"
	"go.uber.org/zap"
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
	// Получаем все задания пользователя, кроме claimed и cancelled
	statuses := []string{
		models.TaskStatusPending,
		models.TaskStatusInProgress,
		models.TaskStatusCompleted,
	}

	tasks, err := s.taskRepo.GetUserTasks(ctx, userID, statuses)
	if err != nil {
		s.logger.Error("Failed to get user tasks", zap.Error(err), zap.String("userID", userID.String()))
		return nil, fmt.Errorf("failed to get user tasks: %w", err)
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

	// 9. Создаем задание
	task := &models.ProductionTask{
		ID:                    uuid.New(),
		UserID:                userID,
		RecipeID:              recipe.ID,
		OperationClassCode:    recipe.OperationClassCode,
		Status:                models.TaskStatusPending,
		ProductionTimeSeconds: finalProductionTime,
		CreatedAt:             time.Now(),
		AppliedModifiers:      s.modifierService.BuildAppliedModifiersForAudit(calcCtx.ModificationResults),
		OutputItems:           outputItems,
	}

	// 10. Начинаем транзакционную операцию: создание задания + резервирование предметов
	err = s.createTaskWithReservation(ctx, task, itemsToReserve)
	if err != nil {
		return nil, err
	}

	// 11. Логируем применение модификаторов для аудита
	s.modifierService.LogModifiersApplication(userID, task.ID, calcCtx.ModificationResults)

	// 12. Проверяем, можно ли сразу запустить задание
	if s.canStartImmediately(ctx, userID, recipe.OperationClassCode) {
		task.Status = models.TaskStatusInProgress
		task.StartedAt = &[]time.Time{time.Now()}[0]
		completedAt := time.Now().Add(time.Duration(finalProductionTime) * time.Second)
		task.CompletedAt = &completedAt

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
			ItemID:         input.ItemID,
			Quantity:       input.Quantity * request.ExecutionCount,
			CollectionID:   input.CollectionID,
			QualityLevelID: input.QualityLevelID,
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

// createTaskWithReservation создает задание с атомарным резервированием предметов
func (s *TaskService) createTaskWithReservation(ctx context.Context, task *models.ProductionTask, itemsToReserve []models.ReservationItem) error {
	// Сначала создаем задание
	err := s.taskRepo.CreateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Затем резервируем предметы
	err = s.inventoryClient.ReserveItems(ctx, task.UserID, task.ID, itemsToReserve)
	if err != nil {
		// Откатываем создание задания
		_ = s.taskRepo.UpdateTaskStatus(ctx, task.ID, models.TaskStatusFailed)
		return fmt.Errorf("failed to reserve items: %w", err)
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
		if s.canUseSlot(task.OperationClassCode, operationClass, userSlots) {
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
		if task.OperationClassCode == operationClass {
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

	if len(tasksToProcess) == 0 {
		return &models.ClaimResponse{
			Success:       true,
			ItemsReceived: []models.TaskOutputItem{},
		}, nil
	}

	var allItemsReceived []models.TaskOutputItem

	// Обрабатываем каждое задание
	for _, task := range tasksToProcess {
		err := s.processTaskClaim(ctx, task)
		if err != nil {
			s.logger.Error("Failed to process task claim",
				zap.Error(err),
				zap.String("taskID", task.ID.String()),
				zap.String("userID", userID.String()))
			return nil, fmt.Errorf("failed to process claim for task %s: %w", task.ID, err)
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
	updatedQueue, err := s.GetUserQueue(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get updated queue after claim", zap.Error(err))
		// Не возвращаем ошибку, так как claim прошел успешно
	}

	response := &models.ClaimResponse{
		Success:       true,
		ItemsReceived: allItemsReceived,
	}

	if updatedQueue != nil {
		slotInfo := s.CalculateSlotInfo(updatedQueue)
		response.UpdatedQueueStatus = &models.QueueResponse{
			Tasks:          updatedQueue,
			AvailableSlots: slotInfo,
		}
	}

	return response, nil
}

// processTaskClaim обрабатывает claim одного задания атомарно
func (s *TaskService) processTaskClaim(ctx context.Context, task *models.ProductionTask) error {
	// 1. Потребляем зарезервированные материалы
	err := s.inventoryClient.ConsumeReserve(ctx, task.UserID, task.ID)
	if err != nil {
		return fmt.Errorf("failed to consume reserved items: %w", err)
	}

	// 2. Преобразуем выходные предметы в формат для добавления в инвентарь
	itemsToAdd := make([]models.AddItem, len(task.OutputItems))
	for i, outputItem := range task.OutputItems {
		itemsToAdd[i] = models.AddItem{
			ItemID:         outputItem.ItemID,
			Quantity:       outputItem.Quantity,
			CollectionID:   outputItem.CollectionID,
			QualityLevelID: outputItem.QualityLevelID,
		}
	}

	// 3. Добавляем результаты в инвентарь
	err = s.inventoryClient.AddItems(ctx, task.UserID, "main", "craft_result", task.ID, itemsToAdd)
	if err != nil {
		// При ошибке добавления предметов пытаемся вернуть резерв
		if returnErr := s.inventoryClient.ReturnReserve(ctx, task.UserID, task.ID); returnErr != nil {
			s.logger.Error("Failed to return reserve after add items failure",
				zap.Error(returnErr),
				zap.String("taskID", task.ID.String()))
		}
		return fmt.Errorf("failed to add items to inventory: %w", err)
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
	s.tryStartNextTask(ctx, task.UserID, task.OperationClassCode)

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
		if task.OperationClassCode == operationClass {
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
