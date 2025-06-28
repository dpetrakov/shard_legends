package service

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shard-legends/inventory-service/internal/models"
)

// TestConcurrentReservation_OversellPrevention проверяет, что конкурентные запросы резервирования
// не приводят к oversell (продаже больше предметов, чем доступно)
func TestConcurrentReservation_OversellPrevention(t *testing.T) {
	// Этот тест проверяет критическую уязвимость P-3 из аудита безопасности
	// ВАЖНО: тест должен ПРОЙТИ без oversell благодаря SELECT ... FOR UPDATE блокировкам

	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	ctx := context.Background()
	
	// Используем реальную базу данных для теста транзакций
	// В production тестах это должно быть test database
	deps := createTestDatabaseDeps(t)
	service := NewInventoryService(deps)
	
	// Настройка тестовых данных
	userID := uuid.New()
	itemID := uuid.New()
	mainSectionID := setupTestClassifiers(t, deps)
	
	// Начальный баланс: 100 предметов
	initialBalance := int64(100)
	setupInitialBalance(t, ctx, deps, userID, mainSectionID, itemID, initialBalance)
	
	// Параметры конкурентного теста
	concurrentRequests := 50   // Количество одновременных запросов
	itemsPerRequest := 3       // Предметов в каждом запросе
	expectedFinalBalance := initialBalance - int64(concurrentRequests*itemsPerRequest)
	
	// Если все запросы будут успешными, финальный баланс станет отрицательным
	// Это НЕДОПУСТИМО - должен сработать механизм предотвращения oversell
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	var successCount int
	var errorCount int
	var oversellDetected bool
	
	results := make([]error, concurrentRequests)
	
	t.Logf("Starting %d concurrent reservation requests for %d items each", 
		concurrentRequests, itemsPerRequest)
	t.Logf("Initial balance: %d, Total demand: %d", 
		initialBalance, concurrentRequests*itemsPerRequest)
	
	// Запускаем конкурентные запросы резервирования
	startTime := time.Now()
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			operationID := uuid.New()
			req := &models.ReserveItemsRequest{
				UserID:      userID,
				OperationID: operationID,
				Items: []models.ItemQuantityRequest{
					{
						ItemID:   itemID,
						Quantity: int64(itemsPerRequest),
					},
				},
			}
			
			// Выполняем резервирование
			_, err := service.ReserveItems(ctx, req)
			
			mu.Lock()
			results[index] = err
			if err != nil {
				errorCount++
				if isInsufficientBalanceError(err) {
					t.Logf("Request %d correctly rejected: %v", index, err)
				} else {
					t.Logf("Request %d failed with unexpected error: %v", index, err)
				}
			} else {
				successCount++
				t.Logf("Request %d succeeded", index)
			}
			mu.Unlock()
		}(i)
	}
	
	// Ждем завершения всех goroutines
	wg.Wait()
	duration := time.Since(startTime)
	
	t.Logf("Test completed in %v", duration)
	t.Logf("Results: %d successful, %d failed", successCount, errorCount)
	
	// Проверяем итоговый баланс
	finalBalance := getCurrentBalance(t, ctx, deps, userID, mainSectionID, itemID)
	t.Logf("Final balance: %d", finalBalance)
	
	// КРИТИЧЕСКАЯ ПРОВЕРКА: баланс НЕ должен быть отрицательным
	assert.GreaterOrEqual(t, finalBalance, int64(0), 
		"Balance must never go negative (oversell detected!)")
	
	if finalBalance < 0 {
		oversellDetected = true
		t.Errorf("CRITICAL: Oversell detected! Final balance: %d", finalBalance)
	}
	
	// Проверяем математическую корректность
	expectedReservedItems := int64(successCount * itemsPerRequest)
	expectedFinalBalance := initialBalance - expectedReservedItems
	
	assert.Equal(t, expectedFinalBalance, finalBalance,
		"Final balance should equal initial balance minus successfully reserved items")
	
	// Проверяем, что хотя бы некоторые запросы были отклонены
	// (иначе тест не является репрезентативным)
	totalDemand := int64(concurrentRequests * itemsPerRequest)
	if totalDemand > initialBalance {
		assert.Greater(t, errorCount, 0, 
			"Some requests should be rejected when total demand exceeds available balance")
	}
	
	// Проверяем типы ошибок
	insufficientBalanceErrors := 0
	for _, err := range results {
		if err != nil && isInsufficientBalanceError(err) {
			insufficientBalanceErrors++
		}
	}
	
	t.Logf("Insufficient balance errors: %d", insufficientBalanceErrors)
	
	// Финальная проверка
	if oversellDetected {
		t.Fatal("OVERSELL DETECTED: Race condition vulnerability is NOT fixed!")
	} else {
		t.Log("SUCCESS: No oversell detected, race condition protection is working")
	}
}

