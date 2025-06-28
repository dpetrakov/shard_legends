package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name:    "valid UUID",
			id:      uuid.New(),
			wantErr: false,
		},
		{
			name:    "nil UUID",
			id:      uuid.Nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUUID(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidUUID)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "valid code",
			code:    "test_code_123",
			wantErr: false,
		},
		{
			name:    "empty code",
			code:    "",
			wantErr: true,
		},
		{
			name:    "code with uppercase",
			code:    "Test_Code",
			wantErr: true,
		},
		{
			name:    "code with special chars",
			code:    "test-code",
			wantErr: true,
		},
		{
			name:    "code too long",
			code:    "very_long_code_that_exceeds_fifty_characters_limit_and_should_fail",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCode(tt.code)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidCode)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateItemQuantityRequest(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name    string
		req     *ItemQuantityRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &ItemQuantityRequest{
				ItemID:   validUUID,
				Quantity: 10,
			},
			wantErr: false,
		},
		{
			name: "valid request with collection",
			req: &ItemQuantityRequest{
				ItemID:     validUUID,
				Quantity:   5,
				Collection: stringPtr("winter_2025"),
			},
			wantErr: false,
		},
		{
			name: "invalid item ID",
			req: &ItemQuantityRequest{
				ItemID:   uuid.Nil,
				Quantity: 10,
			},
			wantErr: true,
		},
		{
			name: "zero quantity",
			req: &ItemQuantityRequest{
				ItemID:   validUUID,
				Quantity: 0,
			},
			wantErr: true,
		},
		{
			name: "negative quantity",
			req: &ItemQuantityRequest{
				ItemID:   validUUID,
				Quantity: -5,
			},
			wantErr: true,
		},
		{
			name: "invalid collection code",
			req: &ItemQuantityRequest{
				ItemID:     validUUID,
				Quantity:   10,
				Collection: stringPtr("Invalid-Collection"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateItemQuantityRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOperation(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name    string
		op      *Operation
		wantErr bool
	}{
		{
			name: "valid operation positive",
			op: &Operation{
				UserID:          validUUID,
				SectionID:       validUUID,
				ItemID:          validUUID,
				CollectionID:    validUUID,
				QualityLevelID:  validUUID,
				QuantityChange:  10,
				OperationTypeID: validUUID,
			},
			wantErr: false,
		},
		{
			name: "valid operation negative",
			op: &Operation{
				UserID:          validUUID,
				SectionID:       validUUID,
				ItemID:          validUUID,
				CollectionID:    validUUID,
				QualityLevelID:  validUUID,
				QuantityChange:  -5,
				OperationTypeID: validUUID,
			},
			wantErr: false,
		},
		{
			name: "zero quantity change",
			op: &Operation{
				UserID:          validUUID,
				SectionID:       validUUID,
				ItemID:          validUUID,
				CollectionID:    validUUID,
				QualityLevelID:  validUUID,
				QuantityChange:  0,
				OperationTypeID: validUUID,
			},
			wantErr: true,
		},
		{
			name: "invalid user ID",
			op: &Operation{
				UserID:          uuid.Nil,
				SectionID:       validUUID,
				ItemID:          validUUID,
				CollectionID:    validUUID,
				QualityLevelID:  validUUID,
				QuantityChange:  10,
				OperationTypeID: validUUID,
			},
			wantErr: true,
		},
		{
			name: "comment too long",
			op: &Operation{
				UserID:          validUUID,
				SectionID:       validUUID,
				ItemID:          validUUID,
				CollectionID:    validUUID,
				QualityLevelID:  validUUID,
				QuantityChange:  10,
				OperationTypeID: validUUID,
				Comment:         stringPtr(string(make([]byte, 501))),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOperation(tt.op)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAddItemsRequest(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name    string
		req     *AddItemsRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &AddItemsRequest{
				UserID:        validUUID,
				Section:       "main",
				OperationType: "chest_reward",
				OperationID:   validUUID,
				Items: []ItemQuantityRequest{
					{
						ItemID:   validUUID,
						Quantity: 10,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid section",
			req: &AddItemsRequest{
				UserID:        validUUID,
				Section:       "invalid",
				OperationType: "chest_reward",
				OperationID:   validUUID,
				Items: []ItemQuantityRequest{
					{
						ItemID:   validUUID,
						Quantity: 10,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty items",
			req: &AddItemsRequest{
				UserID:        validUUID,
				Section:       "main",
				OperationType: "chest_reward",
				OperationID:   validUUID,
				Items:         []ItemQuantityRequest{},
			},
			wantErr: true,
		},
		{
			name: "invalid operation type",
			req: &AddItemsRequest{
				UserID:        validUUID,
				Section:       "main",
				OperationType: "Invalid-Type",
				OperationID:   validUUID,
				Items: []ItemQuantityRequest{
					{
						ItemID:   validUUID,
						Quantity: 10,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAddItemsRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDailyBalance(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name    string
		balance *DailyBalance
		wantErr bool
	}{
		{
			name: "valid balance",
			balance: &DailyBalance{
				UserID:         validUUID,
				SectionID:      validUUID,
				ItemID:         validUUID,
				CollectionID:   validUUID,
				QualityLevelID: validUUID,
				Quantity:       100,
			},
			wantErr: false,
		},
		{
			name: "zero quantity is valid",
			balance: &DailyBalance{
				UserID:         validUUID,
				SectionID:      validUUID,
				ItemID:         validUUID,
				CollectionID:   validUUID,
				QualityLevelID: validUUID,
				Quantity:       0,
			},
			wantErr: false,
		},
		{
			name: "negative quantity",
			balance: &DailyBalance{
				UserID:         validUUID,
				SectionID:      validUUID,
				ItemID:         validUUID,
				CollectionID:   validUUID,
				QualityLevelID: validUUID,
				Quantity:       -10,
			},
			wantErr: true,
		},
		{
			name: "invalid user ID",
			balance: &DailyBalance{
				UserID:         uuid.Nil,
				SectionID:      validUUID,
				ItemID:         validUUID,
				CollectionID:   validUUID,
				QualityLevelID: validUUID,
				Quantity:       100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDailyBalance(tt.balance)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateReserveItemsRequest(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name    string
		req     *ReserveItemsRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &ReserveItemsRequest{
				UserID:      validUUID,
				OperationID: validUUID,
				Items: []ItemQuantityRequest{
					{
						ItemID:   validUUID,
						Quantity: 10,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing user ID",
			req: &ReserveItemsRequest{
				OperationID: validUUID,
				Items: []ItemQuantityRequest{
					{
						ItemID:   validUUID,
						Quantity: 10,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing operation ID",
			req: &ReserveItemsRequest{
				UserID: validUUID,
				Items: []ItemQuantityRequest{
					{
						ItemID:   validUUID,
						Quantity: 10,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty items",
			req: &ReserveItemsRequest{
				UserID:      validUUID,
				OperationID: validUUID,
				Items:       []ItemQuantityRequest{},
			},
			wantErr: true,
		},
		{
			name: "invalid item in list",
			req: &ReserveItemsRequest{
				UserID:      validUUID,
				OperationID: validUUID,
				Items: []ItemQuantityRequest{
					{
						ItemID:   uuid.Nil,
						Quantity: 10,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateReserveItemsRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAdjustInventoryRequest(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name    string
		req     *AdjustInventoryRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &AdjustInventoryRequest{
				UserID:  validUUID,
				Section: "main",
				Items: []AdjustInventoryItemRequest{
					{
						ItemID:         validUUID,
						QuantityChange: 10,
					},
				},
				Reason: "Test adjustment for validation",
			},
			wantErr: false,
		},
		{
			name: "invalid section",
			req: &AdjustInventoryRequest{
				UserID:  validUUID,
				Section: "invalid_section",
				Items: []AdjustInventoryItemRequest{
					{
						ItemID:         validUUID,
						QuantityChange: 10,
					},
				},
				Reason: "Test adjustment for validation",
			},
			wantErr: true,
		},
		{
			name: "zero quantity change",
			req: &AdjustInventoryRequest{
				UserID:  validUUID,
				Section: "main",
				Items: []AdjustInventoryItemRequest{
					{
						ItemID:         validUUID,
						QuantityChange: 0,
					},
				},
				Reason: "Test adjustment for validation",
			},
			wantErr: true,
		},
		{
			name: "invalid item ID",
			req: &AdjustInventoryRequest{
				UserID:  validUUID,
				Section: "main",
				Items: []AdjustInventoryItemRequest{
					{
						ItemID:         uuid.Nil,
						QuantityChange: 10,
					},
				},
				Reason: "Test adjustment for validation",
			},
			wantErr: true,
		},
		{
			name: "invalid collection code",
			req: &AdjustInventoryRequest{
				UserID:  validUUID,
				Section: "main",
				Items: []AdjustInventoryItemRequest{
					{
						ItemID:         validUUID,
						Collection:     stringPtr("Invalid-Collection"),
						QuantityChange: 10,
					},
				},
				Reason: "Test adjustment for validation",
			},
			wantErr: true,
		},
		{
			name: "invalid quality level code",
			req: &AdjustInventoryRequest{
				UserID:  validUUID,
				Section: "main",
				Items: []AdjustInventoryItemRequest{
					{
						ItemID:         validUUID,
						QualityLevel:   stringPtr("Invalid-Quality"),
						QuantityChange: 10,
					},
				},
				Reason: "Test adjustment for validation",
			},
			wantErr: true,
		},
		{
			name: "reason too short",
			req: &AdjustInventoryRequest{
				UserID:  validUUID,
				Section: "main",
				Items: []AdjustInventoryItemRequest{
					{
						ItemID:         validUUID,
						QuantityChange: 10,
					},
				},
				Reason: "short",
			},
			wantErr: true,
		},
		{
			name: "reason too long",
			req: &AdjustInventoryRequest{
				UserID:  validUUID,
				Section: "main",
				Items: []AdjustInventoryItemRequest{
					{
						ItemID:         validUUID,
						QuantityChange: 10,
					},
				},
				Reason: string(make([]byte, 501)),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAdjustInventoryRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
