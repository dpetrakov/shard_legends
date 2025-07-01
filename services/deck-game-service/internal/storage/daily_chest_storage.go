package storage

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"github.com/shard-legends/deck-game-service/internal/service"
)

// DailyChestStorage implements daily chest data access using PostgreSQL
type DailyChestStorage struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewDailyChestStorage creates a new daily chest storage
func NewDailyChestStorage(pool *pgxpool.Pool, logger *slog.Logger) service.DailyChestRepository {
	return &DailyChestStorage{
		pool:   pool,
		logger: logger,
	}
}

// GetCompletedTasksCountToday returns the count of completed daily chest tasks for user today
func (s *DailyChestStorage) GetCompletedTasksCountToday(ctx context.Context, userID uuid.UUID, recipeID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM production.production_tasks
		WHERE user_id = $1
		  AND recipe_id = $2
		  AND status = 'claimed'
		  AND created_at::date = CURRENT_DATE
	`

	var count int
	err := s.pool.QueryRow(ctx, query, userID, recipeID).Scan(&count)
	if err != nil {
		s.logger.Error("Failed to get completed tasks count",
			"user_id", userID,
			"recipe_id", recipeID,
			"error", err)
		return 0, errors.Wrap(err, "failed to query completed tasks count")
	}

	s.logger.Debug("Completed tasks count retrieved",
		"user_id", userID,
		"recipe_id", recipeID,
		"count", count)

	return count, nil
}

// GetLastRewardTime returns the time of the last completed daily chest task for user
func (s *DailyChestStorage) GetLastRewardTime(ctx context.Context, userID uuid.UUID, recipeID uuid.UUID) (*time.Time, error) {
	query := `
		SELECT created_at
		FROM production.production_tasks
		WHERE user_id = $1
		  AND recipe_id = $2
		  AND status = 'claimed'
		ORDER BY created_at DESC
		LIMIT 1
	`

	var lastRewardAt time.Time
	err := s.pool.QueryRow(ctx, query, userID, recipeID).Scan(&lastRewardAt)
	if err != nil {
		// If no rows found, it means user never got a reward, return nil
		if err.Error() == "no rows in result set" {
			s.logger.Debug("No previous rewards found for user",
				"user_id", userID,
				"recipe_id", recipeID)
			return nil, nil
		}

		s.logger.Error("Failed to get last reward time",
			"user_id", userID,
			"recipe_id", recipeID,
			"error", err)
		return nil, errors.Wrap(err, "failed to query last reward time")
	}

	s.logger.Debug("Last reward time retrieved",
		"user_id", userID,
		"recipe_id", recipeID,
		"last_reward_at", lastRewardAt)

	return &lastRewardAt, nil
}
