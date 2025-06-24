import type { BlueprintType, ProcessedItemType, CraftedToolType } from '@/types/inventory';

export const TOOL_TYPES: BlueprintType[] = ['shovel', 'sickle', 'axe', 'pickaxe'];

export const TOOL_QUALITIES = [
    { id: 'wooden', name: 'Деревянный', material: 'wood_plank' as ProcessedItemType },
    { id: 'stone', name: 'Каменный', material: 'stone_block' as ProcessedItemType },
    { id: 'metal', name: 'Металлический', material: 'metal_ingot' as ProcessedItemType },
    { id: 'diamond', name: 'Бриллиантовый', material: 'cut_diamond' as ProcessedItemType },
];

export const getCraftedToolId = (tool: BlueprintType, qualityId: string): CraftedToolType => {
    return `${qualityId}_${tool}` as CraftedToolType;
};

export const CRAFTING_COST_PER_ITEM = 4;
export const CRAFTING_DURATION_SECONDS_PER_ITEM = 60;
