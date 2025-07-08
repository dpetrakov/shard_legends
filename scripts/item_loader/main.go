package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Translation represents name/description for a language.
type Translation struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type Image struct {
	Collection   string `yaml:"collection"`
	QualityLevel string `yaml:"quality_level"`
	URL          string `yaml:"url"`
}

type Item struct {
	ID                      string                 `yaml:"id"`
	Code                    string                 `yaml:"code"`
	Class                   string                 `yaml:"class"`
	Type                    string                 `yaml:"type"`
	QualityLevelsClassifier string                 `yaml:"quality_levels_classifier"`
	CollectionsClassifier   string                 `yaml:"collections_classifier"`
	Translations            map[string]Translation `yaml:"translations"`
	Images                  []Image                `yaml:"images"`
}

type ItemsFile struct {
	Items []Item `yaml:"items"`
}

// NEW STRUCTS FOR CLASSIFIERS

type ClassifierItem struct {
	ID          string `yaml:"id"`
	Code        string `yaml:"code"`
	Description string `yaml:"description"`
}

type Classifier struct {
	ID          string           `yaml:"id"`
	Code        string           `yaml:"code"`
	Description string           `yaml:"description"`
	Items       []ClassifierItem `yaml:"items"`
}

type ClassifiersFile struct {
	Classifiers []Classifier `yaml:"classifiers"`
}

// === NEW TYPES FOR RECIPES ===

type RecipeInputItem struct {
	ItemID                string `yaml:"item_id"`
	ItemCode              string `yaml:"item_code"`
	QualityLevelCode      string `yaml:"quality_level_code"`
	FixedQualityLevelCode string `yaml:"fixed_quality_level_code"`
	Quantity              int    `yaml:"quantity"`
}

type RecipeOutputItem struct {
	ItemID                string  `yaml:"item_id"`
	ItemCode              string  `yaml:"item_code"`
	MinQuantity           int     `yaml:"min_quantity"`
	MaxQuantity           int     `yaml:"max_quantity"`
	ProbabilityPercent    float64 `yaml:"probability_percent"`
	OutputGroup           string  `yaml:"output_group"`
	FixedQualityLevelCode string  `yaml:"fixed_quality_level_code"`
}

type RecipeTranslation struct {
	LanguageCode string `yaml:"language_code"`
	FieldName    string `yaml:"field_name"`
	Content      string `yaml:"content"`
}

type RecipeLimit struct {
	LimitType string `yaml:"limit_type"`
	MaxUses   int    `yaml:"max_uses"`
}

type Recipe struct {
	ID                    string `yaml:"id"`
	Code                  string `yaml:"code"`
	OperationClassCode    string `yaml:"operation_class_code"`
	IsActive              bool   `yaml:"is_active"`
	ProductionTimeSeconds int    `yaml:"production_time_seconds"`

	InputItems   []RecipeInputItem   `yaml:"input_items"`
	OutputItems  []RecipeOutputItem  `yaml:"output_items"`
	Translations []RecipeTranslation `yaml:"translations"`
	Limits       []RecipeLimit       `yaml:"limits"`
}

type RecipesFile struct {
	Recipes []Recipe `yaml:"recipes"`
}

var (
	clsCache  = make(map[string]string)            // classifier_code -> id
	itemCache = make(map[string]map[string]string) // classifier_code -> (item_code -> id)
	cacheMu   sync.RWMutex
)

func getClassifierID(ctx context.Context, tx pgx.Tx, code string) (string, error) {
	cacheMu.RLock()
	if id, ok := clsCache[code]; ok {
		cacheMu.RUnlock()
		return id, nil
	}
	cacheMu.RUnlock()

	var id string
	err := tx.QueryRow(ctx, `SELECT id FROM inventory.classifiers WHERE code=$1`, code).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("classifier %s not found: %w", code, err)
	}
	cacheMu.Lock()
	clsCache[code] = id
	cacheMu.Unlock()
	return id, nil
}

