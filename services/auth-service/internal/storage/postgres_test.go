package storage

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shard-legends/auth-service/internal/models"
)

// TestPostgresStorage contains the test suite for PostgresStorage
type TestPostgresStorage struct {
	storage *PostgresStorage
	pool    *pgxpool.Pool
}

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *TestPostgresStorage {
	// Skip integration tests if not in CI or if database is not available
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping PostgreSQL integration tests")
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // Reduce noise in tests
	}))

	storage, err := NewPostgresStorage(databaseURL, 5, logger, nil) // Pass nil for metrics in tests
	if err != nil {
		t.Fatalf("Failed to create test storage: %v", err)
	}

	// Clean up test data before each test
	cleanupTestData(t, storage.pool)

	return &TestPostgresStorage{
		storage: storage,
		pool:    storage.pool,
	}
}

// cleanupTestData removes test data from the database
func cleanupTestData(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()
	_, err := pool.Exec(ctx, "DELETE FROM auth.users WHERE telegram_id >= 999999999")
	if err != nil {
		t.Logf("Warning: Failed to cleanup test data: %v", err)
	}
}

// teardownTestDB closes the test database connection
func (ts *TestPostgresStorage) teardownTestDB() {
	if ts.storage != nil {
		ts.storage.Close()
	}
}

// createTestUser creates a test user request
func createTestUser(telegramID int64) *models.CreateUserRequest {
	username := "test_user"
	lastName := "Doe"
	languageCode := "en"
	photoURL := "https://example.com/photo.jpg"

	return &models.CreateUserRequest{
		TelegramID:   telegramID,
		Username:     &username,
		FirstName:    "John",
		LastName:     &lastName,
		LanguageCode: &languageCode,
		IsPremium:    false,
		PhotoURL:     &photoURL,
	}
}

func TestNewPostgresStorage(t *testing.T) {
	tests := []struct {
		name        string
		databaseURL string
		maxConns    int
		wantErr     bool
	}{
		{
			name:        "valid database URL",
			databaseURL: os.Getenv("TEST_DATABASE_URL"),
			maxConns:    5,
			wantErr:     false,
		},
		{
			name:        "invalid database URL",
			databaseURL: "invalid-url",
			maxConns:    5,
			wantErr:     true,
		},
		{
			name:        "empty database URL",
			databaseURL: "",
			maxConns:    5,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid database URL" && tt.databaseURL == "" {
				t.Skip("TEST_DATABASE_URL not set")
			}

			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelError,
			}))

			storage, err := NewPostgresStorage(tt.databaseURL, tt.maxConns, logger, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPostgresStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if storage != nil {
				storage.Close()
			}
		})
	}
}

func TestPostgresStorage_CreateUser(t *testing.T) {
	ts := setupTestDB(t)
	defer ts.teardownTestDB()

	ctx := context.Background()

	tests := []struct {
		name    string
		req     *models.CreateUserRequest
		wantErr bool
		errType error
	}{
		{
			name:    "valid user creation",
			req:     createTestUser(999999999),
			wantErr: false,
		},
		{
			name:    "duplicate telegram_id",
			req:     createTestUser(999999999), // Same as above
			wantErr: true,
			errType: ErrUserAlreadyExists,
		},
		{
			name: "minimal user data",
			req: &models.CreateUserRequest{
				TelegramID: 999999998,
				FirstName:  "Jane",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := ts.storage.CreateUser(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errType != nil && err != tt.errType {
					t.Errorf("CreateUser() error = %v, expected %v", err, tt.errType)
				}
				return
			}

			// Verify created user
			if user == nil {
				t.Error("CreateUser() returned nil user")
				return
			}

			if user.ID == uuid.Nil {
				t.Error("CreateUser() user ID is nil")
			}
			if user.TelegramID != tt.req.TelegramID {
				t.Errorf("CreateUser() telegram_id = %v, want %v", user.TelegramID, tt.req.TelegramID)
			}
			if user.FirstName != tt.req.FirstName {
				t.Errorf("CreateUser() first_name = %v, want %v", user.FirstName, tt.req.FirstName)
			}
			if !user.IsActive {
				t.Error("CreateUser() user should be active by default")
			}
			if user.CreatedAt.IsZero() {
				t.Error("CreateUser() created_at should not be zero")
			}
		})
	}
}

