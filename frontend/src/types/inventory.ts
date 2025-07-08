
import type { ChestType } from './profile';

export type ResourceType = 'stone' | 'wood' | 'ore' | 'diamond';
export type ReagentType = 'abrasive' | 'disc' | 'inductor' | 'paste';
export type BlueprintType = 'shovel' | 'sickle' | 'axe' | 'pickaxe';
export type BoosterType = 'speed_1h' | 'speed_3h' | 'speed_12h'; // Placeholder
export type ProcessedItemType = 'wood_plank' | 'stone_block' | 'metal_ingot' | 'cut_diamond';

// Dynamically create crafted tool types
const qualities = ['wooden', 'stone', 'metal', 'diamond'] as const;
const tools = ['shovel', 'sickle', 'axe', 'pickaxe'] as const;
type Quality = typeof qualities[number];
type Tool = typeof tools[number];
export type CraftedToolType = `${Quality}_${Tool}`;

export type InventoryItemType = ResourceType | ReagentType | BlueprintType | BoosterType | ProcessedItemType | CraftedToolType;

// NEW: Details for an item, fetched from the server.
export interface ItemDetails {
    id: string; // This is the UUID from the server
    slug: InventoryItemType | ChestType;
    name: string;
    description?: string;
    imageUrl?: string;
}

export const AllResourceTypes: ResourceType[] = ['stone', 'wood', 'ore', 'diamond'];
export const AllReagentTypes: ReagentType[] = ['abrasive', 'disc', 'inductor', 'paste'];
export const AllBlueprintTypes: BlueprintType[] = ['shovel', 'sickle', 'axe', 'pickaxe'];
export const AllProcessedItemTypes: ProcessedItemType[] = ['wood_plank', 'stone_block', 'metal_ingot', 'cut_diamond'];
export const AllCraftedToolTypes: CraftedToolType[] = qualities.flatMap(q => tools.map(t => `${q}_${t}` as CraftedToolType));


export type LootResult = { [key in InventoryItemType | ChestType]?: number };

export interface InventoryContextType {
  inventory: LootResult;
  itemDetails: Record<string, ItemDetails>;
  addItems: (itemsToAdd: LootResult) => void;
  spendItems: (itemsToSpend: LootResult) => boolean;
  getItemName: (item: InventoryItemType | ChestType) => string;
  getItemImage: (item: InventoryItemType | ChestType) => string | undefined;
  syncWithServer: () => Promise<void>;
}
