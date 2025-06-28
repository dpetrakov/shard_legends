package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestItemKey_String(t *testing.T) {
	userID := uuid.New()
	sectionID := uuid.New()
	itemID := uuid.New()
	collectionID := uuid.New()
	qualityLevelID := uuid.New()

	key := NewItemKey(userID, sectionID, itemID, collectionID, qualityLevelID)

	expected := userID.String() + ":" + sectionID.String() + ":" +
		itemID.String() + ":" + collectionID.String() + ":" +
		qualityLevelID.String()

	assert.Equal(t, expected, key.String())
}

func TestItemKey_CacheKey(t *testing.T) {
	userID := uuid.New()
	sectionID := uuid.New()
	itemID := uuid.New()
	collectionID := uuid.New()
	qualityLevelID := uuid.New()

	key := NewItemKey(userID, sectionID, itemID, collectionID, qualityLevelID)

	expected := "inventory:" + userID.String() + ":" + sectionID.String() + ":" +
		itemID.String() + ":" + collectionID.String() + ":" +
		qualityLevelID.String()

	assert.Equal(t, expected, key.CacheKey())
}

func TestNewItemKey(t *testing.T) {
	userID := uuid.New()
	sectionID := uuid.New()
	itemID := uuid.New()
	collectionID := uuid.New()
	qualityLevelID := uuid.New()

	key := NewItemKey(userID, sectionID, itemID, collectionID, qualityLevelID)

	assert.Equal(t, userID, key.UserID)
	assert.Equal(t, sectionID, key.SectionID)
	assert.Equal(t, itemID, key.ItemID)
	assert.Equal(t, collectionID, key.CollectionID)
	assert.Equal(t, qualityLevelID, key.QualityLevelID)
}
