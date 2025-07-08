
import type { InventoryItemType } from './inventory';

export interface RecipeIOItem {
    item_slug: InventoryItemType;
    quantity: number;
}

export interface ProductionRecipe {
    id: string; // Recipe UUID
    name: string;
    input_items: RecipeIOItem[];
    output_items: RecipeIOItem[];
    duration_seconds: number;
    // Client-side derived property
    category?: 'refining' | 'crafting';
}

export interface ProductionTask {
    id: string; // Task UUID
    recipe_id: string;
    quantity: number;
    status: 'in_progress' | 'completed' | 'queued';
    ends_at: string; // ISO 8601 timestamp string
}

export interface ProductionContextType {
    recipes: ProductionRecipe[];
    tasks: ProductionTask[];
    isLoading: boolean;
    fetchData: () => Promise<void>;
    startProduction: (recipeId: string, quantity: number) => Promise<boolean>;
    claimCompleted: () => Promise<void>;
    getRecipeById: (recipeId: string) => ProductionRecipe | undefined;
}