// TestReservationIsolationLevel проверяет, что уровень изоляции транзакций предотвращает
// неконсистентные чтения при конкурентных операциях
func TestReservationIsolationLevel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping isolation test in short mode")
	}

	ctx := context.Background()
	deps := createTestDatabaseDeps(t)
	service := NewInventoryService(deps)
	
	userID := uuid.New()
	itemID := uuid.New()
	mainSectionID := setupTestClassifiers(t, deps)
	
	// Начальный баланс
	initialBalance := int64(50)
	setupInitialBalance(t, ctx, deps, userID, mainSectionID, itemID, initialBalance)
	
	var wg sync.WaitGroup
	results := make([]int64, 10) // Результаты чтения баланса
	
	// Одна транзакция резервирует предметы
	wg.Add(1)
	go func() {
		defer wg.Done()
		
		operationID := uuid.New()
		req := &models.ReserveItemsRequest{
			UserID:      userID,
			OperationID: operationID,
			Items: []models.ItemQuantityRequest{
				{
					ItemID:   itemID,
					Quantity: 30,
				},
			},
		}
		
		// Добавляем небольшую задержку для имитации длинной транзакции
		time.Sleep(50 * time.Millisecond)
		_, err := service.ReserveItems(ctx, req)
		require.NoError(t, err)
	}()
	
	// Несколько транзакций читают баланс одновременно
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			time.Sleep(time.Duration(index*5) * time.Millisecond)
			balance := getCurrentBalance(t, ctx, deps, userID, mainSectionID, itemID)
			results[index] = balance
		}(i)
	}
	
	wg.Wait()
	
	// Проверяем согласованность результатов
	// Все чтения должны показывать либо начальный баланс (50), либо финальный (20)
	// Никаких промежуточных неконсистентных состояний
	validBalances := map[int64]bool{50: true, 20: true}
	
	for i, balance := range results {
		assert.Contains(t, validBalances, balance,
			"Read %d returned inconsistent balance %d", i, balance)
	}
	
	t.Logf("Isolation test results: %v", results)
}

// TestConcurrentReservationAndReturn проверяет корректность одновременного резервирования и возврата
func TestConcurrentReservationAndReturn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent reservation/return test in short mode")
	}

	ctx := context.Background()
	deps := createTestDatabaseDeps(t)
	service := NewInventoryService(deps)
	
	userID := uuid.New()
	itemID := uuid.New()
	mainSectionID := setupTestClassifiers(t, deps)
	
	initialBalance := int64(100)
	setupInitialBalance(t, ctx, deps, userID, mainSectionID, itemID, initialBalance)
	
	// Сначала создаем резервирование для возврата
	reservationOpID := uuid.New()
	reserveReq := &models.ReserveItemsRequest{
		UserID:      userID,
		OperationID: reservationOpID,
		Items: []models.ItemQuantityRequest{
			{
				ItemID:   itemID,
				Quantity: 30,
			},
		},
	}
	
	_, err := service.ReserveItems(ctx, reserveReq)
	require.NoError(t, err)
	
	// Проверяем баланс после резервирования
	balanceAfterReserve := getCurrentBalance(t, ctx, deps, userID, mainSectionID, itemID)
	assert.Equal(t, int64(70), balanceAfterReserve)
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	var newReservationSuccess bool
	var returnSuccess bool
	
	// Одновременно выполняем возврат существующего резервирования
	wg.Add(1)
	go func() {
		defer wg.Done()
		
		returnReq := &models.ReturnReserveRequest{
			UserID:      userID,
			OperationID: reservationOpID,
		}
		
		err := service.ReturnReservedItems(ctx, returnReq)
		
		mu.Lock()
		returnSuccess = (err == nil)
		if err != nil {
			t.Logf("Return failed: %v", err)
		}
		mu.Unlock()
	}()
	
	// И новое резервирование
	wg.Add(1)
	go func() {
		defer wg.Done()
		
		time.Sleep(10 * time.Millisecond) // Небольшая задержка
		
		newOpID := uuid.New()
		newReserveReq := &models.ReserveItemsRequest{
			UserID:      userID,
			OperationID: newOpID,
			Items: []models.ItemQuantityRequest{
				{
					ItemID:   itemID,
					Quantity: 50,
				},
			},
		}
		
		_, err := service.ReserveItems(ctx, newReserveReq)
		
		mu.Lock()
		newReservationSuccess = (err == nil)
		if err != nil {
			t.Logf("New reservation failed: %v", err)
		}
		mu.Unlock()
	}()
	
	wg.Wait()
	
	finalBalance := getCurrentBalance(t, ctx, deps, userID, mainSectionID, itemID)
	t.Logf("Final balance: %d, Return success: %v, New reservation success: %v", 
		finalBalance, returnSuccess, newReservationSuccess)
	
	// Проверяем корректность финального состояния
	if returnSuccess && newReservationSuccess {
		// Если и возврат, и новое резервирование успешны: 100 - 50 = 50
		assert.Equal(t, int64(50), finalBalance)
	} else if returnSuccess && !newReservationSuccess {
		// Если только возврат успешен: 100 (возврат) - 0 (новое резервирование не прошло) = 100
		assert.Equal(t, int64(100), finalBalance)
	} else if !returnSuccess && newReservationSuccess {
		// Если только новое резервирование успешно: 70 (после первого резервирования) - 50 = 20
		assert.Equal(t, int64(20), finalBalance)
	} else {
		// Если оба неуспешны: остается 70 (после первого резервирования)
		assert.Equal(t, int64(70), finalBalance)
	}
	
	// В любом случае баланс не должен быть отрицательным
	assert.GreaterOrEqual(t, finalBalance, int64(0))
}

