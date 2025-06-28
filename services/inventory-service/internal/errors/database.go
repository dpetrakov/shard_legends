package errors

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
)

// PostgreSQL error codes
const (
	// Check violation (constraint failed)
	PgErrorCodeCheckViolation = "23514"
	// Unique violation
	PgErrorCodeUniqueViolation = "23505"
	// Foreign key violation
	PgErrorCodeForeignKeyViolation = "23503"
	// Not null violation
	PgErrorCodeNotNullViolation = "23502"
	// Lock not available (FOR UPDATE NOWAIT failed)
	PgErrorCodeLockNotAvailable = "55P03"
)

// InsufficientBalanceError represents insufficient balance during operations
type InsufficientBalanceError struct {
	UserID    string `json:"user_id"`
	ItemID    string `json:"item_id"`
	Requested int64  `json:"requested"`
	Available int64  `json:"available"`
	Message   string `json:"message"`
}

func (e *InsufficientBalanceError) Error() string {
	return e.Message
}

// ConcurrentOperationError represents a race condition or lock contention
type ConcurrentOperationError struct {
	Operation string `json:"operation"`
	Resource  string `json:"resource"`
	Message   string `json:"message"`
}

func (e *ConcurrentOperationError) Error() string {
	return e.Message
}

// HandleDatabaseError converts PostgreSQL errors to business-specific errors
func HandleDatabaseError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// Handle pgx.ErrNoRows
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.Wrap(err, "record not found")
	}

	// Handle PostgreSQL errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return handlePostgreSQLError(pgErr, operation)
	}

	// Handle string-based error messages (for compatibility)
	errMsg := err.Error()

	// Check for insufficient balance from trigger
	if strings.Contains(errMsg, "Insufficient balance") {
		return parseInsufficientBalanceError(errMsg)
	}

	// Check for lock contention
	if strings.Contains(errMsg, "could not obtain lock") {
		return &ConcurrentOperationError{
			Operation: operation,
			Resource:  "inventory_item",
			Message:   "Item is currently being processed by another transaction. Please retry.",
		}
	}

	// Return original error for unhandled cases
	return err
}

// handlePostgreSQLError handles specific PostgreSQL error codes
func handlePostgreSQLError(pgErr *pgconn.PgError, operation string) error {
	switch pgErr.Code {
	case PgErrorCodeCheckViolation:
		// This is likely our balance constraint
		if strings.Contains(pgErr.Message, "Insufficient balance") {
			return parseInsufficientBalanceError(pgErr.Message)
		}
		return errors.Errorf("constraint violation during %s: %s", operation, pgErr.Message)

	case PgErrorCodeLockNotAvailable:
		return &ConcurrentOperationError{
			Operation: operation,
			Resource:  extractResourceFromConstraint(pgErr.ConstraintName),
			Message:   "Resource is currently locked by another transaction. Please retry.",
		}

	case PgErrorCodeUniqueViolation:
		return errors.Errorf("duplicate %s: %s", operation, pgErr.Message)

	case PgErrorCodeForeignKeyViolation:
		return errors.Errorf("invalid reference during %s: %s", operation, pgErr.Message)

	case PgErrorCodeNotNullViolation:
		return errors.Errorf("missing required field during %s: %s", operation, pgErr.Message)

	default:
		return errors.Errorf("database error during %s: %s", operation, pgErr.Message)
	}
}

// parseInsufficientBalanceError extracts balance information from error message
// Expected format: "Insufficient balance for user <uuid> item <uuid>: requested <num>, available <num>"
func parseInsufficientBalanceError(message string) error {
	// Try to extract structured information from the error message
	userID := extractUUIDAfterText(message, "user ")
	itemID := extractUUIDAfterText(message, "item ")
	requested := extractNumberAfterText(message, "requested ")
	available := extractNumberAfterText(message, "available ")

	return &InsufficientBalanceError{
		UserID:    userID,
		ItemID:    itemID,
		Requested: requested,
		Available: available,
		Message:   message,
	}
}

// extractResourceFromConstraint extracts resource name from constraint name
func extractResourceFromConstraint(constraintName string) string {
	if constraintName == "" {
		return "unknown_resource"
	}

	// Convert constraint names to resource names
	switch {
	case strings.Contains(constraintName, "balance"):
		return "inventory_balance"
	case strings.Contains(constraintName, "operation"):
		return "inventory_operation"
	case strings.Contains(constraintName, "daily"):
		return "daily_balance"
	default:
		return constraintName
	}
}

// extractUUIDAfterText extracts UUID that appears after specified text
func extractUUIDAfterText(message, text string) string {
	index := strings.Index(message, text)
	if index == -1 {
		return ""
	}

	start := index + len(text)
	if start >= len(message) {
		return ""
	}

	// Look for space or colon that ends the UUID
	end := start
	for end < len(message) && message[end] != ' ' && message[end] != ':' {
		end++
	}

	if end > start {
		return message[start:end]
	}

	return ""
}

// extractNumberAfterText extracts number that appears after specified text
func extractNumberAfterText(message, text string) int64 {
	index := strings.Index(message, text)
	if index == -1 {
		return 0
	}

	start := index + len(text)
	if start >= len(message) {
		return 0
	}

	// Extract digits
	end := start
	for end < len(message) && message[end] >= '0' && message[end] <= '9' {
		end++
	}

	if end > start {
		var num int64
		_, err := fmt.Sscanf(message[start:end], "%d", &num)
		if err == nil {
			return num
		}
	}

	return 0
}

// IsInsufficientBalanceError checks if error is an insufficient balance error
func IsInsufficientBalanceError(err error) bool {
	var balanceErr *InsufficientBalanceError
	return errors.As(err, &balanceErr)
}

// IsConcurrentOperationError checks if error is a concurrent operation error
func IsConcurrentOperationError(err error) bool {
	var concurrentErr *ConcurrentOperationError
	return errors.As(err, &concurrentErr)
}

// GetInsufficientBalanceDetails extracts details from insufficient balance error
func GetInsufficientBalanceDetails(err error) (*InsufficientBalanceError, bool) {
	var balanceErr *InsufficientBalanceError
	if errors.As(err, &balanceErr) {
		return balanceErr, true
	}
	return nil, false
}

// GetConcurrentOperationDetails extracts details from concurrent operation error
func GetConcurrentOperationDetails(err error) (*ConcurrentOperationError, bool) {
	var concurrentErr *ConcurrentOperationError
	if errors.As(err, &concurrentErr) {
		return concurrentErr, true
	}
	return nil, false
}
