package models

// ProductionSlot представляет производственный слот
type ProductionSlot struct {
	SlotType            string   `json:"slot_type"`
	SupportedOperations []string `json:"supported_operations"`
	Count               int      `json:"count"`
}

// ProductionSlotsResponse - ответ эндпоинта GET /production-slots
type ProductionSlotsResponse struct {
	UserID     string           `json:"user_id"`
	TotalSlots int              `json:"total_slots"`
	Slots      []ProductionSlot `json:"slots"`
}

// VIPModifier представляет VIP модификаторы
type VIPModifier struct {
	Level               string  `json:"level"`
	ProductionSpeedBonus float64 `json:"production_speed_bonus"`
	QualityBonus        float64 `json:"quality_bonus"`
}

// CharacterLevelModifier представляет модификаторы уровня персонажа
type CharacterLevelModifier struct {
	Level        int     `json:"level"`
	CraftingBonus float64 `json:"crafting_bonus"`
}

// Achievement представляет достижение пользователя
type Achievement struct {
	ID         string  `json:"id"`
	BonusType  string  `json:"bonus_type"`
	BonusValue float64 `json:"bonus_value"`
}

// ClanBonuses представляет клановые бонусы
type ClanBonuses struct {
	ProductionSpeed float64 `json:"production_speed"`
}

// ProductionModifiers представляет все модификаторы пользователя
type ProductionModifiers struct {
	VIPStatus      VIPModifier             `json:"vip_status"`
	CharacterLevel CharacterLevelModifier  `json:"character_level"`
	Achievements   []Achievement           `json:"achievements"`
	ClanBonuses    ClanBonuses             `json:"clan_bonuses"`
}

// ProductionModifiersResponse - ответ внутреннего эндпоинта
type ProductionModifiersResponse struct {
	UserID    string              `json:"user_id"`
	Modifiers ProductionModifiers `json:"modifiers"`
}