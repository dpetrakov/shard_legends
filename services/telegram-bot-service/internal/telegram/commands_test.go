package telegram

import (
	"testing"
)

func TestBot_handleCommand(t *testing.T) {
	// Skip tests that require real API calls in unit tests
	// These tests should be moved to integration tests
	t.Skip("Skipping tests that require real API calls - should be integration tests")
}