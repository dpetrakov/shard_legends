
import type { InventoryItemType } from './inventory';

export interface RecipeItem {
    item: InventoryItemType;
    quantity: number;
}

export interface RefiningRecipe {
    output: RecipeItem;
    input: RecipeItem[];
    durationSeconds: number; // Duration for a single batch
}

export interface ActiveProcess {
    id: string;
    recipe: RefiningRecipe;
    quantity: number; // Number of batches
    endTime: number; // JS timestamp (ms)
}

export interface RefiningContextType {
    activeProcesses: ActiveProcess[];
    startProcess: (recipe: RefiningRecipe, quantity: number) => boolean;
    claimProcess: (processId: string) => void;
    processSlots: {
        current: number;
        max: number;
    };
}
