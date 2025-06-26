package storage

// NewRepository создает новый экземпляр Repository со всеми репозиториями
func NewRepository(deps *RepositoryDependencies) *Repository {
	return &Repository{
		Recipe:     NewRecipeRepository(deps),
		Task:       NewTaskRepository(deps),
		Classifier: NewClassifierRepository(deps),
	}
}

// Constants для классификаторов (аналогично inventory-service)
const (
	ClassifierProductionOperationClasses = "production_operation_classes"
	ClassifierRecipeLimitTypes           = "recipe_limit_types"
	ClassifierRecipeLimitObjects         = "recipe_limit_objects"
	ClassifierItemClass                  = "item_class"
	ClassifierQualityLevel               = "quality_level"
	ClassifierCollection                 = "collection"
)
