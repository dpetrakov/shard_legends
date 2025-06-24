package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shard-legends/inventory-service/internal/models"
)

func TestInventoryRepo_GetDailyBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewInventoryRepository(sqlxDB)

	ctx := context.Background()
	userID := uuid.New()
	sectionID := uuid.New()
	itemID := uuid.New()
	collectionID := uuid.New()
	qualityID := uuid.New()
	date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		expectedBalance := &models.DailyBalance{
			UserID:         userID,
			SectionID:      sectionID,
			ItemID:         itemID,
			CollectionID:   collectionID,
			QualityLevelID: qualityID,
			BalanceDate:    date.Add(24*time.Hour - time.Second), // End of day
			Quantity:       100,
			CreatedAt:      time.Now(),
		}

		rows := sqlmock.NewRows([]string{"user_id", "section_id", "item_id", "collection_id", "quality_level_id", "balance_date", "quantity", "created_at"}).
			AddRow(expectedBalance.UserID, expectedBalance.SectionID, expectedBalance.ItemID,
				expectedBalance.CollectionID, expectedBalance.QualityLevelID, expectedBalance.BalanceDate,
				expectedBalance.Quantity, expectedBalance.CreatedAt)

		mock.ExpectQuery("SELECT user_id, section_id, item_id, collection_id, quality_level_id, balance_date, quantity, created_at FROM daily_balance WHERE user_id = \\$1 AND section_id = \\$2 AND item_id = \\$3 AND collection_id = \\$4 AND quality_level_id = \\$5 AND balance_date = \\$6").
			WithArgs(userID, sectionID, itemID, collectionID, qualityID, sqlmock.AnyArg()).
			WillReturnRows(rows)

		result, err := repo.GetDailyBalance(ctx, userID, sectionID, itemID, collectionID, qualityID, date)
		assert.NoError(t, err)
		assert.Equal(t, expectedBalance.UserID, result.UserID)
		assert.Equal(t, expectedBalance.Quantity, result.Quantity)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT user_id, section_id, item_id, collection_id, quality_level_id, balance_date, quantity, created_at FROM daily_balance WHERE user_id = \\$1 AND section_id = \\$2 AND item_id = \\$3 AND collection_id = \\$4 AND quality_level_id = \\$5 AND balance_date = \\$6").
			WithArgs(userID, sectionID, itemID, collectionID, qualityID, sqlmock.AnyArg()).
			WillReturnError(sql.ErrNoRows)

		result, err := repo.GetDailyBalance(ctx, userID, sectionID, itemID, collectionID, qualityID, date)
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
		assert.Nil(t, result)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestInventoryRepo_GetLatestDailyBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewInventoryRepository(sqlxDB)

	ctx := context.Background()
	userID := uuid.New()
	sectionID := uuid.New()
	itemID := uuid.New()
	collectionID := uuid.New()
	qualityID := uuid.New()
	beforeDate := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		balanceDate := time.Date(2024, 1, 3, 23, 59, 59, 0, time.UTC)
		
		rows := sqlmock.NewRows([]string{"user_id", "section_id", "item_id", "collection_id", "quality_level_id", "balance_date", "quantity", "created_at"}).
			AddRow(userID, sectionID, itemID, collectionID, qualityID, balanceDate, 50, time.Now())

		mock.ExpectQuery("SELECT user_id, section_id, item_id, collection_id, quality_level_id, balance_date, quantity, created_at FROM daily_balance WHERE user_id = \\$1 AND section_id = \\$2 AND item_id = \\$3 AND collection_id = \\$4 AND quality_level_id = \\$5 AND balance_date <= \\$6 ORDER BY balance_date DESC LIMIT 1").
			WithArgs(userID, sectionID, itemID, collectionID, qualityID, sqlmock.AnyArg()).
			WillReturnRows(rows)

		result, err := repo.GetLatestDailyBalance(ctx, userID, sectionID, itemID, collectionID, qualityID, beforeDate)
		assert.NoError(t, err)
		assert.Equal(t, int64(50), result.Quantity)
		assert.Equal(t, balanceDate, result.BalanceDate)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestInventoryRepo_CreateDailyBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewInventoryRepository(sqlxDB)

	ctx := context.Background()
	balance := &models.DailyBalance{
		UserID:         uuid.New(),
		SectionID:      uuid.New(),
		ItemID:         uuid.New(),
		CollectionID:   uuid.New(),
		QualityLevelID: uuid.New(),
		BalanceDate:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		Quantity:       100,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO daily_balance").
			WithArgs(balance.UserID, balance.SectionID, balance.ItemID, balance.CollectionID,
				balance.QualityLevelID, sqlmock.AnyArg(), balance.Quantity, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateDailyBalance(ctx, balance)
		assert.NoError(t, err)
		// Check that timestamps were set
		assert.False(t, balance.CreatedAt.IsZero())
		// Check that balance date was normalized to end of day
		expected := time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC)
		assert.Equal(t, expected, balance.BalanceDate)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestInventoryRepo_GetOperations(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewInventoryRepository(sqlxDB)

	ctx := context.Background()
	userID := uuid.New()
	sectionID := uuid.New()
	itemID := uuid.New()
	collectionID := uuid.New()
	qualityID := uuid.New()
	fromDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		op1ID := uuid.New()
		op2ID := uuid.New()
		opTypeID := uuid.New()

		rows := sqlmock.NewRows([]string{"id", "user_id", "section_id", "item_id", "collection_id", "quality_level_id", "quantity_change", "operation_type_id", "operation_id", "recipe_id", "comment", "created_at"}).
			AddRow(op1ID, userID, sectionID, itemID, collectionID, qualityID, 10, opTypeID, nil, nil, nil, time.Now()).
			AddRow(op2ID, userID, sectionID, itemID, collectionID, qualityID, -5, opTypeID, nil, nil, nil, time.Now())

		mock.ExpectQuery("SELECT id, user_id, section_id, item_id, collection_id, quality_level_id, quantity_change, operation_type_id, operation_id, recipe_id, comment, created_at FROM operation WHERE user_id = \\$1 AND section_id = \\$2 AND item_id = \\$3 AND collection_id = \\$4 AND quality_level_id = \\$5 AND created_at >= \\$6 ORDER BY created_at").
			WithArgs(userID, sectionID, itemID, collectionID, qualityID, fromDate).
			WillReturnRows(rows)

		result, err := repo.GetOperations(ctx, userID, sectionID, itemID, collectionID, qualityID, fromDate)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(10), result[0].QuantityChange)
		assert.Equal(t, int64(-5), result[1].QuantityChange)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestInventoryRepo_CreateOperation(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewInventoryRepository(sqlxDB)

	ctx := context.Background()
	operation := &models.Operation{
		UserID:          uuid.New(),
		SectionID:       uuid.New(),
		ItemID:          uuid.New(),
		CollectionID:    uuid.New(),
		QualityLevelID:  uuid.New(),
		QuantityChange:  10,
		OperationTypeID: uuid.New(),
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO operation").
			WithArgs(sqlmock.AnyArg(), operation.UserID, operation.SectionID, operation.ItemID,
				operation.CollectionID, operation.QualityLevelID, operation.QuantityChange,
				operation.OperationTypeID, operation.OperationID, operation.RecipeID,
				operation.Comment, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateOperation(ctx, operation)
		assert.NoError(t, err)
		// Check that ID and timestamp were set
		assert.NotEqual(t, uuid.Nil, operation.ID)
		assert.False(t, operation.CreatedAt.IsZero())

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestInventoryRepo_CreateOperations(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewInventoryRepository(sqlxDB)

	ctx := context.Background()
	operations := []*models.Operation{
		{
			UserID:          uuid.New(),
			SectionID:       uuid.New(),
			ItemID:          uuid.New(),
			CollectionID:    uuid.New(),
			QualityLevelID:  uuid.New(),
			QuantityChange:  10,
			OperationTypeID: uuid.New(),
		},
		{
			UserID:          uuid.New(),
			SectionID:       uuid.New(),
			ItemID:          uuid.New(),
			CollectionID:    uuid.New(),
			QualityLevelID:  uuid.New(),
			QuantityChange:  -5,
			OperationTypeID: uuid.New(),
		},
	}

	t.Run("success", func(t *testing.T) {
		// Begin transaction
		mock.ExpectBegin()
		
		// Batch insert
		mock.ExpectExec("INSERT INTO operation").
			WillReturnResult(sqlmock.NewResult(2, 2))
		
		// Commit transaction
		mock.ExpectCommit()

		err := repo.CreateOperations(ctx, operations)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty operations", func(t *testing.T) {
		err := repo.CreateOperations(ctx, []*models.Operation{})
		assert.NoError(t, err)
	})
}

func TestInventoryRepo_GetOperationsByExternalID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewInventoryRepository(sqlxDB)

	ctx := context.Background()
	externalID := uuid.New()

	t.Run("success", func(t *testing.T) {
		op1ID := uuid.New()
		op2ID := uuid.New()

		rows := sqlmock.NewRows([]string{"id", "user_id", "section_id", "item_id", "collection_id", "quality_level_id", "quantity_change", "operation_type_id", "operation_id", "recipe_id", "comment", "created_at"}).
			AddRow(op1ID, uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), 10, uuid.New(), externalID, nil, nil, time.Now()).
			AddRow(op2ID, uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), -5, uuid.New(), externalID, nil, nil, time.Now())

		mock.ExpectQuery("SELECT id, user_id, section_id, item_id, collection_id, quality_level_id, quantity_change, operation_type_id, operation_id, recipe_id, comment, created_at FROM operation WHERE operation_id = \\$1 ORDER BY created_at").
			WithArgs(externalID).
			WillReturnRows(rows)

		result, err := repo.GetOperationsByExternalID(ctx, externalID)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, &externalID, result[0].OperationID)
		assert.Equal(t, &externalID, result[1].OperationID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestInventoryRepo_GetUserInventoryItems(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewInventoryRepository(sqlxDB)

	ctx := context.Background()
	userID := uuid.New()
	sectionID := uuid.New()

	t.Run("success", func(t *testing.T) {
		item1ID := uuid.New()
		item2ID := uuid.New()
		collectionID := uuid.New()
		qualityID := uuid.New()

		rows := sqlmock.NewRows([]string{"user_id", "section_id", "item_id", "collection_id", "quality_level_id"}).
			AddRow(userID, sectionID, item1ID, collectionID, qualityID).
			AddRow(userID, sectionID, item2ID, collectionID, qualityID)

		mock.ExpectQuery("SELECT DISTINCT user_id, section_id, item_id, collection_id, quality_level_id FROM").
			WithArgs(userID, sectionID, userID, sectionID).
			WillReturnRows(rows)

		result, err := repo.GetUserInventoryItems(ctx, userID, sectionID)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		if len(result) >= 2 {
			assert.Equal(t, item1ID, result[0].ItemID)
			assert.Equal(t, item2ID, result[1].ItemID)
		}

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestInventoryRepo_TransactionMethods(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewInventoryRepository(sqlxDB)

	ctx := context.Background()

	t.Run("transaction lifecycle", func(t *testing.T) {
		// Begin transaction
		mock.ExpectBegin()
		tx, err := repo.BeginTransaction(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, tx)

		// Commit transaction
		mock.ExpectCommit()
		err = repo.CommitTransaction(tx)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback transaction", func(t *testing.T) {
		// Begin transaction
		mock.ExpectBegin()
		tx, err := repo.BeginTransaction(ctx)
		assert.NoError(t, err)

		// Rollback transaction
		mock.ExpectRollback()
		err = repo.RollbackTransaction(tx)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}