func getClassifierItemID(ctx context.Context, tx pgx.Tx, classifierCode, itemCode string) (string, error) {
	cacheMu.RLock()
	if m, ok := itemCache[classifierCode]; ok {
		if id, ok2 := m[itemCode]; ok2 {
			cacheMu.RUnlock()
			return id, nil
		}
	}
	cacheMu.RUnlock()

	clsID, err := getClassifierID(ctx, tx, classifierCode)
	if err != nil {
		return "", err
	}
	var id string
	err = tx.QueryRow(ctx, `SELECT id FROM inventory.classifier_items WHERE classifier_id=$1 AND code=$2`, clsID, itemCode).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("classifier_item %s:%s not found: %w", classifierCode, itemCode, err)
	}
	cacheMu.Lock()
	if _, ok := itemCache[classifierCode]; !ok {
		itemCache[classifierCode] = make(map[string]string)
	}
	itemCache[classifierCode][itemCode] = id
	cacheMu.Unlock()
	return id, nil
}

// cache for item code -> id
var (
	itemCodeCache = make(map[string]string)
)

func getItemIDByCode(ctx context.Context, tx pgx.Tx, code string) (string, error) {
	cacheMu.RLock()
	if id, ok := itemCodeCache[code]; ok {
		cacheMu.RUnlock()
		return id, nil
	}
	cacheMu.RUnlock()

	var id string
	err := tx.QueryRow(ctx, `SELECT id FROM inventory.items WHERE code=$1`, code).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("item code %s not found: %w", code, err)
	}
	cacheMu.Lock()
	itemCodeCache[code] = id
	cacheMu.Unlock()
	return id, nil
}

// mapping item class -> type classifier code
var typeClassifierByClass = map[string]string{
	"resources":  "resource_type",
	"reagents":   "reagent_type",
	"boosters":   "booster_type",
	"tools":      "tool_type",
	"blueprints": "tool_type",
	"keys":       "key_type",
	"chests":     "chest_type",
	"currencies": "currency_type",
	"processed":  "processed_type",
}

func usage() {
	fmt.Println("Usage: item_loader [--all | --files file1.yaml,file2.yaml] [--dsn postgres://user:pass@host/db]")
	os.Exit(1)
}

type ImportStats struct {
	fileTotal       int
	itemFiles       int
	classifierFiles int
	recipeFiles     int

	itemUpserts       int
	classifierUpserts int
	elementUpserts    int
	recipeUpserts     int

	failed int
}

func main() {
	// Load env files first so that flags default correctly
	_ = godotenv.Overload(".env", "../../.env", "env.sample", "../../env.sample")

	var (
		allFlag   bool
		filesList string
		dirFlag   string
		dsn       string
	)

	flag.BoolVar(&allFlag, "all", false, "Load all yaml files from game-data/items directory")
	flag.StringVar(&filesList, "files", "", "Comma-separated list of yaml files to load")
	flag.StringVar(&dirFlag, "dir", "game-data", "Base directory to scan when --all is set")
	flag.StringVar(&dsn, "dsn", os.Getenv("DATABASE_URL"), "Postgres connection string")
	flag.Parse()

	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		log.Fatal("DSN is required (set --dsn flag or DATABASE_URL env var / .env file)")
	}

	var yamlFiles []string

	if allFlag {
		// Allow positional argument to override dirFlag
		if flag.NArg() > 0 {
			dirFlag = flag.Arg(0)
		}
		baseDir := dirFlag
		err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".yaml") {
				yamlFiles = append(yamlFiles, path)
			}
			return nil
		})
		if err != nil {
			log.Fatalf("walkDir error: %v", err)
		}
	} else if filesList != "" {
		yamlFiles = strings.Split(filesList, ",")
	} else {
		usage()
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	stats := &ImportStats{}

	for _, fp := range yamlFiles {
		// Determine file type by YAML root key
		kind, err := detectFileKind(fp)
		if err != nil {
			log.Printf("❌ %s parse error: %v", fp, err)
			stats.failed++
			continue
		}
		if kind == "classifier" {
			stats.classifierFiles++
			if err := processClassifierFile(ctx, pool, fp, stats); err != nil {
				log.Printf("❌ classifier file %s error: %v", fp, err)
				stats.failed++
			}
			continue
		}
		if kind == "recipe" {
			stats.recipeFiles++
			if err := processRecipeFile(ctx, pool, fp, stats); err != nil {
				log.Printf("❌ recipe file %s error: %v", fp, err)
				stats.failed++
			}
			continue
		}

		stats.itemFiles++
		if err := processItemFile(ctx, pool, fp, stats); err != nil {
			log.Printf("❌ item file %s error: %v", fp, err)
		}
	}

	// summary
	fmt.Println("==== Import summary ====")
	fmt.Printf("Classifier files: %d\n", stats.classifierFiles)
	fmt.Printf("Classifiers upserted: %d\n", stats.classifierUpserts)
	fmt.Printf("Elements upserted: %d\n", stats.elementUpserts)
	fmt.Printf("Item files: %d\n", stats.itemFiles)
	fmt.Printf("Items upserted: %d\n", stats.itemUpserts)
	fmt.Printf("Recipe files: %d\n", stats.recipeFiles)
	fmt.Printf("Recipes upserted: %d\n", stats.recipeUpserts)
	fmt.Printf("Failed : %d\n", stats.failed)
}