func TestPostgresStorage_GetUserByID(t *testing.T) {
	ts := setupTestDB(t)
	defer ts.teardownTestDB()

	ctx := context.Background()

	// Create a test user first
	req := createTestUser(999999997)
	createdUser, err := ts.storage.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
		errType error
	}{
		{
			name:    "existing user",
			id:      createdUser.ID,
			wantErr: false,
		},
		{
			name:    "non-existent user",
			id:      uuid.New(),
			wantErr: true,
			errType: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := ts.storage.GetUserByID(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errType != nil && err != tt.errType {
					t.Errorf("GetUserByID() error = %v, expected %v", err, tt.errType)
				}
				return
			}

			// Verify retrieved user
			if user == nil {
				t.Error("GetUserByID() returned nil user")
				return
			}

			if user.ID != tt.id {
				t.Errorf("GetUserByID() id = %v, want %v", user.ID, tt.id)
			}
			if user.TelegramID != createdUser.TelegramID {
				t.Errorf("GetUserByID() telegram_id = %v, want %v", user.TelegramID, createdUser.TelegramID)
			}
		})
	}
}

func TestPostgresStorage_GetUserByTelegramID(t *testing.T) {
	ts := setupTestDB(t)
	defer ts.teardownTestDB()

	ctx := context.Background()

	// Create a test user first
	telegramID := int64(999999996)
	req := createTestUser(telegramID)
	createdUser, err := ts.storage.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name       string
		telegramID int64
		wantErr    bool
		errType    error
	}{
		{
			name:       "existing user",
			telegramID: telegramID,
			wantErr:    false,
		},
		{
			name:       "non-existent user",
			telegramID: 888888888,
			wantErr:    true,
			errType:    ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := ts.storage.GetUserByTelegramID(ctx, tt.telegramID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByTelegramID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errType != nil && err != tt.errType {
					t.Errorf("GetUserByTelegramID() error = %v, expected %v", err, tt.errType)
				}
				return
			}

			// Verify retrieved user
			if user == nil {
				t.Error("GetUserByTelegramID() returned nil user")
				return
			}

			if user.TelegramID != tt.telegramID {
				t.Errorf("GetUserByTelegramID() telegram_id = %v, want %v", user.TelegramID, tt.telegramID)
			}
			if user.ID != createdUser.ID {
				t.Errorf("GetUserByTelegramID() id = %v, want %v", user.ID, createdUser.ID)
			}
		})
	}
}

func TestPostgresStorage_UpdateUser(t *testing.T) {
	ts := setupTestDB(t)
	defer ts.teardownTestDB()

	ctx := context.Background()

	// Create a test user first
	req := createTestUser(999999995)
	createdUser, err := ts.storage.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name       string
		id         uuid.UUID
		updateReq  *models.UpdateUserRequest
		wantErr    bool
		errType    error
		checkField func(*models.User) bool
	}{
		{
			name: "update username",
			id:   createdUser.ID,
			updateReq: &models.UpdateUserRequest{
				Username: stringPtr("updated_user"),
			},
			wantErr: false,
			checkField: func(u *models.User) bool {
				return u.Username != nil && *u.Username == "updated_user"
			},
		},
		{
			name: "update first name",
			id:   createdUser.ID,
			updateReq: &models.UpdateUserRequest{
				FirstName: stringPtr("UpdatedName"),
			},
			wantErr: false,
			checkField: func(u *models.User) bool {
				return u.FirstName == "UpdatedName"
			},
		},
		{
			name: "update premium status",
			id:   createdUser.ID,
			updateReq: &models.UpdateUserRequest{
				IsPremium: boolPtr(true),
			},
			wantErr: false,
			checkField: func(u *models.User) bool {
				return u.IsPremium == true
			},
		},
		{
			name: "non-existent user",
			id:   uuid.New(),
			updateReq: &models.UpdateUserRequest{
				FirstName: stringPtr("Test"),
			},
			wantErr: true,
			errType: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := ts.storage.UpdateUser(ctx, tt.id, tt.updateReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errType != nil && err != tt.errType {
					t.Errorf("UpdateUser() error = %v, expected %v", err, tt.errType)
				}
				return
			}

			// Verify updated user
			if user == nil {
				t.Error("UpdateUser() returned nil user")
				return
			}

			if tt.checkField != nil && !tt.checkField(user) {
				t.Error("UpdateUser() field check failed")
			}

			// Check that updated_at was updated
			if !user.UpdatedAt.After(createdUser.UpdatedAt) {
				t.Error("UpdateUser() updated_at should be newer than created_at")
			}
		})
	}
}

