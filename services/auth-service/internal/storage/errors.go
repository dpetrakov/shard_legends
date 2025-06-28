package storage

import "fmt"

// DatabaseError represents a database-related error
type DatabaseError struct {
	Message string
	Cause   error
}

func (e *DatabaseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("database error: %s (caused by: %v)", e.Message, e.Cause)
	}
	return fmt.Sprintf("database error: %s", e.Message)
}

func (e *DatabaseError) Unwrap() error {
	return e.Cause
}
