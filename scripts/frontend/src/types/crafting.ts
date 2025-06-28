
import type { InventoryItemType, BlueprintType } from './inventory';

export interface CraftingRecipeInfo {
    tool: BlueprintType;
    quality: string; // e.g. 'wooden', 'stone'
}

export interface ActiveCraftingProcess {
    id: string;
    recipe: CraftingRecipeInfo;
    quantity: number;
    endTime: number;
    outputItem: InventoryItemType;
    outputItemName: string;
}

export interface CraftingContextType {
    activeProcesses: ActiveCraftingProcess[];
    startProcess: (recipeInfo: CraftingRecipeInfo, quantity: number) => boolean;
    claimProcess: (processId: string) => void;
    processSlots: {
        current: number;
        max: number;
    };
}
