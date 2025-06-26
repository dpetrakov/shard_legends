package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shard-legends/production-service/internal/models"
	"github.com/shard-legends/production-service/internal/storage"
)

// codeConverterService реализует CodeConverterService
type codeConverterService struct {
	repository *storage.Repository
	cache      storage.CacheInterface
	metrics    storage.MetricsInterface
}

// NewCodeConverterService создает новый экземпляр сервиса преобразования кодов
func NewCodeConverterService(deps *ServiceDependencies) CodeConverterService {
	return &codeConverterService{
		repository: deps.Repository,
		cache:      deps.Cache,
		metrics:    deps.Metrics,
	}
}

// ConvertRecipeFromCodes преобразует коды в UUID для рецепта
func (s *codeConverterService) ConvertRecipeFromCodes(ctx context.Context, recipe *models.ProductionRecipe) error {
	// Преобразуем входные предметы
	for i := range recipe.InputItems {
		input := &recipe.InputItems[i]

		// Преобразуем коллекцию
		if input.CollectionCode != nil {
			collectionID, err := s.ConvertCodeToUUID(ctx, storage.ClassifierCollection, *input.CollectionCode)
			if err != nil {
				return fmt.Errorf("failed to convert collection code '%s': %w", *input.CollectionCode, err)
			}
			input.CollectionID = collectionID
		}

		// Преобразуем уровень качества
		if input.QualityLevelCode != nil {
			qualityID, err := s.ConvertCodeToUUID(ctx, storage.ClassifierQualityLevel, *input.QualityLevelCode)
			if err != nil {
				return fmt.Errorf("failed to convert quality level code '%s': %w", *input.QualityLevelCode, err)
			}
			input.QualityLevelID = qualityID
		}
	}

	// Преобразуем выходные предметы
	for i := range recipe.OutputItems {
		output := &recipe.OutputItems[i]

		// Преобразуем фиксированную коллекцию
		if output.FixedCollectionCode != nil {
			collectionID, err := s.ConvertCodeToUUID(ctx, storage.ClassifierCollection, *output.FixedCollectionCode)
			if err != nil {
				return fmt.Errorf("failed to convert fixed collection code '%s': %w", *output.FixedCollectionCode, err)
			}
			output.FixedCollectionID = collectionID
		}

		// Преобразуем фиксированный уровень качества
		if output.FixedQualityLevelCode != nil {
			qualityID, err := s.ConvertCodeToUUID(ctx, storage.ClassifierQualityLevel, *output.FixedQualityLevelCode)
			if err != nil {
				return fmt.Errorf("failed to convert fixed quality level code '%s': %w", *output.FixedQualityLevelCode, err)
			}
			output.FixedQualityLevelID = qualityID
		}
	}

	return nil
}

// ConvertRecipeToCodes преобразует UUID в коды для рецепта
func (s *codeConverterService) ConvertRecipeToCodes(ctx context.Context, recipe *models.ProductionRecipe) error {
	// Преобразуем входные предметы
	for i := range recipe.InputItems {
		input := &recipe.InputItems[i]

		// Преобразуем коллекцию
		if input.CollectionID != nil {
			collectionCode, err := s.ConvertUUIDToCode(ctx, storage.ClassifierCollection, *input.CollectionID)
			if err != nil {
				return fmt.Errorf("failed to convert collection ID '%s': %w", *input.CollectionID, err)
			}
			input.CollectionCode = collectionCode
		}

		// Преобразуем уровень качества
		if input.QualityLevelID != nil {
			qualityCode, err := s.ConvertUUIDToCode(ctx, storage.ClassifierQualityLevel, *input.QualityLevelID)
			if err != nil {
				return fmt.Errorf("failed to convert quality level ID '%s': %w", *input.QualityLevelID, err)
			}
			input.QualityLevelCode = qualityCode
		}
	}

	// Преобразуем выходные предметы
	for i := range recipe.OutputItems {
		output := &recipe.OutputItems[i]

		// Преобразуем фиксированную коллекцию
		if output.FixedCollectionID != nil {
			collectionCode, err := s.ConvertUUIDToCode(ctx, storage.ClassifierCollection, *output.FixedCollectionID)
			if err != nil {
				return fmt.Errorf("failed to convert fixed collection ID '%s': %w", *output.FixedCollectionID, err)
			}
			output.FixedCollectionCode = collectionCode
		}

		// Преобразуем фиксированный уровень качества
		if output.FixedQualityLevelID != nil {
			qualityCode, err := s.ConvertUUIDToCode(ctx, storage.ClassifierQualityLevel, *output.FixedQualityLevelID)
			if err != nil {
				return fmt.Errorf("failed to convert fixed quality level ID '%s': %w", *output.FixedQualityLevelID, err)
			}
			output.FixedQualityLevelCode = qualityCode
		}
	}

	return nil
}

// ConvertCodeToUUID преобразует код в UUID
func (s *codeConverterService) ConvertCodeToUUID(ctx context.Context, classifierName, code string) (*uuid.UUID, error) {
	return s.repository.Classifier.ConvertCodeToUUID(ctx, classifierName, code)
}

// ConvertUUIDToCode преобразует UUID в код
func (s *codeConverterService) ConvertUUIDToCode(ctx context.Context, classifierName string, id uuid.UUID) (*string, error) {
	return s.repository.Classifier.ConvertUUIDToCode(ctx, classifierName, id)
}

// ConvertTaskOutputItemToCodes преобразует UUID в коды для выходного предмета задания
func (s *codeConverterService) ConvertTaskOutputItemToCodes(ctx context.Context, item *models.TaskOutputItem) error {
	// Преобразуем коллекцию
	if item.CollectionID != nil {
		collectionCode, err := s.ConvertUUIDToCode(ctx, storage.ClassifierCollection, *item.CollectionID)
		if err != nil {
			return fmt.Errorf("failed to convert collection ID '%s': %w", *item.CollectionID, err)
		}
		item.CollectionCode = collectionCode
	}

	// Преобразуем уровень качества
	if item.QualityLevelID != nil {
		qualityCode, err := s.ConvertUUIDToCode(ctx, storage.ClassifierQualityLevel, *item.QualityLevelID)
		if err != nil {
			return fmt.Errorf("failed to convert quality level ID '%s': %w", *item.QualityLevelID, err)
		}
		item.QualityLevelCode = qualityCode
	}

	return nil
}

// ConvertTaskOutputItemFromCodes преобразует коды в UUID для выходного предмета задания
func (s *codeConverterService) ConvertTaskOutputItemFromCodes(ctx context.Context, item *models.TaskOutputItem) error {
	// Преобразуем коллекцию
	if item.CollectionCode != nil {
		collectionID, err := s.ConvertCodeToUUID(ctx, storage.ClassifierCollection, *item.CollectionCode)
		if err != nil {
			return fmt.Errorf("failed to convert collection code '%s': %w", *item.CollectionCode, err)
		}
		item.CollectionID = collectionID
	}

	// Преобразуем уровень качества
	if item.QualityLevelCode != nil {
		qualityID, err := s.ConvertCodeToUUID(ctx, storage.ClassifierQualityLevel, *item.QualityLevelCode)
		if err != nil {
			return fmt.Errorf("failed to convert quality level code '%s': %w", *item.QualityLevelCode, err)
		}
		item.QualityLevelID = qualityID
	}

	return nil
}
