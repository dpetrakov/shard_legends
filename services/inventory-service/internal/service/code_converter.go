package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	
	"github.com/shard-legends/inventory-service/internal/models"
)

// codeConverter implements CodeConverter interface
type codeConverter struct {
	deps *ServiceDependencies
}

// NewCodeConverter creates a new code converter
func NewCodeConverter(deps *ServiceDependencies) CodeConverter {
	return &codeConverter{
		deps: deps,
	}
}

// ConvertClassifierCodes converts between classifier codes and UUIDs
func (cc *codeConverter) ConvertClassifierCodes(ctx context.Context, req *CodeConversionRequest) (*CodeConversionResponse, error) {
	if req == nil {
		return nil, errors.New("code conversion request cannot be nil")
	}

	if req.Direction != "toUUID" && req.Direction != "fromUUID" {
		return nil, errors.New("direction must be 'toUUID' or 'fromUUID'")
	}

	if req.Data == nil {
		return nil, errors.New("data cannot be nil")
	}

	// Create a copy of the input data for conversion
	result := make(map[string]interface{})
	for k, v := range req.Data {
		result[k] = v
	}

	if req.Direction == "toUUID" {
		err := cc.convertCodesToUUIDs(ctx, result)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert codes to UUIDs")
		}
	} else {
		err := cc.convertUUIDsToCodes(ctx, result)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert UUIDs to codes")
		}
	}

	return &CodeConversionResponse{
		Data: result,
	}, nil
}

// convertCodesToUUIDs converts classifier codes to UUIDs in the data map
func (cc *codeConverter) convertCodesToUUIDs(ctx context.Context, data map[string]interface{}) error {
	// Define mapping of field names to their classifier types
	fieldMappings := map[string]string{
		"section":         models.ClassifierInventorySection,
		"item_class":      models.ClassifierItemClass,
		"item_type":       models.ClassifierResourceType, // This would need to be determined dynamically
		"collection":      models.ClassifierCollection,
		"quality_level":   models.ClassifierQualityLevel,
		"operation_type":  models.ClassifierOperationType,
	}

	for fieldName, classifierType := range fieldMappings {
		if codeValue, exists := data[fieldName]; exists && codeValue != nil {
			codeStr, ok := codeValue.(string)
			if !ok {
				continue // Skip non-string values
			}

			if codeStr == "" {
				continue // Skip empty strings
			}

			// Get the UUID mapping for this classifier
			mapping, err := cc.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, classifierType)
			if err != nil {
				return errors.Wrapf(err, "failed to get code mapping for classifier %s", classifierType)
			}

			// Convert code to UUID
			if uuid, found := mapping[codeStr]; found {
				data[fieldName+"_id"] = uuid
			} else {
				// Handle unknown codes - use default if available
				defaultUUID, hasDefault := cc.getDefaultUUID(classifierType, codeStr)
				if hasDefault {
					data[fieldName+"_id"] = defaultUUID
				} else {
					return errors.Errorf("unknown code '%s' for classifier %s", codeStr, classifierType)
				}
			}

			// Remove the original code field
			delete(data, fieldName)
		}
	}

	return nil
}

// convertUUIDsToCodes converts UUIDs to classifier codes in the data map
func (cc *codeConverter) convertUUIDsToCodes(ctx context.Context, data map[string]interface{}) error {
	// Define mapping of field names to their classifier types
	fieldMappings := map[string]string{
		"section_id":         models.ClassifierInventorySection,
		"item_class_id":      models.ClassifierItemClass,
		"item_type_id":       models.ClassifierResourceType, // This would need to be determined dynamically
		"collection_id":      models.ClassifierCollection,
		"quality_level_id":   models.ClassifierQualityLevel,
		"operation_type_id":  models.ClassifierOperationType,
	}

	for fieldName, classifierType := range fieldMappings {
		if uuidValue, exists := data[fieldName]; exists && uuidValue != nil {
			var uuidObj uuid.UUID
			var ok bool

			// Handle different UUID representations
			switch v := uuidValue.(type) {
			case uuid.UUID:
				uuidObj = v
				ok = true
			case string:
				var err error
				uuidObj, err = uuid.Parse(v)
				ok = err == nil
			}

			if !ok {
				continue // Skip invalid UUIDs
			}

			// Get the reverse mapping for this classifier
			mapping, err := cc.deps.Repositories.Classifier.GetUUIDToCodeMapping(ctx, classifierType)
			if err != nil {
				return errors.Wrapf(err, "failed to get UUID mapping for classifier %s", classifierType)
			}

			// Convert UUID to code
			if code, found := mapping[uuidObj]; found {
				// Remove the "_id" suffix from field name
				codeFieldName := strings.TrimSuffix(fieldName, "_id")
				data[codeFieldName] = code
			} else {
				return errors.Errorf("unknown UUID '%s' for classifier %s", uuidObj.String(), classifierType)
			}

			// Remove the original UUID field
			delete(data, fieldName)
		}
	}

	return nil
}

