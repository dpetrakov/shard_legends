package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// classifierService implements ClassifierService interface
type classifierService struct {
	deps *ServiceDependencies
}

// NewClassifierService creates a new classifier service
func NewClassifierService(deps *ServiceDependencies) ClassifierService {
	return &classifierService{
		deps: deps,
	}
}

// GetClassifierMapping returns a mapping of codes to UUIDs for a classifier
func (cs *classifierService) GetClassifierMapping(ctx context.Context, classifierCode string) (map[string]uuid.UUID, error) {
	if classifierCode == "" {
		return nil, errors.New("classifier code cannot be empty")
	}

	mapping, err := cs.deps.Repositories.Classifier.GetCodeToUUIDMapping(ctx, classifierCode)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get classifier mapping for code %s", classifierCode)
	}

	return mapping, nil
}

// GetReverseClassifierMapping returns a mapping of UUIDs to codes for a classifier
func (cs *classifierService) GetReverseClassifierMapping(ctx context.Context, classifierCode string) (map[uuid.UUID]string, error) {
	if classifierCode == "" {
		return nil, errors.New("classifier code cannot be empty")
	}

	mapping, err := cs.deps.Repositories.Classifier.GetUUIDToCodeMapping(ctx, classifierCode)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get reverse classifier mapping for code %s", classifierCode)
	}

	return mapping, nil
}

// RefreshClassifierCache invalidates the cache for a specific classifier
func (cs *classifierService) RefreshClassifierCache(ctx context.Context, classifierCode string) error {
	if classifierCode == "" {
		return errors.New("classifier code cannot be empty")
	}

	err := cs.deps.Repositories.Classifier.InvalidateCache(ctx, classifierCode)
	if err != nil {
		return errors.Wrapf(err, "failed to refresh classifier cache for code %s", classifierCode)
	}

	return nil
}
