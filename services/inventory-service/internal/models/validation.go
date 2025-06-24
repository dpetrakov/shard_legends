package models

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var (
	// ErrInvalidUUID indicates that a UUID is invalid
	ErrInvalidUUID = errors.New("invalid UUID")
	
	// ErrInvalidQuantity indicates that a quantity is invalid
	ErrInvalidQuantity = errors.New("invalid quantity")
	
	// ErrInvalidOperationType indicates that an operation type is invalid
	ErrInvalidOperationType = errors.New("invalid operation type")
	
	// ErrInvalidSection indicates that a section is invalid
	ErrInvalidSection = errors.New("invalid section")
	
	// ErrEmptyItems indicates that items list is empty
	ErrEmptyItems = errors.New("items list cannot be empty")
	
	// ErrInvalidCode indicates that a code contains invalid characters
	ErrInvalidCode = errors.New("invalid code format")
)

var (
	// codeRegex validates that codes contain only lowercase letters, numbers, and underscores
	codeRegex = regexp.MustCompile(`^[a-z0-9_]+$`)
	
	// validate is the validator instance
	validate = validator.New()
)

// ValidateUUID validates that a UUID is not zero
func ValidateUUID(id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: UUID cannot be nil", ErrInvalidUUID)
	}
	return nil
}

// ValidateCode validates a classifier or item code
func ValidateCode(code string) error {
	if code == "" {
		return fmt.Errorf("%w: code cannot be empty", ErrInvalidCode)
	}
	if len(code) > 50 {
		return fmt.Errorf("%w: code cannot exceed 50 characters", ErrInvalidCode)
	}
	if !codeRegex.MatchString(code) {
		return fmt.Errorf("%w: code must contain only lowercase letters, numbers, and underscores", ErrInvalidCode)
	}
	return nil
}

// ValidateItemQuantityRequest validates an ItemQuantityRequest
func ValidateItemQuantityRequest(req *ItemQuantityRequest) error {
	if err := ValidateUUID(req.ItemID); err != nil {
		return fmt.Errorf("item_id: %w", err)
	}
	
	if req.Quantity <= 0 {
		return fmt.Errorf("%w: quantity must be positive", ErrInvalidQuantity)
	}
	
	if req.Collection != nil {
		if err := ValidateCode(*req.Collection); err != nil {
			return fmt.Errorf("collection: %w", err)
		}
	}
	
	if req.QualityLevel != nil {
		if err := ValidateCode(*req.QualityLevel); err != nil {
			return fmt.Errorf("quality_level: %w", err)
		}
	}
	
	return nil
}

// ValidateOperation validates an Operation
func ValidateOperation(op *Operation) error {
	if err := ValidateUUID(op.UserID); err != nil {
		return fmt.Errorf("user_id: %w", err)
	}
	if err := ValidateUUID(op.SectionID); err != nil {
		return fmt.Errorf("section_id: %w", err)
	}
	if err := ValidateUUID(op.ItemID); err != nil {
		return fmt.Errorf("item_id: %w", err)
	}
	if err := ValidateUUID(op.CollectionID); err != nil {
		return fmt.Errorf("collection_id: %w", err)
	}
	if err := ValidateUUID(op.QualityLevelID); err != nil {
		return fmt.Errorf("quality_level_id: %w", err)
	}
	if err := ValidateUUID(op.OperationTypeID); err != nil {
		return fmt.Errorf("operation_type_id: %w", err)
	}
	
	if op.QuantityChange == 0 {
		return fmt.Errorf("%w: quantity_change cannot be zero", ErrInvalidQuantity)
	}
	
	if op.Comment != nil && len(*op.Comment) > 500 {
		return errors.New("comment cannot exceed 500 characters")
	}
	
	return nil
}

// ValidateReserveItemsRequest validates a ReserveItemsRequest
func ValidateReserveItemsRequest(req *ReserveItemsRequest) error {
	if err := validate.Struct(req); err != nil {
		return err
	}
	
	for i, item := range req.Items {
		if err := ValidateItemQuantityRequest(&item); err != nil {
			return fmt.Errorf("items[%d]: %w", i, err)
		}
	}
	
	return nil
}

// ValidateAddItemsRequest validates an AddItemsRequest
func ValidateAddItemsRequest(req *AddItemsRequest) error {
	if err := validate.Struct(req); err != nil {
		return err
	}
	
	// Validate section
	section := strings.ToLower(req.Section)
	if section != SectionMain && section != SectionFactory && section != SectionTrade {
		return fmt.Errorf("%w: %s", ErrInvalidSection, req.Section)
	}
	
	// Validate operation type code
	if err := ValidateCode(req.OperationType); err != nil {
		return fmt.Errorf("operation_type: %w", err)
	}
	
	// Validate each item
	for i, item := range req.Items {
		if err := ValidateItemQuantityRequest(&item); err != nil {
			return fmt.Errorf("items[%d]: %w", i, err)
		}
	}
	
	return nil
}

// ValidateAdjustInventoryRequest validates an AdjustInventoryRequest
func ValidateAdjustInventoryRequest(req *AdjustInventoryRequest) error {
	if err := validate.Struct(req); err != nil {
		return err
	}
	
	// Validate section
	section := strings.ToLower(req.Section)
	if section != SectionMain && section != SectionFactory && section != SectionTrade {
		return fmt.Errorf("%w: %s", ErrInvalidSection, req.Section)
	}
	
	// Validate each item
	for i, item := range req.Items {
		if err := ValidateUUID(item.ItemID); err != nil {
			return fmt.Errorf("items[%d].item_id: %w", i, err)
		}
		
		if item.QuantityChange == 0 {
			return fmt.Errorf("items[%d]: %w: quantity_change cannot be zero", i, ErrInvalidQuantity)
		}
		
		if item.Collection != nil {
			if err := ValidateCode(*item.Collection); err != nil {
				return fmt.Errorf("items[%d].collection: %w", i, err)
			}
		}
		
		if item.QualityLevel != nil {
			if err := ValidateCode(*item.QualityLevel); err != nil {
				return fmt.Errorf("items[%d].quality_level: %w", i, err)
			}
		}
	}
	
	return nil
}

// ValidateDailyBalance validates a DailyBalance
func ValidateDailyBalance(balance *DailyBalance) error {
	if err := ValidateUUID(balance.UserID); err != nil {
		return fmt.Errorf("user_id: %w", err)
	}
	if err := ValidateUUID(balance.SectionID); err != nil {
		return fmt.Errorf("section_id: %w", err)
	}
	if err := ValidateUUID(balance.ItemID); err != nil {
		return fmt.Errorf("item_id: %w", err)
	}
	if err := ValidateUUID(balance.CollectionID); err != nil {
		return fmt.Errorf("collection_id: %w", err)
	}
	if err := ValidateUUID(balance.QualityLevelID); err != nil {
		return fmt.Errorf("quality_level_id: %w", err)
	}
	
	if balance.Quantity < 0 {
		return fmt.Errorf("%w: quantity cannot be negative", ErrInvalidQuantity)
	}
	
	return nil
}