// getDefaultUUID returns a default UUID for unknown codes based on classifier type
func (cc *codeConverter) getDefaultUUID(classifierType, code string) (uuid.UUID, bool) {
	// For collections and quality levels, we might have defaults
	switch classifierType {
	case models.ClassifierCollection:
		// Use a standard collection UUID or return the first available one
		return uuid.Nil, false // No default for now
	case models.ClassifierQualityLevel:
		// Default quality level could be "common" or similar
		return uuid.Nil, false // No default for now
	default:
		return uuid.Nil, false
	}
}

// ConvertItemQuantityRequest converts codes in ItemQuantityRequest to UUIDs
func (cc *codeConverter) ConvertItemQuantityRequest(ctx context.Context, req *models.ItemQuantityRequest) error {
	if req == nil {
		return errors.New("item quantity request cannot be nil")
	}

	// Convert collection code to UUID if provided
	if req.Collection != nil && *req.Collection != "" {
		mapping, err := cc.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierCollection)
		if err != nil {
			return errors.Wrap(err, "failed to get collection mapping")
		}

		if collectionUUID, found := mapping[*req.Collection]; found {
			// Store the UUID (this would need to be added to the struct)
			_ = collectionUUID // TODO: Add CollectionID field to ItemQuantityRequest
		} else {
			return errors.Errorf("unknown collection code: %s", *req.Collection)
		}
	}

	// Convert quality level code to UUID if provided
	if req.QualityLevel != nil && *req.QualityLevel != "" {
		mapping, err := cc.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, models.ClassifierQualityLevel)
		if err != nil {
			return errors.Wrap(err, "failed to get quality level mapping")
		}

		if qualityUUID, found := mapping[*req.QualityLevel]; found {
			// Store the UUID (this would need to be added to the struct)
			_ = qualityUUID // TODO: Add QualityLevelID field to ItemQuantityRequest
		} else {
			return errors.Errorf("unknown quality level code: %s", *req.QualityLevel)
		}
	}

	return nil
}

// ConvertInventoryResponse converts UUIDs in InventoryItemResponse to codes
func (cc *codeConverter) ConvertInventoryResponse(ctx context.Context, items []*models.InventoryItemResponse) error {
	if len(items) == 0 {
		return nil
	}

	// Note: In the current models.InventoryItemResponse, Collection and QualityLevel are already *string
	// This suggests they should contain codes, not UUIDs
	// The conversion would happen in the repository layer when building the response
	
	// For future implementation when UUIDs need to be converted to codes
	_ = ctx  // Use context parameter
	_ = items // Use items parameter

	return nil
}

// Helper function to safely get string value from interface{}
func getStringValue(value interface{}) (string, bool) {
	if value == nil {
		return "", false
	}
	
	str, ok := value.(string)
	return str, ok && str != ""
}

// Helper function to safely get UUID value from interface{}
func getUUIDValue(value interface{}) (uuid.UUID, bool) {
	if value == nil {
		return uuid.Nil, false
	}
	
	switch v := value.(type) {
	case uuid.UUID:
		return v, true
	case string:
		if parsed, err := uuid.Parse(v); err == nil {
			return parsed, true
		}
	}
	
	return uuid.Nil, false
}