package models

import (
	"time"

	"github.com/google/uuid"
)

// DailyBalance represents a snapshot of item balance at the end of a day
type DailyBalance struct {
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	SectionID      uuid.UUID `json:"section_id" db:"section_id"`
	ItemID         uuid.UUID `json:"item_id" db:"item_id"`
	CollectionID   uuid.UUID `json:"collection_id" db:"collection_id"`
	QualityLevelID uuid.UUID `json:"quality_level_id" db:"quality_level_id"`
	BalanceDate    time.Time `json:"balance_date" db:"balance_date"`
	Quantity       int64     `json:"quantity" db:"quantity"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// Operation represents an inventory operation (addition or removal of items)
type Operation struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	SectionID       uuid.UUID  `json:"section_id" db:"section_id"`
	ItemID          uuid.UUID  `json:"item_id" db:"item_id"`
	CollectionID    uuid.UUID  `json:"collection_id" db:"collection_id"`
	QualityLevelID  uuid.UUID  `json:"quality_level_id" db:"quality_level_id"`
	QuantityChange  int64      `json:"quantity_change" db:"quantity_change"`
	OperationTypeID uuid.UUID  `json:"operation_type_id" db:"operation_type_id"`
	OperationID     *uuid.UUID `json:"operation_id,omitempty" db:"operation_id"`
	RecipeID        *uuid.UUID `json:"recipe_id,omitempty" db:"recipe_id"`
	Comment         *string    `json:"comment,omitempty" db:"comment"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

// ItemKey represents a unique combination of item parameters
type ItemKey struct {
	UserID         uuid.UUID
	SectionID      uuid.UUID
	ItemID         uuid.UUID
	CollectionID   uuid.UUID
	QualityLevelID uuid.UUID
}

// NewItemKey creates a new ItemKey
func NewItemKey(userID, sectionID, itemID, collectionID, qualityLevelID uuid.UUID) ItemKey {
	return ItemKey{
		UserID:         userID,
		SectionID:      sectionID,
		ItemID:         itemID,
		CollectionID:   collectionID,
		QualityLevelID: qualityLevelID,
	}
}

// String returns a string representation of the ItemKey for caching
func (k ItemKey) String() string {
	return k.UserID.String() + ":" + k.SectionID.String() + ":" +
		k.ItemID.String() + ":" + k.CollectionID.String() + ":" +
		k.QualityLevelID.String()
}

// CacheKey returns the Redis cache key for this item
func (k ItemKey) CacheKey() string {
	return "inventory:" + k.String()
}

// OperationBatch represents a batch of operations to be executed together
type OperationBatch struct {
	Operations []*Operation
	UserID     uuid.UUID
	ExternalID uuid.UUID // External operation ID for linking
}
