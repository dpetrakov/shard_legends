package service

import (
	"fmt"

	"github.com/google/uuid"
)

// Chest to recipe mapping based on the migration file
var chestRecipeMapping = map[string]map[string]uuid.UUID{
	"resource_chest": {
		"small":  uuid.MustParse("7d0afba0-985e-4d74-b027-3b2a32bb2760"), // resource_chest_s_open
		"medium": uuid.MustParse("4a5026a1-b851-48e9-88d1-156d2e3aa8b9"), // resource_chest_m_open
		"large":  uuid.MustParse("9f4c3d36-b4e1-4a61-b4f1-0ed2a8b16c77"), // resource_chest_l_open
	},
	// TODO: Add other chest types when their recipes are implemented
	// "reagent_chest": {
	//     "small": uuid.MustParse("..."),
	//     ...
	// },
}

// Chest item IDs based on the items migration
var chestItemMapping = map[string]map[string]uuid.UUID{
	"resource_chest": {
		"small":  uuid.MustParse("9421cc9f-a56e-4c7d-b636-4c8fdfef7166"), // resource_chest_s
		"medium": uuid.MustParse("6c0f7fd6-4a6e-4d42-b596-a1a2b775cdbc"), // resource_chest_m
		"large":  uuid.MustParse("0f8aa2c1-25b8-4aed-9d6b-8c1e927bf71f"), // resource_chest_l
	},
	"reagent_chest": {
		"small":  uuid.MustParse("a2e20668-380d-43eb-87db-cb19e4fed0ab"), // reagent_chest_s
		"medium": uuid.MustParse("b6dde60a-6530-4fa3-836b-415520d05f37"), // reagent_chest_m
		"large":  uuid.MustParse("359e86d5-d094-4b2b-b96e-6114e3c66d6b"), // reagent_chest_l
	},
	"booster_chest": {
		"small":  uuid.MustParse("3b5c8322-c00d-44e2-875e-d5bd9097d1c4"), // booster_chest_s
		"medium": uuid.MustParse("d9a3e79a-50d3-4ab5-be86-8137145c34e3"), // booster_chest_m
		"large":  uuid.MustParse("aa58eb38-5e91-47f0-bd4e-6ed02cb059b1"), // booster_chest_l
	},
	"blueprint_chest": {
		"small": uuid.MustParse("012d9076-a37d-4e9d-a49a-fbc7a07e5bd9"), // blueprint_chest (no quality variants)
	},
}

// GetRecipeIDForChest returns the recipe ID for opening a specific chest type and quality
func GetRecipeIDForChest(chestType, qualityLevel string) (uuid.UUID, error) {
	chestTypes, exists := chestRecipeMapping[chestType]
	if !exists {
		return uuid.Nil, fmt.Errorf("unsupported chest type: %s", chestType)
	}

	recipeID, exists := chestTypes[qualityLevel]
	if !exists {
		return uuid.Nil, fmt.Errorf("unsupported quality level %s for chest type %s", qualityLevel, chestType)
	}

	return recipeID, nil
}

// GetItemIDForChest returns the item ID for a specific chest type and quality
func GetItemIDForChest(chestType, qualityLevel string) (uuid.UUID, error) {
	chestTypes, exists := chestItemMapping[chestType]
	if !exists {
		return uuid.Nil, fmt.Errorf("unknown chest type: %s", chestType)
	}

	// Handle blueprint_chest which doesn't have quality variants
	if chestType == "blueprint_chest" {
		itemID, exists := chestTypes["small"] // Use "small" as default key for blueprint
		if !exists {
			return uuid.Nil, fmt.Errorf("blueprint chest item ID not found")
		}
		return itemID, nil
	}

	itemID, exists := chestTypes[qualityLevel]
	if !exists {
		return uuid.Nil, fmt.Errorf("unknown quality level %s for chest type %s", qualityLevel, chestType)
	}

	return itemID, nil
}
