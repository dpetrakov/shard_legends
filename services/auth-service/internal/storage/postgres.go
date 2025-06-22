package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shard-legends/auth-service/internal/models"
)

// PostgresStorage implements UserRepository interface using PostgreSQL
type PostgresStorage struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	// CreateUser creates a new user and returns the created user
	CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error)

	// GetUserByID retrieves a user by their internal UUID
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)

	// GetUserByTelegramID retrieves a user by their Telegram ID
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error)

	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, id uuid.UUID, req *models.UpdateUserRequest) (*models.User, error)

	// UpdateLastLogin updates the last login timestamp for a user
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error

	// DeactivateUser soft-deletes a user by setting is_active to false
	DeactivateUser(ctx context.Context, id uuid.UUID) error

	// ListActiveUsers returns a paginated list of active users
	ListActiveUsers(ctx context.Context, limit, offset int) ([]*models.User, error)

	// GetUserCount returns the total count of active users
	GetUserCount(ctx context.Context) (int64, error)

	// Close closes the database connection pool
	Close()

	// Health checks the database connection health
	Health(ctx context.Context) error
}

// ErrUserNotFound is returned when a user is not found
var ErrUserNotFound = errors.New("user not found")

// ErrUserAlreadyExists is returned when trying to create a user with existing Telegram ID
var ErrUserAlreadyExists = errors.New("user with this telegram_id already exists")

// NewPostgresStorage creates a new PostgreSQL storage instance
func NewPostgresStorage(databaseURL string, maxConns int, logger *slog.Logger) (*PostgresStorage, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool
	config.MaxConns = int32(maxConns)
	config.MinConns = 1
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("PostgreSQL connection established",
		slog.String("max_conns", fmt.Sprintf("%d", maxConns)),
	)

	return &PostgresStorage{
		pool:   pool,
		logger: logger,
	}, nil
}

