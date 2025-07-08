package services

import (
	"math/rand"
	"time"
	"user-service/internal/models"
)

// MockDataService предоставляет заглушечные данные для временной версии
type MockDataService struct{}

// NewMockDataService создает новый сервис моковых данных
func NewMockDataService() *MockDataService {
	return &MockDataService{}
}

// GenerateVIPStatus генерирует случайный VIP статус
func (s *MockDataService) GenerateVIPStatus() models.VIPStatus {
	// 30% вероятность активного VIP
	if rand.Float32() < 0.3 {
		// Случайная дата окончания от 1 до 30 дней
		expirationDays := rand.Intn(30) + 1
		expiresAt := time.Now().AddDate(0, 0, expirationDays)
		return models.VIPStatus{
			IsActive:  true,
			ExpiresAt: &expiresAt,
		}
	}
	return models.VIPStatus{
		IsActive:  false,
		ExpiresAt: nil,
	}
}

// GetProfileData возвращает моковые данные профиля пользователя
func (s *MockDataService) GetProfileData(userID string, telegramID int64) models.ProfileResponse {
	return models.ProfileResponse{
		UserID:            userID,
		TelegramID:        telegramID,
		DisplayName:       "John Doe", // Статическое имя
		VIPStatus:         s.GenerateVIPStatus(),
		ProfileImage:      "https://t.me/i/userpic/320/abc123.jpg", // Статическое изображение
		Level:             15,                                      // Статический уровень
		Experience:        2850,                                    // Статический опыт
		AchievementsCount: 7,                                       // Статическое количество достижений
		Clan: &models.Clan{
			ID:   "clan-uuid-12345",
			Name: "Dragon Warriors", // Статический клан
			Role: "member",
		},
	}
}

// GetProductionSlots возвращает информацию о производственных слотах
func (s *MockDataService) GetProductionSlots(userID string) models.ProductionSlotsResponse {
	return models.ProductionSlotsResponse{
		UserID:     userID,
		TotalSlots: 4, // 1 chest_opening + 2 smithy + 1 trade_purchase
		Slots: []models.ProductionSlot{
			{
				SlotType:            "chest_opening",
				SupportedOperations: []string{"chest_opening"},
				Count:               1,
			},
			{
				SlotType:            "smithy",
				SupportedOperations: []string{"crafting", "smelting"},
				Count:               2,
			},
			{
				SlotType:            "trade_purchase",
				SupportedOperations: []string{"trade_purchase"},
				Count:               1,
			},
		},
	}
}

// GetProductionModifiers возвращает нулевые модификаторы
func (s *MockDataService) GetProductionModifiers(userID string) models.ProductionModifiersResponse {
	return models.ProductionModifiersResponse{
		UserID: userID,
		Modifiers: models.ProductionModifiers{
			VIPStatus: models.VIPModifier{
				Level:                "none",
				ProductionSpeedBonus: 0.0,
				QualityBonus:         0.0,
			},
			CharacterLevel: models.CharacterLevelModifier{
				Level:         1,
				CraftingBonus: 0.0,
			},
			Achievements: []models.Achievement{},
			ClanBonuses: models.ClanBonuses{
				ProductionSpeed: 0.0,
			},
		},
	}
}
