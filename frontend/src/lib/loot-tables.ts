
import type { ChestType } from '@/types/profile';
import type { LootResult, ResourceType, ReagentType, BlueprintType } from '@/types/inventory';

// This file's contents are deprecated as chest opening logic is now handled by the server.
// The file is kept to prevent breaking imports, but its functions are no longer used for chest opening.

export function openChest(chestType: ChestType): LootResult {
    console.warn("openChest is deprecated and should not be used. Chest opening is handled by the server.");
    return {};
}
