
import type { ChestType } from '@/types/profile';
import type { LootResult, ResourceType, ReagentType, BlueprintType } from '@/types/inventory';

const resourceLootTable: { item: ResourceType; chance: number }[] = [
  { item: 'stone', chance: 0.40 },
  { item: 'wood', chance: 0.40 },
  { item: 'ore', chance: 0.15 },
  { item: 'diamond', chance: 0.05 },
];

const reagentLootTable: { item: ReagentType; chance: number }[] = [
  { item: 'abrasive', chance: 0.40 },
  { item: 'disc', chance: 0.40 },
  { item: 'inductor', chance: 0.15 },
  { item: 'paste', chance: 0.05 },
];

const blueprintLootTable: { item: BlueprintType; chance: number }[] = [
    { item: 'shovel', chance: 0.25 },
    { item: 'sickle', chance: 0.25 },
    { item: 'axe', chance: 0.25 },
    { item: 'pickaxe', chance: 0.25 },
];

// ChestType has unions like 'resource_small' | 'reagent_medium' etc.
// We extract only the ones that have capacities defined here.
type CapacityChestType = Extract<ChestType, `${'resource' | 'reagent'}_${string}`>;

const chestBaseCapacity: Record<CapacityChestType, number> = {
    'resource_small': 100,
    'resource_medium': 3500,
    'resource_large': 47000,
    'reagent_small': 10, // 10x less
    'reagent_medium': 350,
    'reagent_large': 4700,
};

function rollOnTable<T extends string>(table: { item: T; chance: number }[]): T {
    let rand = Math.random();
    let cumulativeChance = 0;
    for (const entry of table) {
        cumulativeChance += entry.chance;
        if (rand < cumulativeChance) {
            return entry.item;
        }
    }
    // Fallback to the last item in case of floating point inaccuracies
    return table[table.length - 1].item;
}

export function openChest(chestType: ChestType): LootResult {
    const loot: LootResult = {};
    const [category, size] = chestType.split('_');

    if (category === 'resource' && size) {
        const key = chestType as CapacityChestType;
        const capacity = chestBaseCapacity[key];
        const rolledItem = rollOnTable(resourceLootTable);
        const itemChance = resourceLootTable.find(i => i.item === rolledItem)!.chance;
        const quantity = Math.round(capacity * itemChance);
        if (quantity > 0) {
            loot[rolledItem] = quantity;
        }
    } else if (category === 'reagent' && size) {
        const key = chestType as CapacityChestType;
        const capacity = chestBaseCapacity[key];
        const rolledItem = rollOnTable(reagentLootTable);
        const itemChance = reagentLootTable.find(i => i.item === rolledItem)!.chance;
        // With guarantee of at least 1, and flooring as implied by user's example
        const quantity = Math.max(1, Math.floor(capacity * itemChance));
        loot[rolledItem] = quantity;
    } else if (category === 'booster' && size) {
        // Booster logic to be implemented in the future
        console.log(`Opening booster chest: ${chestType}. Logic not defined yet.`);
    } else if (category === 'blueprint') {
        const rolledItem = rollOnTable(blueprintLootTable);
        loot[rolledItem] = 1;
    }

    return loot;
}
