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

func TestItemRepo_GetItemByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewItemRepository(sqlxDB)

	ctx := context.Background()
	itemID := uuid.New()
	classID := uuid.New()
	typeID := uuid.New()
	qualityID := uuid.New()
	collectionID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "item_class_id", "item_type_id", "quality_levels_classifier_id", "collections_classifier_id", "created_at", "updated_at"}).
			AddRow(itemID, classID, typeID, qualityID, collectionID, time.Now(), time.Now())

		mock.ExpectQuery("SELECT id, item_class_id, item_type_id, quality_levels_classifier_id, collections_classifier_id, created_at, updated_at FROM item WHERE id = \\$1").
			WithArgs(itemID).
			WillReturnRows(rows)

		result, err := repo.GetItemByID(ctx, itemID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, itemID, result.ID)
		assert.Equal(t, classID, result.ItemClassID)
		assert.Equal(t, typeID, result.ItemTypeID)
		assert.IsType(t, &models.Item{}, result)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, item_class_id, item_type_id, quality_levels_classifier_id, collections_classifier_id, created_at, updated_at FROM item WHERE id = \\$1").
			WithArgs(itemID).
			WillReturnError(sql.ErrNoRows)

		result, err := repo.GetItemByID(ctx, itemID)
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
		assert.Nil(t, result)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestItemRepo_GetItemsByClass(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewItemRepository(sqlxDB)

	ctx := context.Background()
	classCode := "tools"

	t.Run("success", func(t *testing.T) {
		item1ID := uuid.New()
		item2ID := uuid.New()

		rows := sqlmock.NewRows([]string{"id", "item_class_id", "item_type_id", "quality_levels_classifier_id", "collections_classifier_id", "created_at", "updated_at"}).
			AddRow(item1ID, uuid.New(), uuid.New(), uuid.New(), uuid.New(), time.Now(), time.Now()).
			AddRow(item2ID, uuid.New(), uuid.New(), uuid.New(), uuid.New(), time.Now(), time.Now())

		mock.ExpectQuery("SELECT i.id, i.item_class_id, i.item_type_id, i.quality_levels_classifier_id, i.collections_classifier_id, i.created_at, i.updated_at FROM item i JOIN classifier_item ci ON ci.id = i.item_class_id WHERE ci.code = \\$1 AND ci.is_active = true ORDER BY i.created_at").
			WithArgs(classCode).
			WillReturnRows(rows)

		result, err := repo.GetItemsByClass(ctx, classCode)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, item1ID, result[0].ID)
		assert.Equal(t, item2ID, result[1].ID)
		assert.IsType(t, []*models.Item{}, result)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "item_class_id", "item_type_id", "quality_levels_classifier_id", "collections_classifier_id", "created_at", "updated_at"})

		mock.ExpectQuery("SELECT i.id, i.item_class_id, i.item_type_id, i.quality_levels_classifier_id, i.collections_classifier_id, i.created_at, i.updated_at FROM item i JOIN classifier_item ci ON ci.id = i.item_class_id WHERE ci.code = \\$1 AND ci.is_active = true ORDER BY i.created_at").
			WithArgs(classCode).
			WillReturnRows(rows)

		result, err := repo.GetItemsByClass(ctx, classCode)
		assert.NoError(t, err)
		assert.Empty(t, result)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestItemRepo_GetItemsByClassAndType(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewItemRepository(sqlxDB)

	ctx := context.Background()
	classCode := "tools"
	typeCode := "axe"

	t.Run("success", func(t *testing.T) {
		itemID := uuid.New()

		rows := sqlmock.NewRows([]string{"id", "item_class_id", "item_type_id", "quality_levels_classifier_id", "collections_classifier_id", "created_at", "updated_at"}).
			AddRow(itemID, uuid.New(), uuid.New(), uuid.New(), uuid.New(), time.Now(), time.Now())

		mock.ExpectQuery("SELECT i.id, i.item_class_id, i.item_type_id, i.quality_levels_classifier_id, i.collections_classifier_id, i.created_at, i.updated_at FROM item i JOIN classifier_item ci_class ON ci_class.id = i.item_class_id JOIN classifier_item ci_type ON ci_type.id = i.item_type_id WHERE ci_class.code = \\$1 AND ci_type.code = \\$2 AND ci_class.is_active = true AND ci_type.is_active = true ORDER BY i.created_at").
			WithArgs(classCode, typeCode).
			WillReturnRows(rows)

		result, err := repo.GetItemsByClassAndType(ctx, classCode, typeCode)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, itemID, result[0].ID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestItemRepo_GetItemImage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewItemRepository(sqlxDB)

	ctx := context.Background()
	itemID := uuid.New()
	collectionID := uuid.New()
	qualityID := uuid.New()

	t.Run("success", func(t *testing.T) {
		imageURL := "https://example.com/image.png"

		rows := sqlmock.NewRows([]string{"item_id", "collection_id", "quality_level_id", "image_url", "is_active", "created_at", "updated_at"}).
			AddRow(itemID, collectionID, qualityID, imageURL, true, time.Now(), time.Now())

		mock.ExpectQuery("SELECT item_id, collection_id, quality_level_id, image_url, is_active, created_at, updated_at FROM item_image WHERE item_id = \\$1 AND collection_id = \\$2 AND quality_level_id = \\$3 AND is_active = true").
			WithArgs(itemID, collectionID, qualityID).
			WillReturnRows(rows)

		result, err := repo.GetItemImage(ctx, itemID, collectionID, qualityID)
		assert.NoError(t, err)
		assert.Equal(t, itemID, result.ItemID)
		assert.Equal(t, collectionID, result.CollectionID)
		assert.Equal(t, qualityID, result.QualityLevelID)
		assert.Equal(t, imageURL, result.ImageURL)
		assert.IsType(t, &models.ItemImage{}, result)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT item_id, collection_id, quality_level_id, image_url, is_active, created_at, updated_at FROM item_image WHERE item_id = \\$1 AND collection_id = \\$2 AND quality_level_id = \\$3 AND is_active = true").
			WithArgs(itemID, collectionID, qualityID).
			WillReturnError(sql.ErrNoRows)

		result, err := repo.GetItemImage(ctx, itemID, collectionID, qualityID)
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
		assert.Nil(t, result)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestItemRepo_GetItemImages(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewItemRepository(sqlxDB)

	ctx := context.Background()
	itemID := uuid.New()

	t.Run("success", func(t *testing.T) {
		collection1ID := uuid.New()
		collection2ID := uuid.New()
		qualityID := uuid.New()

		rows := sqlmock.NewRows([]string{"item_id", "collection_id", "quality_level_id", "image_url", "is_active", "created_at", "updated_at"}).
			AddRow(itemID, collection1ID, qualityID, "https://example.com/image1.png", true, time.Now(), time.Now()).
			AddRow(itemID, collection2ID, qualityID, "https://example.com/image2.png", true, time.Now(), time.Now())

		mock.ExpectQuery("SELECT item_id, collection_id, quality_level_id, image_url, is_active, created_at, updated_at FROM item_image WHERE item_id = \\$1 AND is_active = true ORDER BY collection_id, quality_level_id").
			WithArgs(itemID).
			WillReturnRows(rows)

		result, err := repo.GetItemImages(ctx, itemID)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, collection1ID, result[0].CollectionID)
		assert.Equal(t, collection2ID, result[1].CollectionID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"item_id", "collection_id", "quality_level_id", "image_url", "is_active", "created_at", "updated_at"})

		mock.ExpectQuery("SELECT item_id, collection_id, quality_level_id, image_url, is_active, created_at, updated_at FROM item_image WHERE item_id = \\$1 AND is_active = true ORDER BY collection_id, quality_level_id").
			WithArgs(itemID).
			WillReturnRows(rows)

		result, err := repo.GetItemImages(ctx, itemID)
		assert.NoError(t, err)
		assert.Empty(t, result)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestItemRepo_GetItemWithDetails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewItemRepository(sqlxDB)

	ctx := context.Background()
	itemID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "item_class_id", "item_type_id", "quality_levels_classifier_id", "collections_classifier_id", "created_at", "updated_at", "item_class", "item_type"}).
			AddRow(itemID, uuid.New(), uuid.New(), uuid.New(), uuid.New(), time.Now(), time.Now(), "tools", "axe")

		mock.ExpectQuery("SELECT i.id, i.item_class_id, i.item_type_id, i.quality_levels_classifier_id, i.collections_classifier_id, i.created_at, i.updated_at, ci_class.code as item_class, ci_type.code as item_type FROM item i JOIN classifier_item ci_class ON ci_class.id = i.item_class_id JOIN classifier_item ci_type ON ci_type.id = i.item_type_id WHERE i.id = \\$1").
			WithArgs(itemID).
			WillReturnRows(rows)

		result, err := repo.GetItemWithDetails(ctx, itemID)
		assert.NoError(t, err)
		assert.Equal(t, itemID, result.ID)
		assert.Equal(t, "tools", result.ItemClass)
		assert.Equal(t, "axe", result.ItemType)
		assert.IsType(t, &models.ItemWithDetails{}, result)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT i.id, i.item_class_id, i.item_type_id, i.quality_levels_classifier_id, i.collections_classifier_id, i.created_at, i.updated_at, ci_class.code as item_class, ci_type.code as item_type FROM item i JOIN classifier_item ci_class ON ci_class.id = i.item_class_id JOIN classifier_item ci_type ON ci_type.id = i.item_type_id WHERE i.id = \\$1").
			WithArgs(itemID).
			WillReturnError(sql.ErrNoRows)

		result, err := repo.GetItemWithDetails(ctx, itemID)
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
		assert.Nil(t, result)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}