// processItemFile handles items yaml
func processItemFile(ctx context.Context, pool *pgxpool.Pool, path string, stats *ImportStats) error {

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	var f ItemsFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return fmt.Errorf("yaml parse: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("tx begin: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, item := range f.Items {
		if err := upsertItem(ctx, tx, item); err != nil {
			log.Printf("item %s error: %v", item.Code, err)
		} else {
			stats.itemUpserts++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx commit: %w", err)
	}
	return nil
}

// upsertItem performs deletion of translations/images, then upserts item and re-inserts related data.
func upsertItem(ctx context.Context, tx pgx.Tx, it Item) error {
	// Resolve classifier & item IDs
	itemClassID, err := getClassifierItemID(ctx, tx, "item_class", it.Class)
	if err != nil {
		return err
	}
	typeClassifier, ok := typeClassifierByClass[it.Class]
	if !ok {
		return fmt.Errorf("unknown item class %s", it.Class)
	}
	itemTypeID, err := getClassifierItemID(ctx, tx, typeClassifier, it.Type)
	if err != nil {
		return err
	}
	qlClassifierCode := it.QualityLevelsClassifier
	if qlClassifierCode == "" {
		qlClassifierCode = "quality_level"
	}
	qlClassifierID, err := getClassifierID(ctx, tx, qlClassifierCode)
	if err != nil {
		return err
	}
	colClassifierCode := it.CollectionsClassifier
	if colClassifierCode == "" {
		colClassifierCode = "collection"
	}
	colClassifierID, err := getClassifierID(ctx, tx, colClassifierCode)
	if err != nil {
		return err
	}

	// Remove existing translations & images
	if _, err := tx.Exec(ctx, `DELETE FROM i18n.translations WHERE entity_type='item' AND entity_id=$1`, it.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM inventory.item_images WHERE item_id=$1`, it.ID); err != nil {
		return err
	}

	// Upsert core item
	if _, err := tx.Exec(ctx, `INSERT INTO inventory.items (id, item_class_id, item_type_id, quality_levels_classifier_id, collections_classifier_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,now(),now())
		ON CONFLICT (id) DO UPDATE SET item_class_id=EXCLUDED.item_class_id, item_type_id=EXCLUDED.item_type_id,
		  quality_levels_classifier_id=EXCLUDED.quality_levels_classifier_id, collections_classifier_id=EXCLUDED.collections_classifier_id, updated_at=now()`,
		it.ID, itemClassID, itemTypeID, qlClassifierID, colClassifierID); err != nil {
		return err
	}

	// Insert translations
	for lang, t := range it.Translations {
		if err := upsertTranslation(ctx, tx, it.ID, "name", lang, t.Name); err != nil {
			return err
		}
		if err := upsertTranslation(ctx, tx, it.ID, "description", lang, t.Description); err != nil {
			return err
		}
	}

	// Insert images
	for _, img := range it.Images {
		colID, err := getClassifierItemID(ctx, tx, colClassifierCode, img.Collection)
		if err != nil {
			return err
		}
		qlvlID, err := getClassifierItemID(ctx, tx, qlClassifierCode, img.QualityLevel)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO inventory.item_images (item_id, collection_id, quality_level_id, image_url, created_at, updated_at)
			VALUES ($1,$2,$3,$4,now(),now())
			ON CONFLICT (item_id,collection_id,quality_level_id) DO UPDATE SET image_url=EXCLUDED.image_url, updated_at=now()`,
			it.ID, colID, qlvlID, img.URL); err != nil {
			return err
		}
	}
	// Cache code->id for later recipe resolution
	cacheMu.Lock()
	itemCodeCache[it.Code] = it.ID
	cacheMu.Unlock()
	return nil
}

func upsertTranslation(ctx context.Context, tx pgx.Tx, entityID, field, lang, content string) error {
	_, err := tx.Exec(ctx, `INSERT INTO i18n.translations (entity_type, entity_id, field_name, language_code, content, created_at, updated_at)
		VALUES ('item',$1,$2,$3,$4,now(),now())
		ON CONFLICT (entity_type,entity_id,field_name,language_code) DO UPDATE
		SET content=EXCLUDED.content, updated_at=now()`,
		entityID, field, lang, content)
	return err
}

// === Classifier processing ===
func processClassifierFile(ctx context.Context, pool *pgxpool.Pool, path string, stats *ImportStats) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	var cf ClassifiersFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return fmt.Errorf("yaml parse: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("tx begin: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, cls := range cf.Classifiers {
		// Upsert classifier (by code). Если ID не задан — полагаемся на DEFAULT.
		var dbClsID string
		if cls.ID != "" {
			if err := tx.QueryRow(ctx, `INSERT INTO inventory.classifiers (id, code, description, created_at, updated_at)
				VALUES ($1,$2,$3,now(),now())
				ON CONFLICT (code) DO UPDATE SET description=EXCLUDED.description, updated_at=now()
				RETURNING id`, cls.ID, cls.Code, cls.Description).Scan(&dbClsID); err != nil {
				return fmt.Errorf("classifier %s insert: %w", cls.Code, err)
			}
		} else {
			if err := tx.QueryRow(ctx, `INSERT INTO inventory.classifiers (code, description, created_at, updated_at)
				VALUES ($1,$2,now(),now())
				ON CONFLICT (code) DO UPDATE SET description=EXCLUDED.description, updated_at=now()
				RETURNING id`, cls.Code, cls.Description).Scan(&dbClsID); err != nil {
				return fmt.Errorf("classifier %s insert: %w", cls.Code, err)
			}
		}

		stats.classifierUpserts++

		for _, it := range cls.Items {
			if it.Code == "" {
				continue // skip blank codes
			}
			if _, err := tx.Exec(ctx, `INSERT INTO inventory.classifier_items (classifier_id, code, description, created_at, updated_at)
				VALUES ($1,$2,$3,now(),now())
				ON CONFLICT (classifier_id, code) DO UPDATE SET description=EXCLUDED.description, updated_at=now()`,
				dbClsID, it.Code, it.Description); err != nil {
				log.Printf("⚠️  item %s/%s upsert warning: %v", cls.Code, it.Code, err)
			}
			stats.elementUpserts++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx commit: %w", err)
	}
	return nil
}

// detectFileKind returns "classifier" or "item" based on top-level key
func detectFileKind(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var m map[string]interface{}
	if err := yaml.Unmarshal(data, &m); err != nil {
		return "", err
	}
	if _, ok := m["classifiers"]; ok {
		return "classifier", nil
	}
	if _, ok := m["items"]; ok {
		return "item", nil
	}
	if _, ok := m["recipes"]; ok {
		return "recipe", nil
	}
	return "", fmt.Errorf("unknown yaml structure")
}

func processRecipeFile(ctx context.Context, pool *pgxpool.Pool, path string, stats *ImportStats) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	var rf RecipesFile
	if err := yaml.Unmarshal(data, &rf); err != nil {
		return fmt.Errorf("yaml parse: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("tx begin: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, rec := range rf.Recipes {
		if err := upsertRecipe(ctx, tx, rec); err != nil {
			log.Printf("recipe %s error: %v", rec.Code, err)
		} else {
			stats.recipeUpserts++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx commit: %w", err)
	}
	return nil
}

func upsertRecipe(ctx context.Context, tx pgx.Tx, r Recipe) error {
	// Cleanup dependent tables (translations, limits, io items)
	if _, err := tx.Exec(ctx, `DELETE FROM i18n.translations WHERE entity_type='recipe' AND entity_id=$1`, r.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM production.recipe_limits WHERE recipe_id=$1`, r.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM production.recipe_output_items WHERE recipe_id=$1`, r.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM production.recipe_input_items WHERE recipe_id=$1`, r.ID); err != nil {
		return err
	}

	// Upsert core recipe record
	if _, err := tx.Exec(ctx, `INSERT INTO production.recipes (id, code, operation_class_code, is_active, production_time_seconds)
        VALUES ($1,$2,$3,$4,$5)
        ON CONFLICT (id) DO UPDATE SET
            code=EXCLUDED.code,
            operation_class_code=EXCLUDED.operation_class_code,
            is_active=EXCLUDED.is_active,
            production_time_seconds=EXCLUDED.production_time_seconds`,
		r.ID, r.Code, r.OperationClassCode, r.IsActive, r.ProductionTimeSeconds); err != nil {
		return err
	}

	// Insert input items
	for _, inItem := range r.InputItems {
		// Resolve item ID by code if necessary
		itemID := inItem.ItemID
		if itemID == "" && inItem.ItemCode != "" {
			var err error
			itemID, err = getItemIDByCode(ctx, tx, inItem.ItemCode)
			if err != nil {
				return err
			}
		}
		if itemID == "" {
			return fmt.Errorf("input item missing id/code in recipe %s", r.Code)
		}

		qlvl := inItem.QualityLevelCode
		if qlvl == "" {
			qlvl = inItem.FixedQualityLevelCode
		}

		if _, err := tx.Exec(ctx, `INSERT INTO production.recipe_input_items (recipe_id, item_id, quality_level_code, quantity)
            VALUES ($1,$2,$3,$4)`,
			r.ID, itemID, qlvl, inItem.Quantity); err != nil {
			return err
		}
	}

	// Insert output items
	for _, out := range r.OutputItems {
		var fq interface{}
		if out.FixedQualityLevelCode == "" {
			fq = nil
		} else {
			fq = out.FixedQualityLevelCode
		}
		itemID := out.ItemID
		if itemID == "" && out.ItemCode != "" {
			var err error
			itemID, err = getItemIDByCode(ctx, tx, out.ItemCode)
			if err != nil {
				return err
			}
		}
		if itemID == "" {
			return fmt.Errorf("output item missing id/code in recipe %s", r.Code)
		}

		if _, err := tx.Exec(ctx, `INSERT INTO production.recipe_output_items (recipe_id, item_id, min_quantity, max_quantity, probability_percent, output_group, fixed_quality_level_code)
            VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			r.ID, itemID, out.MinQuantity, out.MaxQuantity, out.ProbabilityPercent, out.OutputGroup, fq); err != nil {
			return err
		}
	}

	// Insert limits
	for _, lim := range r.Limits {
		if _, err := tx.Exec(ctx, `INSERT INTO production.recipe_limits (recipe_id, limit_type, max_uses)
            VALUES ($1,$2,$3)`,
			r.ID, lim.LimitType, lim.MaxUses); err != nil {
			return err
		}
	}

	// Insert translations
	for _, tr := range r.Translations {
		if _, err := tx.Exec(ctx, `INSERT INTO i18n.translations (entity_type, entity_id, field_name, language_code, content, created_at, updated_at)
            VALUES ('recipe',$1,$2,$3,$4,now(),now())
            ON CONFLICT (entity_type,entity_id,field_name,language_code) DO UPDATE SET content=EXCLUDED.content, updated_at=now()`,
			r.ID, tr.FieldName, tr.LanguageCode, tr.Content); err != nil {
			return err
		}
	}

	return nil
}