func TestPostgresStorage_UpdateLastLogin(t *testing.T) {
	ts := setupTestDB(t)
	defer ts.teardownTestDB()

	ctx := context.Background()

	// Create a test user first
	req := createTestUser(999999994)
	createdUser, err := ts.storage.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
		errType error
	}{
		{
			name:    "existing user",
			id:      createdUser.ID,
			wantErr: false,
		},
		{
			name:    "non-existent user",
			id:      uuid.New(),
			wantErr: true,
			errType: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ts.storage.UpdateLastLogin(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateLastLogin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errType != nil && err != tt.errType {
					t.Errorf("UpdateLastLogin() error = %v, expected %v", err, tt.errType)
				}
				return
			}

			// Verify last login was updated
			user, err := ts.storage.GetUserByID(ctx, tt.id)
			if err != nil {
				t.Errorf("Failed to get user after UpdateLastLogin: %v", err)
				return
			}

			if user.LastLoginAt == nil {
				t.Error("UpdateLastLogin() last_login_at should not be nil")
			} else if user.LastLoginAt.Before(time.Now().Add(-5 * time.Second)) {
				t.Error("UpdateLastLogin() last_login_at should be recent")
			}
		})
	}
}

func TestPostgresStorage_DeactivateUser(t *testing.T) {
	ts := setupTestDB(t)
	defer ts.teardownTestDB()

	ctx := context.Background()

	// Create a test user first
	req := createTestUser(999999993)
	createdUser, err := ts.storage.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
		errType error
	}{
		{
			name:    "existing user",
			id:      createdUser.ID,
			wantErr: false,
		},
		{
			name:    "non-existent user",
			id:      uuid.New(),
			wantErr: true,
			errType: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ts.storage.DeactivateUser(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeactivateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errType != nil && err != tt.errType {
					t.Errorf("DeactivateUser() error = %v, expected %v", err, tt.errType)
				}
				return
			}

			// Verify user is deactivated (should not be found by GetUserByID which filters active users)
			_, err = ts.storage.GetUserByID(ctx, tt.id)
			if err != ErrUserNotFound {
				t.Error("DeactivateUser() user should not be found after deactivation")
			}
		})
	}
}

func TestPostgresStorage_ListActiveUsers(t *testing.T) {
	ts := setupTestDB(t)
	defer ts.teardownTestDB()

	ctx := context.Background()

	// Create multiple test users
	var createdUsers []*models.User
	for i := 0; i < 5; i++ {
		req := createTestUser(int64(999999990 - i))
		user, err := ts.storage.CreateUser(ctx, req)
		if err != nil {
			t.Fatalf("Failed to create test user %d: %v", i, err)
		}
		createdUsers = append(createdUsers, user)
	}

	tests := []struct {
		name       string
		limit      int
		offset     int
		wantCount  int
		wantErr    bool
	}{
		{
			name:      "list all users",
			limit:     10,
			offset:    0,
			wantCount: 5,
			wantErr:   false,
		},
		{
			name:      "limit to 3 users",
			limit:     3,
			offset:    0,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "offset by 2",
			limit:     10,
			offset:    2,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "limit 0",
			limit:     0,
			offset:    0,
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, err := ts.storage.ListActiveUsers(ctx, tt.limit, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListActiveUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(users) != tt.wantCount {
				t.Errorf("ListActiveUsers() count = %v, want %v", len(users), tt.wantCount)
			}

			// Verify users are sorted by created_at DESC
			for i := 1; i < len(users); i++ {
				if users[i-1].CreatedAt.Before(users[i].CreatedAt) {
					t.Error("ListActiveUsers() users should be sorted by created_at DESC")
					break
				}
			}
		})
	}
}

func TestPostgresStorage_GetUserCount(t *testing.T) {
	ts := setupTestDB(t)
	defer ts.teardownTestDB()

	ctx := context.Background()

	// Get initial count
	initialCount, err := ts.storage.GetUserCount(ctx)
	if err != nil {
		t.Fatalf("Failed to get initial user count: %v", err)
	}

	// Create test users
	numUsers := 3
	for i := 0; i < numUsers; i++ {
		req := createTestUser(int64(999999980 - i))
		_, err := ts.storage.CreateUser(ctx, req)
		if err != nil {
			t.Fatalf("Failed to create test user %d: %v", i, err)
		}
	}

	// Get count after creating users
	finalCount, err := ts.storage.GetUserCount(ctx)
	if err != nil {
		t.Fatalf("Failed to get final user count: %v", err)
	}

	expectedCount := initialCount + int64(numUsers)
	if finalCount != expectedCount {
		t.Errorf("GetUserCount() = %v, want %v", finalCount, expectedCount)
	}
}

func TestPostgresStorage_Health(t *testing.T) {
	ts := setupTestDB(t)
	defer ts.teardownTestDB()

	ctx := context.Background()

	err := ts.storage.Health(ctx)
	if err != nil {
		t.Errorf("Health() error = %v, want nil", err)
	}

	// Test health check with cancelled context
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()

	err = ts.storage.Health(cancelledCtx)
	if err == nil {
		t.Error("Health() with cancelled context should return error")
	}
}

// Helper functions for pointer types
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}