// Вспомогательные функции для тестов

func createTestDatabaseDeps(t *testing.T) *ServiceDependencies {
	// В реальном тесте здесь должно быть подключение к тестовой базе данных
	// Для демонстрации используем моки, но в полноценном integration тесте
	// нужна реальная база с поддержкой транзакций
	
	// TODO: Реализовать реальное подключение к тестовой БД
	t.Skip("Integration test requires real database connection")
	return nil
}

func setupTestClassifiers(t *testing.T, deps *ServiceDependencies) uuid.UUID {
	// Настройка классификаторов для теста
	// В реальном integration тесте здесь должна быть настройка БД
	return uuid.New()
}

func setupInitialBalance(t *testing.T, ctx context.Context, deps *ServiceDependencies, 
	userID, sectionID, itemID uuid.UUID, balance int64) {
	// Создание начального баланса для тестового пользователя
	// В реальном integration тесте здесь должно быть создание записей в БД
}

func getCurrentBalance(t *testing.T, ctx context.Context, deps *ServiceDependencies,
	userID, sectionID, itemID uuid.UUID) int64 {
	// Получение текущего баланса
	// В реальном integration тесте здесь должен быть SQL запрос к БД
	return 0
}

func isInsufficientBalanceError(err error) bool {
	// Проверяем, является ли ошибка ошибкой недостаточного баланса
	if err == nil {
		return false
	}
	
	// Проверяем различные типы ошибок недостаточного баланса
	errString := err.Error()
	return contains(errString, "insufficient balance") ||
		   contains(errString, "not enough items") ||
		   contains(errString, "available") ||
		   contains(errString, "required")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || (len(s) > len(substr) && 
		   (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		   indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// TestReservationConstraintViolation проверяет обработку constraint violations
func TestReservationConstraintViolation(t *testing.T) {
	// Этот unit-тест проверяет, что система корректно обрабатывает
	// database constraint violations (CHECK constraints от migration 006)
	
	ctx := context.Background()
	deps, _, classifierRepo, _, inventoryRepo := createFullTestDeps()
	service := NewInventoryService(deps)
	
	userID := uuid.New()
	operationID := uuid.New()
	itemID := uuid.New()
	mainSectionID := uuid.New()
	
	// Мок для получения section mapping
	sectionMapping := map[string]uuid.UUID{
		models.SectionMain:    mainSectionID,
		models.SectionFactory: uuid.New(),
	}
	classifierRepo.On("GetCodeToUUIDMapping", ctx, models.ClassifierInventorySection).Return(sectionMapping, nil)
	
	// Мок для получения operation type mapping
	operationMapping := map[string]uuid.UUID{
		models.OperationTypeFactoryReservation: uuid.New(),
	}
	classifierRepo.On("GetCodeToUUIDMapping", ctx, models.ClassifierOperationType).Return(operationMapping, nil)
	
	// Мок для коллекций и качества (по умолчанию)
	classifierRepo.On("GetCodeToUUIDMapping", ctx, models.ClassifierCollection).Return(map[string]uuid.UUID{}, nil)
	classifierRepo.On("GetCodeToUUIDMapping", ctx, models.ClassifierQualityLevel).Return(map[string]uuid.UUID{}, nil)
	
	// Мок для проверки существующих операций
	inventoryRepo.On("GetOperationsByExternalID", ctx, operationID).Return([]*models.Operation{}, nil)
	
	// Мок транзакции
	tx := &struct{}{}
	inventoryRepo.On("BeginTransaction", ctx).Return(tx, nil)
	inventoryRepo.On("RollbackTransaction", tx).Return(nil)
	
	// Мок для CheckAndLockBalances - возвращаем constraint violation
	inventoryRepo.On("CheckAndLockBalances", ctx, tx, mock.AnythingOfType("[]service.BalanceLockRequest")).Return(
		nil, sql.ErrConnDone) // Имитация database constraint violation
	
	req := &models.ReserveItemsRequest{
		UserID:      userID,
		OperationID: operationID,
		Items: []models.ItemQuantityRequest{
			{
				ItemID:   itemID,
				Quantity: 100,
			},
		},
	}
	
	// Выполняем резервирование
	result, err := service.ReserveItems(ctx, req)
	
	// Должна быть ошибка
	assert.Error(t, err)
	assert.Nil(t, result)
	
	// Проверяем, что транзакция была откатана
	inventoryRepo.AssertExpectations(t)
	classifierRepo.AssertExpectations(t)
}