// CreateUser creates a new user in the database
func (p *PostgresStorage) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	user := req.ToUser()

	query := `
		INSERT INTO auth.users (id, telegram_id, username, first_name, last_name, language_code, is_premium, photo_url, created_at, updated_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, telegram_id, username, first_name, last_name, language_code, is_premium, photo_url, created_at, updated_at, last_login_at, is_active`

	row := p.pool.QueryRow(ctx, query,
		user.ID, user.TelegramID, user.Username, user.FirstName, user.LastName,
		user.LanguageCode, user.IsPremium, user.PhotoURL, user.CreatedAt, user.UpdatedAt, user.IsActive,
	)

	var createdUser models.User
	err := row.Scan(
		&createdUser.ID, &createdUser.TelegramID, &createdUser.Username, &createdUser.FirstName,
		&createdUser.LastName, &createdUser.LanguageCode, &createdUser.IsPremium, &createdUser.PhotoURL,
		&createdUser.CreatedAt, &createdUser.UpdatedAt, &createdUser.LastLoginAt, &createdUser.IsActive,
	)

	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "users_telegram_id_key"` {
			p.logger.Warn("Attempted to create user with existing Telegram ID",
				slog.Int64("telegram_id", req.TelegramID),
			)
			return nil, ErrUserAlreadyExists
		}
		p.logger.Error("Failed to create user",
			slog.String("error", err.Error()),
			slog.Int64("telegram_id", req.TelegramID),
		)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	p.logger.Info("User created successfully",
		slog.String("user_id", createdUser.ID.String()),
		slog.Int64("telegram_id", createdUser.TelegramID),
	)

	return &createdUser, nil
}

// GetUserByID retrieves a user by their internal UUID
func (p *PostgresStorage) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name, language_code, is_premium, photo_url,
		       created_at, updated_at, last_login_at, is_active
		FROM auth.users
		WHERE id = $1 AND is_active = true`

	row := p.pool.QueryRow(ctx, query, id)

	var user models.User
	err := row.Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FirstName,
		&user.LastName, &user.LanguageCode, &user.IsPremium, &user.PhotoURL,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.IsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		p.logger.Error("Failed to get user by ID",
			slog.String("error", err.Error()),
			slog.String("user_id", id.String()),
		)
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetUserByTelegramID retrieves a user by their Telegram ID
func (p *PostgresStorage) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name, language_code, is_premium, photo_url,
		       created_at, updated_at, last_login_at, is_active
		FROM auth.users
		WHERE telegram_id = $1 AND is_active = true`

	row := p.pool.QueryRow(ctx, query, telegramID)

	var user models.User
	err := row.Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FirstName,
		&user.LastName, &user.LanguageCode, &user.IsPremium, &user.PhotoURL,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.IsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		p.logger.Error("Failed to get user by Telegram ID",
			slog.String("error", err.Error()),
			slog.Int64("telegram_id", telegramID),
		)
		return nil, fmt.Errorf("failed to get user by Telegram ID: %w", err)
	}

	return &user, nil
}

// UpdateUser updates an existing user
func (p *PostgresStorage) UpdateUser(ctx context.Context, id uuid.UUID, req *models.UpdateUserRequest) (*models.User, error) {
	// Build dynamic update query based on provided fields
	setParts := []string{"updated_at = NOW()"}
	args := []interface{}{id}
	argIndex := 2

	if req.Username != nil {
		setParts = append(setParts, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, req.Username)
		argIndex++
	}
	if req.FirstName != nil {
		setParts = append(setParts, fmt.Sprintf("first_name = $%d", argIndex))
		args = append(args, req.FirstName)
		argIndex++
	}
	if req.LastName != nil {
		setParts = append(setParts, fmt.Sprintf("last_name = $%d", argIndex))
		args = append(args, req.LastName)
		argIndex++
	}
	if req.LanguageCode != nil {
		setParts = append(setParts, fmt.Sprintf("language_code = $%d", argIndex))
		args = append(args, req.LanguageCode)
		argIndex++
	}
	if req.IsPremium != nil {
		setParts = append(setParts, fmt.Sprintf("is_premium = $%d", argIndex))
		args = append(args, req.IsPremium)
		argIndex++
	}
	if req.PhotoURL != nil {
		setParts = append(setParts, fmt.Sprintf("photo_url = $%d", argIndex))
		args = append(args, req.PhotoURL)
		argIndex++
	}
	if req.LastLoginAt != nil {
		setParts = append(setParts, fmt.Sprintf("last_login_at = $%d", argIndex))
		args = append(args, req.LastLoginAt)
		argIndex++
	}
	if req.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, req.IsActive)
		argIndex++
	}

	// Build the final query
	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf(`
		UPDATE auth.users
		SET %s
		WHERE id = $1 AND is_active = true
		RETURNING id, telegram_id, username, first_name, last_name, language_code, is_premium, photo_url,
		          created_at, updated_at, last_login_at, is_active`,
		setClause)

	row := p.pool.QueryRow(ctx, query, args...)

	var user models.User
	err := row.Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FirstName,
		&user.LastName, &user.LanguageCode, &user.IsPremium, &user.PhotoURL,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.IsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		p.logger.Error("Failed to update user",
			slog.String("error", err.Error()),
			slog.String("user_id", id.String()),
		)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	p.logger.Info("User updated successfully",
		slog.String("user_id", user.ID.String()),
	)

	return &user, nil
}

// UpdateLastLogin updates the last login timestamp for a user
func (p *PostgresStorage) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE auth.users
		SET last_login_at = NOW()
		WHERE id = $1 AND is_active = true`

	result, err := p.pool.Exec(ctx, query, id)
	if err != nil {
		p.logger.Error("Failed to update last login",
			slog.String("error", err.Error()),
			slog.String("user_id", id.String()),
		)
		return fmt.Errorf("failed to update last login: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeactivateUser soft-deletes a user by setting is_active to false
func (p *PostgresStorage) DeactivateUser(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE auth.users
		SET is_active = false, updated_at = NOW()
		WHERE id = $1`

	result, err := p.pool.Exec(ctx, query, id)
	if err != nil {
		p.logger.Error("Failed to deactivate user",
			slog.String("error", err.Error()),
			slog.String("user_id", id.String()),
		)
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	p.logger.Info("User deactivated successfully",
		slog.String("user_id", id.String()),
	)

	return nil
}

// ListActiveUsers returns a paginated list of active users
func (p *PostgresStorage) ListActiveUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name, language_code, is_premium, photo_url,
		       created_at, updated_at, last_login_at, is_active
		FROM auth.users
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := p.pool.Query(ctx, query, limit, offset)
	if err != nil {
		p.logger.Error("Failed to list active users",
			slog.String("error", err.Error()),
			slog.Int("limit", limit),
			slog.Int("offset", offset),
		)
		return nil, fmt.Errorf("failed to list active users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.TelegramID, &user.Username, &user.FirstName,
			&user.LastName, &user.LanguageCode, &user.IsPremium, &user.PhotoURL,
			&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.IsActive,
		)
		if err != nil {
			p.logger.Error("Failed to scan user row",
				slog.String("error", err.Error()),
			)
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		p.logger.Error("Error reading user rows",
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("error reading user rows: %w", err)
	}

	return users, nil
}

// GetUserCount returns the total count of active users
func (p *PostgresStorage) GetUserCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM auth.users WHERE is_active = true`

	var count int64
	err := p.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		p.logger.Error("Failed to get user count",
			slog.String("error", err.Error()),
		)
		return 0, fmt.Errorf("failed to get user count: %w", err)
	}

	return count, nil
}

// Health checks the database connection health
func (p *PostgresStorage) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := p.pool.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// Close closes the database connection pool
func (p *PostgresStorage) Close() {
	p.pool.Close()
	p.logger.Info("PostgreSQL connection pool closed")
}