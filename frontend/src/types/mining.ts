
import type { CraftedToolType, LootResult } from './inventory';

export type ToolType = 'axe' | 'pickaxe' | 'shovel' | 'sickle';

export const ALL_TOOL_TYPES: ToolType[] = ['axe', 'pickaxe', 'shovel', 'sickle'];

export interface ToolStats {
    strength: number;
    speed: number;
    luck: number;
    durability: number;
}

export interface EquippedTool {
    item: CraftedToolType;
    stats: ToolStats;
}

export type EquippedTools = Partial<Record<ToolType, EquippedTool>>;

export type MiningLocation = 'cave' | 'forest';

export interface ActiveMiningProcess {
    location: MiningLocation;
    endTime: number; // JS timestamp
}

export interface MiningContextType {
    equippedTools: EquippedTools;
    equipTool: (toolType: ToolType, item: CraftedToolType) => boolean;
    unequipTool: (toolType: ToolType) => boolean;
    getToolIcon: (toolType: ToolType) => React.ComponentType<{ className?: string }>;
    getToolName: (toolType: ToolType) => string;
    activeProcess: ActiveMiningProcess | null;
    startMining: (location: MiningLocation) => boolean;
    claimMining: () => { location: MiningLocation; rewards: LootResult } | null;
}
