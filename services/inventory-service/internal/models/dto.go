package models

import (
	"github.com/google/uuid"
)

// ItemQuantityRequest represents a request for item quantity (with codes)
type ItemQuantityRequest struct {
	ItemID       uuid.UUID `json:"item_id" validate:"required"`
	Collection   *string   `json:"collection,omitempty"`
	QualityLevel *string   `json:"quality_level,omitempty"`
	Quantity     int64     `json:"quantity" validate:"required,min=1"`
}

// InventoryItemResponse represents an inventory item in API responses (with codes)
type InventoryItemResponse struct {
	ItemID       uuid.UUID `json:"item_id"`
	ItemClass    string    `json:"item_class"`
	ItemType     string    `json:"item_type"`
	Collection   *string   `json:"collection,omitempty"`
	QualityLevel *string   `json:"quality_level,omitempty"`
	Quantity     int64     `json:"quantity"`
}

// ItemBalance represents internal structure for balance calculations
type ItemBalance struct {
	UserID         uuid.UUID
	SectionID      uuid.UUID
	ItemID         uuid.UUID
	CollectionID   uuid.UUID
	QualityLevelID uuid.UUID
	CurrentBalance int64
}

// InventoryResponse represents the response for inventory queries
type InventoryResponse struct {
	Items []InventoryItemResponse `json:"items"`
}

// ReserveItemsRequest represents a request to reserve items
type ReserveItemsRequest struct {
	UserID      uuid.UUID             `json:"user_id" validate:"required"`
	OperationID uuid.UUID             `json:"operation_id" validate:"required"`
	Items       []ItemQuantityRequest `json:"items" validate:"required,min=1,dive"`
}

// ReturnReserveRequest represents a request to return reserved items
type ReturnReserveRequest struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	OperationID uuid.UUID `json:"operation_id" validate:"required"`
}

// ConsumeReserveRequest represents a request to consume reserved items
type ConsumeReserveRequest struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	OperationID uuid.UUID `json:"operation_id" validate:"required"`
}

// AddItemsRequest represents a request to add items to inventory
type AddItemsRequest struct {
	UserID        uuid.UUID             `json:"user_id" validate:"required"`
	Section       string                `json:"section" validate:"required,oneof=main factory trade"`
	OperationType string                `json:"operation_type" validate:"required"`
	OperationID   uuid.UUID             `json:"operation_id" validate:"required"`
	Items         []ItemQuantityRequest `json:"items" validate:"required,min=1,dive"`
	Comment       *string               `json:"comment,omitempty" validate:"omitempty,max=500"`
}

// AdjustInventoryRequest represents an admin inventory adjustment request
type AdjustInventoryRequest struct {
	UserID  uuid.UUID                    `json:"user_id" validate:"required"`
	Section string                       `json:"section" validate:"required,oneof=main factory trade"`
	Items   []AdjustInventoryItemRequest `json:"items" validate:"required,min=1,dive"`
	Reason  string                       `json:"reason" validate:"required,min=10,max=500"`
}

// AdjustInventoryItemRequest represents a single item in an adjustment request
type AdjustInventoryItemRequest struct {
	ItemID         uuid.UUID `json:"item_id" validate:"required"`
	Collection     *string   `json:"collection,omitempty"`
	QualityLevel   *string   `json:"quality_level,omitempty"`
	QuantityChange int64     `json:"quantity_change" validate:"required,ne=0"`
}

// OperationResponse represents a generic operation response
type OperationResponse struct {
	Success      bool        `json:"success"`
	OperationIDs []uuid.UUID `json:"operation_ids"`
}

// AdjustInventoryResponse represents the response for inventory adjustment
type AdjustInventoryResponse struct {
	Success       bool                    `json:"success"`
	OperationIDs  []uuid.UUID             `json:"operation_ids"`
	FinalBalances []InventoryItemResponse `json:"final_balances"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// InsufficientItemsError represents an error when there are not enough items
type InsufficientItemsError struct {
	ErrorCode    string        `json:"error"`
	Message      string        `json:"message"`
	MissingItems []MissingItem `json:"missing_items"`
}

// Error implements the error interface
func (e *InsufficientItemsError) Error() string {
	return e.Message
}

// MissingItem represents an item that is missing or insufficient
type MissingItem struct {
	ItemID       uuid.UUID `json:"item_id"`
	Collection   *string   `json:"collection,omitempty"`
	QualityLevel *string   `json:"quality_level,omitempty"`
	Required     int64     `json:"required"`
	Available    int64     `json:"available"`
}
